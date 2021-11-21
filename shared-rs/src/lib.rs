pub mod shared {

    use std::{io::Read, sync::Arc};

    use rand::Rng;
    use tokio::{net::TcpStream, sync::Mutex};

    pub fn get_random_buf() -> String {
        let mut rng = rand::thread_rng();
        let mut buf = String::new();
        for _ in 0..32 {
            buf.push(rng.gen_range(b'a'..b'z') as char);
        }
        buf
    }

    pub fn random_bytes(len: usize) -> Vec<u8> {
        let mut rng = rand::thread_rng();
        let mut buf = Vec::with_capacity(len);
        for _ in 0..len {
            buf.push(rng.gen_range(0..256) as u8);
        }
        buf
    }

    pub async fn write(stream: &Arc<Mutex<TcpStream>>, data: &[u8]) {
        let mut buf = Vec::new();
        {
            use byteorder::{BigEndian, WriteBytesExt};
            buf.write_u32::<BigEndian>(data.len() as u32);
            buf.extend_from_slice(&data);
        }
        {
            use tokio::io::AsyncWriteExt;
            stream.lock().await.write(&buf).await;
        }
    }

    pub async fn standard_write(stream: &mut TcpStream, data: &[u8]) {
        let mut buf = Vec::new();
        {
            use byteorder::{BigEndian, WriteBytesExt};
            buf.write_u32::<BigEndian>(data.len() as u32);
            buf.extend_from_slice(&data);
        }
        {
            use tokio::io::AsyncWriteExt;
            stream.write(&buf).await;
        }
    }

    pub fn mode() -> &'static str {
        let mut mode = "prod";
        let mut found = false;
        for (i, arg) in std::env::args().enumerate() {
            if arg.trim() == "--help" {
                println!("Usage: exam-orchestrator\n\t[-dev] run in dev mode\n\t[-env] the .env file to be used");
                std::process::exit(0);
            }

            if arg.trim() == "-dev" {
                std::env::set_var("RUST_LOG", "debug");
                mode = "dev";
            }

            if arg.trim() == "-env" {
                let args: Vec<String> = std::env::args().collect();

                if i + 1 >= args.len() {
                    println!("-env path not provided");
                    std::process::exit(1);
                } else {
                    dotenv::from_path(args[i + 1].clone()).ok();
                    found = true;
                }
            }
        }

        if !found {
            dotenv::dotenv().ok();
        }
        mode
    }

    pub fn read_u32(data: &[u8]) -> u32 {
        use byteorder::{BigEndian, ReadBytesExt};
        let mut reader = std::io::Cursor::new(data);
        reader.read_u32::<BigEndian>().unwrap()
    }

    pub fn read_bytes_as_string(data: &[u8], len: usize) -> Result<String, std::io::Error> {
        let mut reader = std::io::Cursor::new(data);
        let mut buf = Vec::new();
        buf.reserve_exact(len);

        let res = reader.read_exact(&mut buf);
        if res.is_err() {
            return Err(res.err().unwrap());
        }

        Ok(String::from_utf8(buf.to_vec()).unwrap())
    }
}
