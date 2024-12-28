pub mod types;
pub mod calldata;
pub mod storage;

use serde::{Serialize, Deserialize};
use wasmx_env::wasmxwrap::{revert, finish, get_call_data};
use crate::calldata::{get_call_data_wrap, get_call_data_initialize};
use crate::types::{CallData, MsgSet};
use crate::storage::{set, get};

#[no_mangle]
pub extern "C" fn wasmx_env_i64_2() {}

#[no_mangle]
pub extern "C" fn memory_rust_i64_1() {}

#[no_mangle]
pub extern "C" fn instantiate() {
    let calld = get_call_data_initialize();
    let msg_set = &MsgSet {
        key: calld.key,
        value: calld.value,
    };
    set(msg_set);
}

#[no_mangle]
pub extern "C" fn main() {
    let calld = get_call_data_wrap();
    let result = main_internal(&calld);
    finish(result.as_slice());
}

fn main_internal(calld: &CallData) -> Vec<u8> {
    if let Some(set_data) = &calld.set {
        set(set_data);
        Vec::new()
    } else if let Some(get_data) = &calld.get {
        get(get_data)
    } else {
        let calldraw = get_call_data();
        let calldstr = String::from_utf8(calldraw).unwrap_or_else(|_| "Invalid UTF-8".to_string());
        revert(&format!("Invalid function call data: {}", calldstr));
        Vec::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn it_works() {
        assert_eq!(4, 4);
    }
}
