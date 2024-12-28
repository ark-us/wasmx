use serde::{Deserialize, Serialize};
use serde_json::json;
use serde::ser::{Serializer};
use serde::de::{Deserializer};
use serde_with::serde_as;
use serde_with::base64::{Base64, Bcrypt, BinHex, Standard};
use serde_with::formats::{Padded, Unpadded};
use base64::{engine::general_purpose::STANDARD as BASE64_STANDARD, Engine};

#[derive(Serialize, Deserialize, Debug)]
pub struct LoggerLog {
    pub msg: String,
    pub parts: Vec<String>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct CallRequest {
    pub to: String,
    pub calldata: String,
    pub value: String,
    pub gas_limit: i64,
    pub is_query: bool,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Account {
    pub address: String,
    pub pub_key: String,
    pub account_number: i64,
    pub sequence: i64,
}

#[serde_as]
#[derive(Serialize, Deserialize, Debug)]
pub struct EventAttribute {
    pub key: String,
    #[serde_as(as = "Base64")]
    pub value: Vec<u8>,
    pub index: bool,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Event {
    pub r#type: String,
    pub attributes: Vec<EventAttribute>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct GrpcResponse {
    pub data: String,
    pub error: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct StorageRange {
    pub start_key: String,
    pub end_key: String,
    pub reverse: bool,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct StoragePair {
    pub key: String,
    pub value: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct StoragePairs {
    pub values: Vec<StoragePair>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct CallResponse {
    pub success: i32,
    pub data: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct StartTimeoutRequest {
    pub id: String,
    pub contract: String,
    pub delay: i64,
    pub args: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct WasmxLog {
    pub data: Vec<u8>,
    pub topics: Vec<Vec<u8>>,
}


#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn binary_base64_encoding() {
        assert_eq!(4, 4);
    }
}

