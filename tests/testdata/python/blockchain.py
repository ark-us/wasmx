import json
import wasmx

def instantiate(dataObj):
    return store(*dataObj)

def main(dataObj):
    if "justError" in dataObj:
        return justError()
    if "getEnv" in dataObj:
        return getEnv_()
    if "getCallData" in dataObj:
        return getCallData_()
    if "store" in dataObj:
        return store(*dataObj["store"])
    if "load" in dataObj:
        return load()
    if "getBlockHash" in dataObj:
        return getBlockHash_(*dataObj["getBlockHash"])
    if "getAccount" in dataObj:
        return getAccount_(*dataObj["getAccount"])
    if "getBalance" in dataObj:
        return getBalance_(*dataObj["getBalance"])
    if "keccak256" in dataObj:
        return keccak256_(*dataObj["keccak256"])
    if "instantiateAccount" in dataObj:
        return instantiateAccount(*dataObj["instantiateAccount"])
    if "instantiateAccount2" in dataObj:
        return instantiateAccount2(*dataObj["instantiateAccount2"])
    wasmx.set_exit_code(1, 'Invalid function')

def justError():
    wasmx.set_exit_code(1, 'just error')

def getEnv_():
    envbuf = wasmx.get_env()
    envstr = arrayBufferToString(envbuf)
    env = json.loads(envstr)
    return stringToArrayBuffer(json.dumps(env))

def getCallData_():
    calldbuf = wasmx.get_calldata()
    data = json.loads(arrayBufferToString(calldbuf))
    return stringToArrayBuffer(json.dumps(data))

def getBlockHash_(blockNumber):
    return wasmx.get_blockhash(blockNumber)

def getAccount_(address):
    addrbuf = wasmx.bech32_string_to_bytes(address)
    accountbuf = wasmx.get_account(addrbuf)
    account = json.loads(arrayBufferToString(accountbuf))
    return stringToArrayBuffer(json.dumps(account))

def getBalance_(address):
    bz = wasmx.bech32_string_to_bytes(address)
    return wasmx.get_balance(bz)

def keccak256_(data):
    databz = stringToArrayBuffer(data)
    return wasmx.keccak256(databz)

def instantiateAccount(codeId, initMsg, balance):
    msgbuf = hexStringToArrayBuffer(initMsg)
    balancebuf = hexStringToArrayBuffer(balance)
    return wasmx.instantiate(codeId, msgbuf, balancebuf)

def instantiateAccount2(codeId, salt, initMsg, balance):
    return wasmx.instantiate2(codeId, hexStringToArrayBuffer(salt), hexStringToArrayBuffer(initMsg), hexStringToArrayBuffer(balance))

def store(key_, value_):
    key = stringToArrayBuffer(key_)
    value = stringToArrayBuffer(value_)
    wasmx.storage_store(key, value)

def load(key_):
    key = stringToArrayBuffer(key_)
    return wasmx.storage_load(key)

# utils

def stringToArrayBuffer(inputString):
    return inputString.encode()

def arrayBufferToString(arrayBuffer):
    return arrayBuffer.decode()

def hexStringToArrayBuffer(hex_string):
    return bytes.fromhex(hex_string)
