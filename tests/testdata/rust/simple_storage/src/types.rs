use serde::{Deserialize, Serialize};

pub const MODULE_NAME: &str = "simple_storage";

#[derive(Serialize, Deserialize, Debug)]
pub struct MsgSet {
    pub key: String,
    pub value: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct MsgGet {
    pub key: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct MsgInitialize {
    pub key: String,
    pub value: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct CallData {
    pub set: Option<MsgSet>,
    pub get: Option<MsgGet>,
}
