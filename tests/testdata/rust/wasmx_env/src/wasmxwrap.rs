use serde::{Deserialize, Serialize};
use serde_json::json;
use base64::{engine::general_purpose::STANDARD as BASE64_STANDARD, Engine};
use crate::wasmx::*;
use crate::types::*;
use crate::wasmx::{
    revert as revert_raw,
    finish as finish_raw,
    sha256 as sha256_raw,
    log as log_raw,
    addr_canonicalize as addr_canonicalize_raw,
    addr_equivalent as addr_equivalent_raw,
    addr_humanize as addr_humanize_raw,
};

pub fn decode_ptr_len(value: i64) -> (*const u8, usize) {
    let ptr = ((value >> 32) & 0xFFFF_FFFF) as u32 as *const u8;
    let len = (value & 0xFFFF_FFFF) as u32 as usize;
    (ptr, len)
}

fn encode_ptr_len(data: &[u8]) -> i64 {
    let ptr = data.as_ptr() as u32;
    let len = data.len() as u32;
    ((ptr as i64) << 32) | (len as i64)
}

pub fn logger_info(module: &str, msg: &str, parts: &[String]) {
    let full_msg = format!("{}: {}", module, msg);
    let data = LoggerLog {
        msg: full_msg,
        parts: parts.to_vec(),
    };
    let data_bytes = serde_json::to_vec(&data).expect("Serialization failed");
    let data_encoded = encode_ptr_len(&data_bytes);
    unsafe { LoggerInfo(data_encoded) };
}

pub fn logger_error(module: &str, msg: &str, parts: &[String]) {
    let full_msg = format!("{}: {}", module, msg);
    let data = LoggerLog {
        msg: full_msg,
        parts: parts.to_vec(),
    };
    let data_bytes = serde_json::to_vec(&data).expect("Serialization failed");
    let data_encoded = encode_ptr_len(&data_bytes);
    unsafe { LoggerError(data_encoded) };
}

pub fn logger_debug(module: &str, msg: &str, parts: &[String]) {
    let full_msg = format!("{}: {}", module, msg);
    let data = LoggerLog {
        msg: full_msg,
        parts: parts.to_vec(),
    };
    let data_bytes = serde_json::to_vec(&data).expect("Serialization failed");
    let data_encoded = encode_ptr_len(&data_bytes);
    unsafe { LoggerDebug(data_encoded) };
}

pub fn logger_debug_extended(module: &str, msg: &str, parts: &[String]) {
    let full_msg = format!("{}: {}", module, msg);
    let data = LoggerLog {
        msg: full_msg,
        parts: parts.to_vec(),
    };
    let data_bytes = serde_json::to_vec(&data).expect("Serialization failed");
    let data_encoded = encode_ptr_len(&data_bytes);
    unsafe { LoggerDebugExtended(data_encoded) };
}

pub fn sstore(key: &str, value: &str) {
    let key_encoded = encode_ptr_len(key.as_bytes());
    let value_encoded = encode_ptr_len(value.as_bytes());
    unsafe {
        storageStore(key_encoded, value_encoded);
    }
}

pub fn sload(key: &str) -> String {
    let key_encoded = encode_ptr_len(key.as_bytes());
    let result = unsafe { storageLoad(key_encoded) };
    let (ptr, len) = decode_ptr_len(result);

    if ptr.is_null() || len == 0 {
        return String::new();
    }

    unsafe { String::from_utf8_unchecked(std::slice::from_raw_parts(ptr, len).to_vec()) }
}

pub fn sload_range(key_start: &str, key_end: &str, reverse: bool) -> Vec<String> {
    let req = StorageRange {
        start_key: BASE64_STANDARD.encode(key_start),
        end_key: BASE64_STANDARD.encode(key_end),
        reverse,
    };
    let req_bytes = serde_json::to_vec(&req).unwrap();
    let req_encoded = encode_ptr_len(&req_bytes);
    let result = unsafe { storageLoadRange(req_encoded) };
    let (ptr, len) = decode_ptr_len(result);

    if ptr.is_null() || len == 0 {
        return Vec::new();
    }

    let response: Vec<String> = serde_json::from_slice(unsafe {
        std::slice::from_raw_parts(ptr, len)
    }).unwrap();
    response
}

pub fn sload_range_pairs(start_key: &str, end_key: &str, reverse: bool) -> Vec<StoragePair> {
    let request = StorageRange {
        start_key: BASE64_STANDARD.encode(start_key),
        end_key: BASE64_STANDARD.encode(end_key),
        reverse,
    };

    let request_bytes = serde_json::to_vec(&request).unwrap();
    let request_encoded = encode_ptr_len(&request_bytes);
    let result = unsafe { storageLoadRangePairs(request_encoded) };
    let (ptr, len) = decode_ptr_len(result);

    if ptr.is_null() {
        return Vec::new();
    }

    let result_slice = unsafe { std::slice::from_raw_parts(ptr, len) };
    serde_json::from_slice(result_slice).expect("Failed to parse sload range pairs response")
}

pub fn sha256(data: &[u8]) -> Vec<u8> {
    let data_encoded = encode_ptr_len(data);
    let result = unsafe { sha256_raw(data_encoded) };
    let (ptr, len) = decode_ptr_len(result);

    if ptr.is_null() || len == 0 {
        return Vec::new();
    }

    unsafe { std::slice::from_raw_parts(ptr, len).to_vec() }
}

