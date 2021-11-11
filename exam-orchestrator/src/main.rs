use std::{net::SocketAddr, sync::Arc, thread};
use net2::{TcpBuilder, unix::UnixTcpBuilderExt};
use redis::{Commands, RedisResult, Value, cluster::ClusterClient, streams::{StreamReadOptions, StreamReadReply}};
use serde::{Deserialize, Serialize};
use tokio::{net::{TcpListener, TcpStream}, sync::{Mutex, broadcast::{self, Receiver}}};
use xactor::*;

mod client_handler;

#[derive(Debug, Serialize, Deserialize)]
struct Claims {
    exp: usize,
    iat: usize,
    iss: String,
    nbf: usize,
    sub: String,
    user: String
}

#[derive(Clone)]
struct Notification {
    user_id: String,
    data: String
}

fn main() {
    let mut found = false;
    for (i, arg) in std::env::args().enumerate() {
        if arg.trim() == "--help" {
            println!("Usage: exam-orchestrator\n\t[-dev] run in dev mode\n\t[-env] the .env file to be used");
            std::process::exit(0);
        }

        if arg.trim() == "-dev" {
            std::env::set_var("RUST_LOG", "debug");
        }

        if arg.trim() == "-env" {
            let args: Vec<String> = std::env::args().collect();

            if i + 1 >= args.len() {
                println!("-env path not provided");
                std::process::exit(1);
            }
            else {
                dotenv::from_path(args[i + 1].clone()).ok();
                found = true;
            }
        }
    }
    
    if !found {
        dotenv::dotenv().ok();
    }

    let addr: std::net::SocketAddr = std::env::var("listen_addr").ok().unwrap().parse().unwrap();
    let mut threads = Vec::new();
    let (tx, rx) = broadcast::channel::<Notification>(16);

    for i in 0..num_cpus::get() {
        let tx = tx.clone();
        threads.push(thread::spawn(move || {
            println!("Starting worker thread {}", i);

            let rt = tokio::runtime::Builder::new_multi_thread().enable_all().build().unwrap();

            rt.block_on(async move {
                let listener = {
                    let builder = TcpBuilder::new_v4().unwrap();
                    builder.reuse_address(true).unwrap();
					builder.reuse_port(true).unwrap();
					builder.bind(addr).unwrap();
					builder.listen(2048).unwrap()
                };

                let listener = TcpListener::from_std(listener).unwrap();

                loop {
                    if let Ok((socket, addr)) = listener.accept().await {
                        tokio::spawn(process(socket, i, addr,  tx.subscribe()));
                    }
                }
            });
        }));
    }

    threads.push(thread::spawn(move || {
        println!("Starting redis thread");

        let rt = tokio::runtime::Builder::new_multi_thread().enable_all().build().unwrap();

        rt.block_on(async move {
            let nodes_var = std::env::var("redis_addr").ok().unwrap();
            let node_name = std::env::var("node_name").unwrap();
            let redis_password = std::env::var("redis_password").unwrap();
            let nodes: Vec<String> = nodes_var.split(",").map(|e| format!("redis://:{}@{}", redis_password, e.trim())).collect();

            let mut conn = ClusterClient::open(nodes).unwrap().get_connection().unwrap();
            let opts = StreamReadOptions::default().block(1000);

            let mut last_id = [String::from("0")];
            loop {
                let res: RedisResult<StreamReadReply> = conn.xread_options(&[node_name.as_str()], &last_id, &opts);
                if let Ok(items) = res {
                    items.keys.iter().for_each(|item| {
                        item.ids.iter().for_each(|ele| {
                            let mut user_id = "".to_string();
                            let mut notif_data = "".to_string();
                            ele.map.iter().for_each(|(key, value)| {
                                match value {
                                    Value::Data(data) => {
                                        println!("{} : {} : {}", ele.id, key, String::from_utf8(data.to_vec()).unwrap());
                                        if key == "user_id" {
                                            user_id = String::from_utf8(data.to_vec()).unwrap();
                                        }
                                        if key == "data" {
                                            notif_data = String::from_utf8(data.to_vec()).unwrap();
                                        }
                                    },
                                    _ => {}
                                }
                            });

                            if user_id.len() > 0 {
                                tx.send(Notification {
                                    user_id,
                                    data: notif_data,
                                });
                            }
                        })
                    });

                    if items.keys.len() > 0 {
                        let len_item_keys = items.keys.len();
                        let len_ele_ids = items.keys[len_item_keys - 1].ids.len();
                        last_id[0] = String::from(items.keys[len_item_keys - 1].ids[len_ele_ids - 1].id.clone());
                    }
                }
            }

        });
    }));

    println!("Listening on {}", addr);
    for thread in threads {
        thread.join().unwrap();
    }
}

async fn process(socket: TcpStream, worker_id: usize, addr: SocketAddr, mut rx: Receiver<Notification>) {
    let socket = Arc::new(Mutex::new(socket));
    let ctx = client_handler::Client {
        id: worker_id,
        stream: socket.clone(),
        key: None
    };
    let addr = ctx.start().await.unwrap();
    let user_id: String;
    
    if let Ok(init) = addr.call(client_handler::InitClient{}).await {
        if init.0 != 200 {
            send_status(socket.clone(), init.0, init.1.as_bytes());
            return;
        }

        {
            use jsonwebtoken::{decode, DecodingKey, Validation, Algorithm};
            let token = decode::<Claims>(&init.1, &DecodingKey::from_secret("secret".as_ref()), &Validation::new(Algorithm::RS256));
            if let Ok(token_data) = token {
                user_id = token_data.claims.user;
            }
            else {
                send_status(socket.clone(), 400, b"Invalid login token");
                return
            }
        }
    }
    else {
        send_status(socket.clone(), 500, "An unknown error occurred".as_bytes());
        return;
    }

    loop {
        use tokio::io::AsyncReadExt;
        let mut locked_socket = socket.lock().await;
        tokio::select! {
            result = locked_socket.read_u32() => {
                let n = result.unwrap_or_else(|_| 0);
                addr.call(client_handler::Request(n)).await;
            }
            result = rx.recv() => {
                std::mem::drop(locked_socket);
                let msg = result.unwrap();

                if msg.user_id == user_id {
                    addr.call(client_handler::Notification(msg.data)).await;
                }
            }
        }
    }
}

async fn send_status(stream: Arc<Mutex<TcpStream>>, status: u32, data: &[u8]) {
    {
        use byteorder::{WriteBytesExt, BigEndian};

        let mut buf = Vec::new();
        buf.write_u32::<BigEndian>(status);
        buf.extend(data);
        shared_rs::shared::write(&stream, &buf).await;
    }

    {
        use tokio::io::AsyncWriteExt;
        stream.lock().await.shutdown().await;
    }
}