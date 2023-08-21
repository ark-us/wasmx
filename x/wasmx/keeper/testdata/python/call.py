import json
from wasmx import call, call_static

def instantiate():
    pass

def main(input):
    if "store" in input:
        return wrapStore(*input["store"])
    if "load" in input:
        return wrapLoad(*input["load"])
    raise ValueError('Invalid function')

def wrapStore(address, value):
    calldata = json.dumps({"store":[value]})
    res = call(1000000, address, 0, calldata.encode())
    response = json.loads(res.decode())
    return response["data"]

def wrapLoad(address):
    calldata = json.dumps({"load":[]})
    res = call_static(1000000, address, calldata.encode())
    response = json.loads(res)
    data = response["data"]
    return bytes(data) + b'23'
