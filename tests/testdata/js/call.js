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
    return base64ToArrayBuffer2(response.data);
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

// base64 should be a common import

// Base64 character set
const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
const base64Map = {};

// Create a mapping from character to its 6-bit value
for (let i = 0; i < base64Chars.length; i++) {
    base64Map[base64Chars[i]] = i;
}

/**
 * Converts a Base64-encoded string to an ArrayBuffer without using atob/btoa.
 * @param {string} base64 - The Base64-encoded string.
 * @returns {ArrayBuffer} - The decoded ArrayBuffer.
 * @throws Will throw an error if an invalid Base64 character is encountered.
 */
function base64ToArrayBuffer(base64) {
    // Remove any padding characters
    base64 = base64.replace(/=+$/, '');

    let binaryStr = '';

    for (let i = 0; i < base64.length; i++) {
        const c = base64[i];
        const val = base64Map[c];

        if (val === undefined) {
            throw new Error(`Invalid Base64 character: ${c}`);
        }

        binaryStr += val.toString(2).padStart(6, '0');
    }

    // Split the binary string into 8-bit chunks
    const bytes = [];

    for (let i = 0; i < binaryStr.length; i += 8) {
        const byteStr = binaryStr.substr(i, 8);
        if (byteStr.length === 8) { // Only process complete bytes
            bytes.push(parseInt(byteStr, 2));
        }
    }

    // Create an ArrayBuffer and populate it with the bytes
    const arrayBuffer = new ArrayBuffer(bytes.length);
    const uint8Array = new Uint8Array(arrayBuffer);

    for (let i = 0; i < bytes.length; i++) {
        uint8Array[i] = bytes[i];
    }

    return arrayBuffer;
}

/**
 * Optimized Base64 to ArrayBuffer conversion using bitwise operations.
 * @param {string} base64 - The Base64-encoded string.
 * @returns {ArrayBuffer} - The decoded ArrayBuffer.
 * @throws Will throw an error if an invalid Base64 character is encountered.
 */
function base64ToArrayBuffer2(base64) {
    // Remove any padding characters
    base64 = base64.replace(/=+$/, '');

    const byteLength = Math.floor((base64.length * 6) / 8);
    const arrayBuffer = new ArrayBuffer(byteLength);
    const uint8Array = new Uint8Array(arrayBuffer);

    let byteIndex = 0;
    let buffer = 0;
    let bitsLeft = 0;

    for (let i = 0; i < base64.length; i++) {
        const c = base64[i];
        const val = base64Map[c];

        if (val === undefined) {
            throw new Error(`Invalid Base64 character: ${c}`);
        }

        buffer = (buffer << 6) | val;
        bitsLeft += 6;

        if (bitsLeft >= 8) {
            bitsLeft -= 8;
            uint8Array[byteIndex++] = (buffer >> bitsLeft) & 0xFF;
        }
    }

    return arrayBuffer;
}
