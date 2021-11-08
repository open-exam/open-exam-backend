use std::{sync::Arc};
use tokio::{io::AsyncReadExt, net::TcpStream, sync::Mutex};
use x448::{PublicKey, Secret};
use xactor::*;

const MAX_PACKET_SIZE: usize = 1048576;
const INIT_MAX_PACKET_SIZE: usize = 512;

pub struct Client {
    pub id: usize,
    pub stream: Arc<Mutex<TcpStream>>,
    pub key: Option<[u8; 56]>,
}

#[async_trait::async_trait]
impl Actor for Client {}

#[message(result = "(u32, String)")]
pub struct InitClient;

#[async_trait::async_trait]
impl Handler<InitClient> for Client {
    async fn handle(&mut self, _: &mut Context<Self>, _: InitClient) -> (u32, String) {
        let n = self.stream.lock().await.read_u32().await.unwrap_or_else(|_| 0);
        if n == 0 {
            return (400, "Invalid request".to_string());
        }

        if n > INIT_MAX_PACKET_SIZE as u32 {
            return (400, "Request too large".to_string());
        }

        let mut buf = [0u8; INIT_MAX_PACKET_SIZE];
        self.stream.lock().await.read(&mut buf).await.unwrap();

        let req = get_dh_keys();

        let client_pub_key = PublicKey::from_bytes(&buf[0..56]);
        if client_pub_key.is_none() {
            return (400, "Invalid public key".to_string());
        }

        let res = req.1.to_diffie_hellman(&client_pub_key.unwrap());
        if res.is_none() {
            return (500, "Could not create secret".to_string());
        }

        self.key = Some(*res.unwrap().as_bytes());

        use byteorder::{WriteBytesExt, BigEndian};
        use std::io::Write;

        let mut buf = Vec::new();
        buf.write_u32::<BigEndian>(200);
        buf.write(req.0.as_bytes());
        shared_rs::shared::write(self.stream.clone(), &buf).await;

        (200, "OK".to_string())
    }
}

fn get_dh_keys() -> (PublicKey, Secret) {
    let secret = Secret::new(&mut rand_dh::OsRng);
    let pb_key = PublicKey::from(&secret);
    (pb_key, secret)
}