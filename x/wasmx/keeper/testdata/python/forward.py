import json
from wasmx import call, call_static, bech32_string_to_bytes, log, storage_store, storage_load, log

def instantiate():
    store("python")

def main(input):
    print("--py-main", input)
    if "forward" in input:
        return forward(*input["forward"])
    if "forward_get" in input:
        return forward_get(*input["forward_get"])
    raise ValueError('Invalid function')

def forward(value, addresses):
    value = value + load().decode()
    print("--py-forward", value, addresses)
    doLog(value.encode(), [])

    if len(addresses) == 0:
        return value.encode()

    value = value + " -> "
    addressbech32 = addresses.pop(0)
    print("--py-forward call: ", addressbech32, addresses)
    calldata = json.dumps({"forward":[value, addresses]})
    print("--py-forward call data: ", calldata)
    address = bech32_string_to_bytes(addressbech32)
    res = call(1000000, address, 0, calldata.encode())
    response = json.loads(res.decode())
    print("--py-success", response["success"])
    if response["success"] != 0:
        raise ValueError('[py] call failed')
    return bytes(response["data"])

def forward_get(addresses):
    if len(addresses) == 0:
        return load()
    addressbech32 = addresses.pop(0)
    calldata = json.dumps({"forward_get":[addresses]})
    address = bech32_string_to_bytes(addressbech32)
    res = call_static(1000000, address, calldata.encode())
    response = json.loads(res)
    if response["success"] != 0:
        raise ValueError('[py] call_static failed')
    data = bytes(response["data"])
    return load() + " -> ".encode() + data

def store(a):
    value = a.encode()
    key = "key".encode()
    storage_store(key, value)

def load():
    key = "key".encode()
    value = storage_load(key)
    return value

def doLog(databz: bytes, topicsbz: list[bytes]):
    data = [x for x in databz]
    topics = [[y for y in x] for x in topicsbz]
    logdata = json.dumps({"data": data,"topics": topics})
    print("--py-forward logdata-", logdata)
    log(logdata.encode())
