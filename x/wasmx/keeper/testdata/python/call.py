import json
from wasmx import call, call_static, bech32_string_to_bytes

def instantiate():
    pass

def main(input):
    if "store" in input:
        return wrapStore(*input["store"])
    if "load" in input:
        return wrapLoad(*input["load"])
    raise ValueError('Invalid function')

def wrapStore(addressbech32, value):
    calldata = json.dumps({"store":[value]})
    address = bech32_string_to_bytes(addressbech32)
    res = call(1000000, address, 0, calldata.encode())
    response = json.loads(res.decode())
    return response["data"]

def wrapLoad(addressbech32):
    calldata = json.dumps({"load":[]})
    address = bech32_string_to_bytes(addressbech32)
    res = call_static(1000000, address, calldata.encode())
    response = json.loads(res)
    data = response["data"]
    return bytes(data) + b'23'
