import json
import wasmx

def instantiate():
    setOwner()

def main(dataObj):
    if "getOwner" in dataObj:
        return getOwner()
    if "getCaller" in dataObj:
        return getCaller()
    if "getChainId" in dataObj:
        return getChainId()
    if "getBalance" in dataObj:
        return getBalance(*dataObj["getBalance"])
    if "instantiateAccount" in dataObj:
        return instantiateAccount(*dataObj["instantiateAccount"])
    if "instantiateAccount2" in dataObj:
        return instantiateAccount2(*dataObj["instantiateAccount2"])
    if "getAccount" in dataObj:
        return getAccount_(*dataObj["getAccount"])
    if "getCode" in dataObj:
        return getCode(*dataObj["getCode"])
    raise ValueError('Invalid function')

def setOwner():
    caller = getCaller_()
    store("owner", caller)

def getOwner():
    return load("owner")

def getCaller():
    return getCaller_().encode()

def getCaller_():
    env = getEnv_()
    caller = env["currentCall"]["sender"]
    return wasmx.bech32_bytes_to_string(bytes(caller))

def getChainId():
    env = getEnv_()
    return env["chain"]["chainIdFull"].encode()

def getBalance(addressStr):
    bz = wasmx.bech32_string_to_bytes(addressStr)
    return wasmx.get_balance(bz)

def getEnv_():
    envbuf = wasmx.get_env()
    envstr = arrayBufferToString(envbuf)
    return json.loads(envstr)

def getAccount_(address):
    addrbuf = wasmx.bech32_string_to_bytes(address)
    accountbuf = wasmx.get_account(addrbuf)
    account = json.loads(arrayBufferToString(accountbuf))
    return stringToArrayBuffer(json.dumps(account))

def getCode(address):
    addrbuf = wasmx.bech32_string_to_bytes(address)
    return wasmx.get_code(addrbuf)

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
