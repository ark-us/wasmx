import * as wasmx from 'wasmx';

export function instantiate() {
    store("javascript");
}

export function main(dataObj) {
    if (dataObj.forward) {
        return forward(...dataObj.forward);
    } else if (dataObj.forward_get) {
        return forward_get(...dataObj.forward_get);
    }
    throw new Error("invalid function");
}

function forward(value, addresses) {
    value = value + arrayBufferToString(load())
    doLog(stringToArrayBuffer(value), []);

    if (addresses.length == 0) {
        return stringToArrayBuffer(value);
    }
    value = value + " -> "
    let addressbech32 = addresses.shift();
    let calldata = JSON.stringify({"forward":[value, addresses]})
    let address = wasmx.bech32StringToBytes(addressbech32)
    let res = wasmx.call(1000000, address, new ArrayBuffer(32), stringToArrayBuffer(calldata));
    let response = JSON.parse(arrayBufferToString(res));
    if (response.success != 0) {
        throw new Error("[js] call failed");
    }
    let data = new Uint8Array(Object.values(response.data));
    return data.buffer;
}

function forward_get(addresses) {
    if (addresses.length == 0) {
        return load();
    }
    let addressbech32 = addresses.shift();
    let calldata = JSON.stringify({"forward_get":[addresses]})
    let address = wasmx.bech32StringToBytes(addressbech32)
    let res = wasmx.callStatic(1000000, address, stringToArrayBuffer(calldata))
    let response = JSON.parse(arrayBufferToString(res))
    if (response.success != 0) {
        throw new Error("[js] call_static failed");
    }
    let data = new Uint8Array([...new Uint8Array(load()), ...new Uint8Array(stringToArrayBuffer(" -> ")), ...response.data]);
    return data.buffer;
}

function store(value_) {
    const key = stringToArrayBuffer("key");
    const value = stringToArrayBuffer(value_);
    wasmx.storageStore(key, value);
}

function load() {
    const key = stringToArrayBuffer("key");
    return wasmx.storageLoad(key);
}

// utils
function stringToArrayBuffer(inputString) {
    const bytes = new Uint8Array(inputString.length);
    for (let i = 0; i < inputString.length; i++) {
        bytes[i] = inputString.charCodeAt(i) & 0xFF;
    }
    return bytes.buffer;
}

function arrayBufferToString(arrayBuffer) {
    const bytes = new Uint8Array(arrayBuffer);
    let result = "";
    for (let i = 0; i < bytes.length; i++) {
        result += String.fromCharCode(bytes[i]);
    }
    return result;
}

// data: ArrayBuffer, topics: []ArrayBuffer
function doLog(databz, topics) {
    let logdata = JSON.stringify({"data": [...new Uint8Array(databz)],"topics":[]})
    wasmx.log(stringToArrayBuffer(logdata))
}