pub fn ed25519_sign(priv_key: &[u8], message: &[u8]) -> Vec<u8> {
    let priv_key_encoded = encode_ptr_len(priv_key);
    let message_encoded = encode_ptr_len(message);
    let result = unsafe { ed25519Sign(priv_key_encoded, message_encoded) };
    let (ptr, len) = decode_ptr_len(result);

    if ptr.is_null() || len == 0 {
        return Vec::new();
    }

    unsafe { std::slice::from_raw_parts(ptr, len).to_vec() }
}

pub fn ed25519_verify(pub_key: &[u8], signature: &[u8], message: &[u8]) -> bool {
    let pub_key_encoded = encode_ptr_len(pub_key);
    let signature_encoded = encode_ptr_len(signature);
    let message_encoded = encode_ptr_len(message);
    let result = unsafe { ed25519Verify(pub_key_encoded, signature_encoded, message_encoded) };
    result == 1
}

pub fn addr_humanize(value: &[u8]) -> String {
    let value_encoded = encode_ptr_len(value);
    let result = unsafe { addr_humanize_raw(value_encoded) };
    let (ptr, len) = decode_ptr_len(result);

    if ptr.is_null() || len == 0 {
        return String::new();
    }

    unsafe { String::from_utf8_unchecked(std::slice::from_raw_parts(ptr, len).to_vec()) }
}

pub fn addr_canonicalize(value: &str) -> Vec<u8> {
    let value_encoded = encode_ptr_len(value.as_bytes());
    let result = unsafe { addr_canonicalize_raw(value_encoded) };
    let (ptr, len) = decode_ptr_len(result);

    if ptr.is_null() || len == 0 {
        return Vec::new();
    }

    unsafe { std::slice::from_raw_parts(ptr, len).to_vec() }
}

pub fn addr_equivalent(addr1: &str, addr2: &str) -> bool {
    let addr1_encoded = encode_ptr_len(addr1.as_bytes());
    let addr2_encoded = encode_ptr_len(addr2.as_bytes());
    unsafe { addr_equivalent_raw(addr1_encoded, addr2_encoded) == 1 }
}

pub fn revert(message: &str) -> ! {
    let msg_bytes = message.as_bytes();
    let msg_encoded = encode_ptr_len(msg_bytes);
    unsafe { revert_raw(msg_encoded) };
    panic!("{}", message);
}

pub fn get_env() -> Vec<u8> {
    let result = unsafe { getEnv() };
    let (ptr, len) = decode_ptr_len(result);

    if ptr.is_null() || len == 0 {
        return Vec::new();
    }

    unsafe { std::slice::from_raw_parts(ptr, len).to_vec() }
}

pub fn get_chain_id() -> Vec<u8> {
    let result = unsafe { getChainId() };
    let (ptr, len) = decode_ptr_len(result);

    if ptr.is_null() || len == 0 {
        return Vec::new();
    }

    unsafe { std::slice::from_raw_parts(ptr, len).to_vec() }
}

pub fn emit_cosmos_events(events: &[Event]) {
    let events_json = serde_json::to_vec(events).unwrap();
    let events_encoded = encode_ptr_len(&events_json);
    unsafe {
        emitCosmosEvents(events_encoded);
    }
}

pub fn log(event: WasmxLog) {
    let event_json = serde_json::to_vec(&event).unwrap();
    let event_encoded = encode_ptr_len(&event_json);
    unsafe {
        log_raw(event_encoded);
    }
}

pub fn call_contract(req: CallRequest) -> CallResponse {
    let request_bytes = serde_json::to_vec(&req).unwrap();
    let request_encoded = encode_ptr_len(&request_bytes);
    let result = unsafe { call(request_encoded) };
    let (ptr, len) = decode_ptr_len(result);

    if ptr.is_null() {
        panic!("Call response is null");
    }

    let result_slice = unsafe { std::slice::from_raw_parts(ptr, len) };
    serde_json::from_slice(result_slice).expect("Failed to parse call contract response")
}

pub fn get_account(addr: &str) -> Account {
    let canonical_address = addr_canonicalize(addr);
    let address_encoded = encode_ptr_len(&canonical_address);
    let result = unsafe { getAccount(address_encoded) };
    let (ptr, len) = decode_ptr_len(result);

    if ptr.is_null() {
        panic!("Get account response is null");
    }

    let result_slice = unsafe { std::slice::from_raw_parts(ptr, len) };
    serde_json::from_slice(result_slice).expect("Failed to parse account response")
}

pub fn finish(data: &[u8]) {
    let data_encoded = encode_ptr_len(data);
    unsafe {
        finish_raw(data_encoded);
    }
}

pub fn set_finish_data(data: &[u8]) {
    let data_encoded = encode_ptr_len(data);
    unsafe {
        setFinishData(data_encoded);
    }
}

pub fn get_call_data() -> Vec<u8> {
    let result = unsafe { getCallData() };
    let (ptr, len) = decode_ptr_len(result);

    if ptr.is_null() {
        return Vec::new();
    }

    unsafe { std::slice::from_raw_parts(ptr, len).to_vec() }
}

pub fn get_caller() -> Vec<u8> {
    let result = unsafe { getCaller() };
    let (ptr, len) = decode_ptr_len(result);

    if ptr.is_null() {
        return Vec::new();
    }

    unsafe { std::slice::from_raw_parts(ptr, len).to_vec() }
}

pub fn get_address() -> Vec<u8> {
    let result = unsafe { getAddress() };
    let (ptr, len) = decode_ptr_len(result);

    if ptr.is_null() {
        return Vec::new();
    }

    unsafe { std::slice::from_raw_parts(ptr, len).to_vec() }
}
