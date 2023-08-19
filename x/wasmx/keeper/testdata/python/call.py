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
    res = call(1000000, address, 0, calldata)
    print("-wrapStore-res", res)

def wrapLoad(address):
    calldata = json.dumps({"load":[]})
    res = call_static(1000000, address, calldata)
    print("-wrapLoad-res", res)
    response = json.loads(res)
    data = response["data"]
    print("-wrapLoad-data", data)
    datastr = ''.join(map(chr, data))
    print("-wrapLoad-datastr", datastr)
    return datastr
