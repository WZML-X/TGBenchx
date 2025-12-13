use grammers_client::grammers_tl_types::LAYER;
use grammers_client::types;
use grammers_client::{Client, ClientConfiguration};
use grammers_mtsender::SenderPool;
use grammers_session::storages::MemorySession;
use std::env;
use std::sync::Arc;
use std::time::SystemTime;

fn now() -> f64 {
    SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap()
        .as_secs_f64()
}

#[tokio::main]
async fn main() {
    let version = "0.8.1"; // InitParams is gone, using hardcoded version or could fetch from Cargo.toml

    let api_id = env::var("APP_ID")
        .map(|var| var.parse::<i32>().unwrap())
        .unwrap_or(6);
    let api_hash = env::var("API_HASH").unwrap_or(String::new());
    let bot_token = env::var("BOT_TOKEN").unwrap();
    let flood_sleep_threshold = env::var("FLOOD_WAIT_SLEEP_TIME")
        .map(|var| var.parse::<u32>().unwrap())
        .unwrap_or(10);
    let message_link = env::var("MESSAGE_LINK").unwrap();

    let session = Arc::new(MemorySession::default());
    let pool = SenderPool::new(session, api_id);

    // ClientConfiguration only has flood_sleep_threshold according to error message
    // If it has other fields, we rely on default, or struct update syntax if all fields are pub
    // But since we saw "available fields are: flood_sleep_threshold", let's assume that's it or use struct update.
    let config = ClientConfiguration {
        flood_sleep_threshold,
        ..ClientConfiguration::default()
    };

    let app = Client::with_configuration(&pool, config);

    app.bot_sign_in(&bot_token, &api_hash).await.unwrap();

    let mut link = message_link.split("/").skip(3);
    let chat_id = link.next().unwrap();
    let s_message_id = link.next().unwrap().parse::<i32>().unwrap();

    // Peer probably doesn't need .pack() anymore, passing Peer directly or &Peer
    let chat = app.resolve_username(chat_id).await.unwrap().unwrap();

    let _t1 = now();
    let message = app
        .get_messages_by_id(&chat, &[s_message_id])
        .await
        .unwrap()
        .pop()
        .unwrap()
        .unwrap();

    let media = message.media().unwrap();
    let file_size = match media {
        types::Media::Document(ref document) => document.size(),
        _ => panic!("Expected document media"),
    };

    let t2 = now();
    let filename = "g1.tmp";
    // media implements Downloadable directly?
    app.download_media(&media, filename)
        .await
        .unwrap();
    let t3 = now();
    let download_start_time = t2;
    let download_end_time = t2;
    let download_time_taken = t3 - t2;

    let t4 = now();
    let uploaded = app.upload_file(filename).await.unwrap();
    app.send_message(
        &chat,
        types::InputMessage::new().text("Grammers")
            .file(uploaded)
            .reply_to(Some(s_message_id)),
    )
    .await
    .unwrap();
    let t5 = now();
    let upload_start_time = t4;
    let upload_end_time = t5;
    let upload_time_taken = t5 - t4;

    drop(app);

    println!(
        r#"{{
  "version": "{version}",
  "layer": {LAYER},
  "file_size": {file_size},
  "download": {{
    "start_time": {download_start_time},
    "end_time": {download_end_time},
    "time_taken": {download_time_taken}
  }},
  "upload": {{
    "start_time": {upload_start_time},
    "end_time": {upload_end_time},
    "time_taken": {upload_time_taken}
  }}
}}"#,
    );
}
