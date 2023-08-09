import json
from wasmx import storage_store, storage_load

def instantiate(initvalue: str):
    print("simple_storage py instantiate", initvalue)

def main(a):
    print("simple_storage main", a)
    input = json.loads(a)
    if "store" in input:
        return store(*input["store"])
    if "load" in input:
        return load()
    raise ValueError('Invalid function')

def store(a: str):
    # value = a.encode()
    # key = "pystore".encode()
    # storage_store(key, value)
    print("store value", a)
    storage_store("pystore", a)

def load() -> str:
    # key = "pystore".encode()
    # return bytearray.decode(value)
    value = storage_load("pystore")
    print("---load value", value)
    return value
