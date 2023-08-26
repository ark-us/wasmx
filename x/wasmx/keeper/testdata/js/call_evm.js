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
    let calldata = "60fe47b10000000000000000000000000000000000000000000000000000000000000007"
    let address = wasmx.bech32StringToBytes(addressbech32)
    return wasmx.call(1000000, address, new ArrayBuffer(32), hexStringToArrayBuffer(calldata))
}

function wrapLoad(addressbech32) {
    let calldata = "6d4ce63c"
    let address = wasmx.bech32StringToBytes(addressbech32)
    let res = wasmx.callStatic(1000000, address, hexStringToArrayBuffer(calldata))
    let response = JSON.parse(arrayBufferToString(res))
    let data = new Uint8Array(Object.values(response.data));
    return data.buffer;
}

const hexStringToArrayBuffer = hexString => new Uint8Array(hexString.match(/.{1,2}/g).map(byte => parseInt(byte, 16))).buffer;

function arrayBufferToString(arrayBuffer) {
    const bytes = new Uint8Array(arrayBuffer);
    let result = "";
    for (let i = 0; i < bytes.length; i++) {
        result += String.fromCharCode(bytes[i]);
    }
    return result;
}

