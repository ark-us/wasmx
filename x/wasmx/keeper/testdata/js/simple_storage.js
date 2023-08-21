import * as wasmx from 'wasmx';

export function instantiate(dataObj) {
    return store(dataObj);
}

export function main(dataObj) {
    if (dataObj.store) {
        return store(...dataObj.store);
    } else if (dataObj.load) {
        return load();
    }
    throw new Error("no valid function");
}

function store(value_) {
    const key = stringToArrayBuffer("jsstore");
    const value = stringToArrayBuffer(value_);
    wasmx.storageStore(key, value);
}

function load() {
    const key = stringToArrayBuffer("jsstore");
    return wasmx.storageLoad(key);
}

function stringToArrayBuffer(inputString) {
    const bytes = new Uint8Array(inputString.length);
    for (let i = 0; i < inputString.length; i++) {
        bytes[i] = inputString.charCodeAt(i) & 0xFF;
    }
    return bytes.buffer;
}
