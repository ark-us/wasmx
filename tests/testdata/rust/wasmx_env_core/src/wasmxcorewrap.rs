use serde::{Deserialize, Serialize};
use serde_json::json;
use base64::{engine::general_purpose::STANDARD as BASE64_STANDARD, Engine};
use crate::wasmx::*;
use crate::types::*;

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

pub fn start_timeout(req: StartTimeoutRequest) {
    let req_bytes = serde_json::to_vec(&req).unwrap();
    let req_encoded = encode_ptr_len(&req_bytes);
    unsafe {
        startTimeout(req_encoded);
    }
}

pub fn cancel_timeout(id: &str) {
    let req = serde_json::json!({ "id": id }).to_string();
    let req_encoded = encode_ptr_len(req.as_bytes());
    unsafe {
        cancelTimeout(req_encoded);
    }
}

pub fn grpc_request(ip: &str, contract: &[u8], data: &str) -> GrpcResponse {
    let contract_address = BASE64_STANDARD.encode(contract);
    let req = serde_json::json!({
        "ip_address": ip,
        "contract": contract_address,
        "data": data,
    })
    .to_string();
    let req_bytes = req.into_bytes();
    let req_encoded = encode_ptr_len(&req_bytes);

    let result = unsafe { grpcRequest(req_encoded) };
    let (ptr, len) = decode_ptr_len(result);

    if ptr.is_null() {
        panic!("GRPC response is null");
    }

    let result_slice = unsafe { std::slice::from_raw_parts(ptr, len) };
    serde_json::from_slice(result_slice).expect("Invalid GRPC response")
}
