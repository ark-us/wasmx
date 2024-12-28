from wasmx import storage_store, storage_load

def instantiate(initvalue: str):
    store(initvalue)

def main(input=None):
    if input and "store" in input:
        return store(*input["store"])
    if input and "load" in input:
        return load()
    else:
        raise ValueError('Invalid function')

def store(a: str):
    value = a.encode()
    key = "pystore".encode()
    storage_store(key, value)

def load() -> bytes:
    key = "pystore".encode()
    value = storage_load(key)
    return value
