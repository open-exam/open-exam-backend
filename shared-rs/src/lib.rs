pub mod shared {

    use std::sync::Arc;

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
            use byteorder::{WriteBytesExt, BigEndian};
            buf.write_u32::<BigEndian>(data.len() as u32);
            buf.extend_from_slice(&data);
        }
        {
            use tokio::io::AsyncWriteExt;
            stream.lock().await.write(&buf).await;
        }
    }
}