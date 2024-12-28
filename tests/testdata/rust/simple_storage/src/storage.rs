use wasmx_env::wasmxwrap::{sstore, sload, logger_debug, log};
use wasmx_env::types::{WasmxLog};
use crate::types::{MsgGet, MsgSet, MODULE_NAME};

pub fn set(req: &MsgSet) {
    logger_debug(MODULE_NAME, "set", &["key".to_string(), req.key.clone(), "value".to_string(), req.value.clone()]);

    let event = WasmxLog{
        data: Vec::new(),
        topics: vec![b"hello".to_vec()],
    };
    log(event);
    sstore(&req.key, &req.value);
}

pub fn get(req: &MsgGet) -> Vec<u8> {
    let resp = sload(&req.key);
    logger_debug(MODULE_NAME, "get", &["key".to_string(), req.key.clone(), "value".to_string(), resp.clone()]);
    resp.into_bytes()
}
