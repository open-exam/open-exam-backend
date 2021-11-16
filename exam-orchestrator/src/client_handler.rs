use aes::Aes128;
use block_modes::{block_padding::Pkcs7, BlockMode, Cbc};
use std::{net::SocketAddr, sync::Arc};
use tokio::{net::TcpStream, sync::Mutex};
use x448::{PublicKey, Secret};
use xactor::*;

const MAX_PACKET_SIZE: usize = 1048576;
const INIT_MAX_PACKET_SIZE: usize = 512;

type Aes128Cbc = Cbc<Aes128, Pkcs7>;

pub struct Client {
    pub id: usize,
    pub stream: Arc<Mutex<TcpStream>>,
    pub key: Option<[u8; 56]>,
    exam_service: Option<SocketAddr>,
    client_integrity_service: Option<SocketAddr>,
}

impl Client {
    pub fn new(id: usize, stream: Arc<Mutex<TcpStream>>, key: Option<[u8; 56]>) -> Self {
        Client {
            id,
            stream,
            key,
            exam_service: None,
            client_integrity_service: None,
        }
    }
}

#[async_trait::async_trait]
impl Actor for Client {
    async fn started(&mut self, ctx: &mut Context<Self>) -> Result<()> {
        self.exam_service = Some(
            std::env::var("exam_service")
                .ok()
                .unwrap()
                .parse::<SocketAddr>()
                .unwrap(),
        );
        self.client_integrity_service = Some(
            std::env::var("client_integrity_service")
                .ok()
                .unwrap()
                .parse::<SocketAddr>()
                .unwrap(),
        );
        Ok(())
    }
}

#[message(result = "(u32, String)")]
pub struct InitClient;

#[async_trait::async_trait]
impl Handler<InitClient> for Client {
    async fn handle(&mut self, _: &mut Context<Self>, _: InitClient) -> (u32, String) {
        use tokio::io::AsyncReadExt;
        let n = self
            .stream
            .lock()
            .await
            .read_u32()
            .await
            .unwrap_or_else(|_| 0);
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

        let mut buf = Vec::new();
        {
            use byteorder::{BigEndian, WriteBytesExt};
            use std::io::Write;
            buf.write_u32::<BigEndian>(200);
            buf.write(req.0.as_bytes());
        }
        shared_rs::shared::write(&self.stream, &buf).await;

        (200, String::from_utf8(buf[56..].to_vec()).unwrap())
    }
}

#[message(result = "()")]
pub struct Request(pub u32);

#[async_trait::async_trait]
impl Handler<Request> for Client {
    async fn handle(&mut self, ctx: &mut Context<Self>, req: Request) {
        use tokio::io::AsyncReadExt;
        let n = req.0;
        if n <= 4 {
            return;
        }

        if n > MAX_PACKET_SIZE as u32 {
            return;
        }

        let mut buf = Vec::new();
        let service = self
            .stream
            .lock()
            .await
            .read_u8()
            .await
            .unwrap_or_else(|_| 0);
        self.stream.lock().await.read(&mut buf).await.unwrap();

        let raw_data = decrypt(buf, &self.key.unwrap()[..]);
        if raw_data.len() < 20 {
            ctx.address()
                .call(Notification(
                    0,
                    r#"{"error": "could not parse request"#.to_string(),
                ))
                .await;
            return;
        }

        let counter = shared_rs::shared::read_u32(&raw_data[0..4]);
        let request_name = shared_rs::shared::read_bytes_as_string(&raw_data[4..20], 16).unwrap();

        let client = hyper::Client::new();
        let service_req = hyper::Request::builder()
            .method("POST")
            .uri(format!(
                "http://{}/{}",
                self.exam_service.unwrap(),
                request_name
            ))
            .body(hyper::Body::from(raw_data[4..].to_vec()))
            .unwrap();

        let resp = client.request(service_req).await;

        if resp.is_err() {
            ctx.address()
                .call(Notification(
                    counter,
                    r#","error":"#.to_string() + "Could not connect to internal service",
                ))
                .await;
            return;
        }

        let res = resp.unwrap();
        ctx.address()
            .call(Notification(
                counter,
                String::from_utf8(
                    hyper::body::to_bytes(res.into_body()).await.unwrap()[..].to_vec(),
                )
                .unwrap(),
            ))
            .await;
    }
}

#[message(result = "()")]
pub struct Notification(pub u32, pub String);

#[async_trait::async_trait]
impl Handler<Notification> for Client {
    async fn handle(&mut self, _: &mut Context<Self>, req: Notification) {
        let mut buf: Vec<u8> = Vec::new();
        {
            use byteorder::{BigEndian, WriteBytesExt};
            buf.write_u32::<BigEndian>(req.0);
        }

        shared_rs::shared::write(
            &self.stream,
            &encrypt(
                buf.into_iter()
                    .chain(req.1.as_bytes().to_vec())
                    .collect::<Vec<u8>>(),
                &self.key.unwrap()[..],
            ),
        )
        .await;
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
    let iv = shared_rs::shared::random_bytes(16);

    buf.reserve_exact((buf.len() / 16 + 1) * 16);

    let cipher = Aes128Cbc::new_from_slices(&key, &iv.to_vec()).unwrap();
    let pos = buf.len();

    iv.into_iter()
        .chain(cipher.encrypt(&mut buf, pos).unwrap().to_vec().into_iter())
        .collect::<Vec<u8>>()
}
