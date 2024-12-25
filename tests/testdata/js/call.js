import * as wasmx from 'wasmx';

export function instantiate() {}

export function main(dataObj) {
    if (dataObj.store) {
        return wrapStore(...dataObj.store);
    } else if (dataObj.load) {
        return wrapLoad(...dataObj.load);
    }
    throw new Error("invalid function");
}

function wrapStore(addressbech32, value) {
    let calldata = JSON.stringify({"store":[value]})
    let address = wasmx.bech32StringToBytes(addressbech32)
    return wasmx.call(50000000, address, new ArrayBuffer(32), stringToArrayBuffer(calldata))
}

function wrapLoad(addressbech32) {
    let calldata = JSON.stringify({"load":[]})
    let address = wasmx.bech32StringToBytes(addressbech32)
    let res = wasmx.callStatic(50000000, address, stringToArrayBuffer(calldata))
    let response = JSON.parse(arrayBufferToString(res))
    let data = new Uint8Array(Object.values(response.data));
    return data.buffer;
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
