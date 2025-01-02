use serde::{Deserialize, Serialize};
use serde_json::json;
use serde::ser::{Serializer};
use serde::de::{Deserializer};
use serde_with::serde_as;
use serde_with::base64::{Base64, Bcrypt, BinHex, Standard};
use serde_with::formats::{Padded, Unpadded};
use base64::{engine::general_purpose::STANDARD as BASE64_STANDARD, Engine};

#[derive(Serialize, Deserialize, Debug)]
pub struct GrpcResponse {
    pub data: String,
    pub error: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct StartTimeoutRequest {
    pub id: String,
    pub contract: String,
    pub delay: i64,
    pub args: String,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn binary_base64_encoding() {
        assert_eq!(4, 4);
    }
}

