use std::{net::SocketAddr, sync::Arc, thread};
use net2::{TcpBuilder, unix::UnixTcpBuilderExt};
use rand::Rng;
use tokio::{net::{TcpListener, TcpStream}, sync::Mutex};
use xactor::*;
use x448::{Secret, PublicKey};

mod client_handler;

#[tokio::main]
async fn main() {
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

    for i in 0..num_cpus::get() {
        threads.push(thread::spawn(move || {
            println!("Starting worker thread {}", i);

            let rt = tokio::runtime::Builder::new_multi_thread()
            .enable_all()
            .build().unwrap();

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
                        tokio::spawn(async move {
                            process(socket, i, addr).await;
                        });
                    }
                }
            });
        }));
    }

    println!("Listening on {}", addr);
    for thread in threads {
        thread.join().unwrap();
    }
}

async fn process(mut socket: TcpStream, workerId: usize, addr: SocketAddr) {
    let mut socket = Arc::new(Mutex::new(socket));
    let ctx = client_handler::Client {
        id: workerId,
        stream: socket.clone(),
        key: None
    };
    let addr = ctx.start().await.unwrap();
    
    if let Ok(init) = addr.call(client_handler::InitClient{}).await {
        if init.0 != 200 {
            let mut free_socket= socket.lock().await;
            if free_socket.writable().await.is_ok() {
                let mut buf = Vec::new();
                {
                    use byteorder::{WriteBytesExt, BigEndian};
                    use std::io::Write;
                    buf.write_u32::<BigEndian>(8 + init.1.len() as u32);
                    buf.write_u32::<BigEndian>(init.0);
                    buf.write(init.1.as_bytes());
                }
                {
                    use tokio::io::AsyncWriteExt;
                    free_socket.write(&buf).await;
                }
            }
            {
                use tokio::io::AsyncWriteExt;
                free_socket.shutdown().await;
            }
            return;
        }

        
    }
}