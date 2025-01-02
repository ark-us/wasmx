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

#[link(wasm_import_module = "wasmxcore")]
extern "C" {
    pub fn grpcRequest(data: i64) -> i64;
    pub fn startTimeout(req: i64);
    pub fn cancelTimeout(req: i64);
    pub fn startBackgroundProcess(req: i64);
    pub fn writeToBackgroundProcess(req: i64) -> i64;
    pub fn readFromBackgroundProcess(req: i64) -> i64;
    pub fn externalCall(data: i64) -> i64;
}
