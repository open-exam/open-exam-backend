use std::{sync::Arc};
use aes::Aes128;
use block_modes::{BlockMode, Cbc, block_padding::Pkcs7};
use tokio::{io::AsyncReadExt, net::TcpStream, sync::Mutex};
use x448::{PublicKey, Secret};
use xactor::*;

const MAX_PACKET_SIZE: usize = 1048576;
const INIT_MAX_PACKET_SIZE: usize = 512;

type Aes128Cbc = Cbc<Aes128, Pkcs7>;

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

        let mut buf = Vec::new();
        {
            self.stream.lock().await.read(&mut buf).await.unwrap();
        }

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
        shared_rs::shared::write(&self.stream, &buf).await;

        (200, String::from_utf8(buf[56..].to_vec()).unwrap())
    }
}

#[message(result = "()")]
pub struct Request(pub u32);

#[async_trait::async_trait]
impl Handler<Request> for Client {
    async fn handle(&mut self, _: &mut Context<Self>, req: Request) -> () {
        let n = req.0;
        if n <= 4 {
            return;
        }

        if n > MAX_PACKET_SIZE as u32 {
            return;
        }

        let mut buf = Vec::new();
        let service = self.stream.lock().await.read_u32().await.unwrap_or_else(|_| 0);
        self.stream.lock().await.read(&mut buf).await.unwrap();

        let raw_data = decrypt(buf, &self.key.unwrap()[..]);
        
        match service {
            1 => {

            },
            2 => {

            },
            _ => {

            }
        }
    }
}

#[message(result = "()")]
pub struct Notification(pub String);

#[async_trait::async_trait]
impl Handler<Notification> for Client {
    async fn handle(&mut self, _: &mut Context<Self>, req: Notification) -> () {
        use byteorder::{WriteBytesExt, BigEndian};
        use std::io::Write;

        shared_rs::shared::write(&self.stream, &encrypt(req.0.as_bytes().to_vec(), &self.key.unwrap()[..])).await;
    }
}

fn get_dh_keys() -> (PublicKey, Secret) {
    let secret = Secret::new(&mut rand_dh::OsRng);
    let pb_key = PublicKey::from(&secret);
    (pb_key, secret)
}

fn decrypt(buf: Vec<u8>, key: &[u8]) -> Vec<u8> {
    let cipher = Aes128Cbc::new_from_slices(&key, &buf[0..4].to_vec()).unwrap();
    cipher.decrypt(&mut buf[4..].to_vec()).unwrap().to_vec()
}

fn encrypt(mut buf: Vec<u8>, key: &[u8]) -> Vec<u8> {
    let iv = shared_rs::shared::random_bytes(4);
    let cipher = Aes128Cbc::new_from_slices(&key, &iv.to_vec()).unwrap();
    let pos = buf.len();
    iv.into_iter().chain(cipher.encrypt(&mut buf, pos).unwrap().to_vec().into_iter()).collect::<Vec<u8>>()
}