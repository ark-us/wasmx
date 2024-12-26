import json
import base64
from wasmx import call, call_static, bech32_string_to_bytes

def instantiate():
    pass

def main(input=None):
    if input and "store" in input:
        return wrapStore(*input["store"])
    if input and "load" in input:
        return wrapLoad(*input["load"])
    else:
        raise ValueError('Invalid function')

def wrapStore(addressbech32, value):
    calldata = json.dumps({"store":[value]})
    address = bech32_string_to_bytes(addressbech32)
    res = call(50000000, address, 0, calldata.encode())
    response = json.loads(res.decode())
    data = response["data"]
    if data == "":
        return b''
    try:
        decoded_data = base64.b64decode(data)
    except base64.binascii.Error as e:
        print(f"error decoding base64 data: {e}")
        raise ValueError("invalid base64 data received from load")
    return decoded_data

def wrapLoad(addressbech32):
    calldata = json.dumps({"load":[]})
    address = bech32_string_to_bytes(addressbech32)
    res = call_static(50000000, address, calldata.encode())
    response = json.loads(res)
    data = response["data"]
    if data == "":
        return b''
    decoded_data = b''
    try:
        decoded_data = base64.b64decode(data)
    except base64.binascii.Error as e:
        print(f"error decoding base64 data: {e}")
        raise ValueError("invalid base64 data received from load")
    return decoded_data + b'23'
