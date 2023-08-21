from wasmx import storage_store, storage_load

def instantiate(initvalue: str):
    store(initvalue)

def main(input):
    if "store" in input:
        return store(*input["store"])
    if "load" in input:
        return load()
    raise ValueError('Invalid function')

def store(a: str):
    value = a.encode()
    key = "pystore".encode()
    storage_store(key, value)

def load() -> str:
    key = "pystore".encode()
    value = storage_load(key)
    return value
