#[no_mangle]
pub extern "C" fn alloc(size: usize) -> *mut u8 {
    let mut buffer = Vec::with_capacity(size);
    let ptr = buffer.as_mut_ptr();
    std::mem::forget(buffer); // Prevent Rust from deallocating the memory
    ptr
}

#[no_mangle]
pub extern "C" fn free(ptr: *mut u8) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[link(wasm_import_module = "wasmx")]
extern "C" {
    pub fn getEnv() -> i64;
    pub fn getChainId() -> i64;
    pub fn getCallData() -> i64;
    pub fn getCaller() -> i64;
    pub fn getAddress() -> i64;
    pub fn getBalance(address: i64) -> i64;
    pub fn getCurrentBlock() -> i64;

    pub fn storageStore(key: i64, value: i64);
    pub fn storageLoad(key: i64) -> i64;
    pub fn storageLoadRange(key: i64) -> i64;
    pub fn storageLoadRangePairs(key: i64) -> i64;

    pub fn log(value: i64);
    pub fn emitCosmosEvents(value: i64);
    pub fn getFinishData() -> i64;
    pub fn setFinishData(value: i64);
    pub fn finish(value: i64);
    pub fn revert(message: i64);

    pub fn getAccount(address: i64) -> i64;
    pub fn call(data: i64) -> i64;

    pub fn createAccount(data: i64) -> i64;
    pub fn create2Account(data: i64) -> i64;

    pub fn sha256(value: i64) -> i64;

    pub fn MerkleHash(value: i64) -> i64;

    pub fn LoggerInfo(value: i64);
    pub fn LoggerError(value: i64);
    pub fn LoggerDebug(value: i64);
    pub fn LoggerDebugExtended(value: i64);

    pub fn ed25519Sign(priv_key: i64, msgbz: i64) -> i64;
    pub fn ed25519Verify(pub_key: i64, signature: i64, msgbz: i64) -> i32;
    pub fn ed25519PubToHex(pub_key: i64) -> i64;

    pub fn validate_bech32_address(value: i64) -> i32;
    pub fn addr_humanize(value: i64) -> i64;
    pub fn addr_canonicalize(value: i64) -> i64;
    pub fn addr_equivalent(addr1: i64, addr2: i64) -> i32;
    pub fn addr_humanize_mc(value: i64, prefix: i64) -> i64;
    pub fn addr_canonicalize_mc(value: i64) -> i64;

    pub fn getAddressByRole(value: i64) -> i64;
    pub fn getRoleByAddress(value: i64) -> i64;

    pub fn executeCosmosMsg(value: i64) -> i64;

    pub fn decodeCosmosTxToJson(value: i64) -> i64;
    pub fn verifyCosmosTx(value: i64) -> i64;
}
