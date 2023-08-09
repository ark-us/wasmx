import sys
import os

from wasmx import storage_store, storage_load

# data = json.loads('{"one" : "1", "two" : "2", "three" : "3"}')

def instantiate():
    print("simple_storage py instantiate")

def main(a):
    print("simple_storage main", a)
    store(a)
    # return ""
    return load()

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

res = main(*sys.argv[1:])
print("----res", res)

# print("__file__", __file__)
# print("__name__", __name__)
resfilepath = "./testdata/result.py"

file1 = open(resfilepath, "w")
# content = file1.read()
# print("content", content)
file1.write(res)
file1.close()
