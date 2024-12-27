use serde::{Deserialize, Serialize};
use crate::types::{CallData, MsgInitialize};
use wasmx_env::wasmxwrap::{get_call_data, revert};

pub fn get_call_data_initialize() -> MsgInitialize {
    // Deserialize initialization call data
    serde_json::from_slice(&get_call_data())
        .unwrap_or_else(|err| revert(&format!("Invalid call data for initialization: {}", err)))
}

pub fn get_call_data_wrap() -> CallData {
    // Deserialize wrapper call data
    serde_json::from_slice(&get_call_data())
        .unwrap_or_else(|err| revert(&format!("Invalid wrapper call data: {}", err)))
}
