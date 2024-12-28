use wasmx_env::wasmxwrap::{LoggerDebug, LoggerError, LoggerInfo};
use crate::types::MODULE_NAME;

pub fn logger_info(msg: &str, parts: &[(&str, &str)]) {
    LoggerInfo(
        MODULE_NAME,
        msg,
        &parts.iter().map(|(k, v)| format!("{}:{}", k, v)).collect::<Vec<String>>(),
    );
}

pub fn logger_error(msg: &str, parts: &[(&str, &str)]) {
    LoggerError(
        MODULE_NAME,
        msg,
        &parts.iter().map(|(k, v)| format!("{}:{}", k, v)).collect::<Vec<String>>(),
    );
}

pub fn logger_debug(msg: &str, parts: &[(&str, &str)]) {
    LoggerDebug(
        MODULE_NAME,
        msg,
        &parts.iter().map(|(k, v)| format!("{}:{}", k, v)).collect::<Vec<String>>(),
    );
}

pub fn revert(message: &str) -> ! {
    logger_error("revert", &[("err", message), ("module", MODULE_NAME)]);
    wasmx::revert(message.as_bytes().to_vec());
    panic!("{}", message);
}
