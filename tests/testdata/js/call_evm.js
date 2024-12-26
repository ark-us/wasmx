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
    return base64ToArrayBuffer2(response.data);
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


// base64 should be a common import

// Base64 character set
const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
const base64Map = {};

// Create a mapping from character to its 6-bit value
for (let i = 0; i < base64Chars.length; i++) {
    base64Map[base64Chars[i]] = i;
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
