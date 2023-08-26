import * as wasmx from 'wasmx';

export function instantiate(dataObj) {
    return store(...dataObj);
}

export function main(dataObj) {
    if (dataObj.justError) return justError();
    if (dataObj.getEnv) return getEnv_();
    if (dataObj.getCallData) return getCallData_();
    if (dataObj.store) return store(...dataObj.store);
    if (dataObj.load) return load();
    if (dataObj.getBlockHash) return getBlockHash_(...dataObj.getBlockHash);
    if (dataObj.getAccount) return getAccount_(...dataObj.getAccount);
    if (dataObj.getBalance) return getBalance_(...dataObj.getBalance);
    if (dataObj.keccak256) return keccak256_(...dataObj.keccak256);
    if (dataObj.instantiateAccount) return instantiateAccount(...dataObj.instantiateAccount);
    if (dataObj.instantiateAccount2) return instantiateAccount2(...dataObj.instantiateAccount2);
    throw new Error("no valid function");
}

function justError() {
    throw new Error("just error");
}

function getEnv_() {
    const envbuf = wasmx.getEnv();
    const envstr = arrayBufferToString(envbuf);
    const env = JSON.parse(envstr);
    return stringToArrayBuffer(JSON.stringify(env));
}

function getCallData_() {
    const calldbuf = wasmx.getCallData();
    const data = JSON.parse(arrayBufferToString(calldbuf));
    return stringToArrayBuffer(JSON.stringify(data));
}

function getBlockHash_(blockNumber) {
    return wasmx.getBlockHash(blockNumber);
}

function getAccount_(address) {
    const addrbuf = wasmx.bech32StringToBytes(address);
    const accountbuf = wasmx.getAccount(addrbuf);
    const account = JSON.parse(arrayBufferToString(accountbuf));
    return stringToArrayBuffer(JSON.stringify(account))
}

function getBalance_(address) {
    const bz = wasmx.bech32StringToBytes(address);
    return wasmx.getBalance(bz);
}

function keccak256_(data) {
    const databz = stringToArrayBuffer(data);
    return wasmx.keccak256(databz);
}

function instantiateAccount(codeId, initMsg, balance) {
    const msgbuf = hexStringToArrayBuffer(initMsg);
    const balancebuf = hexStringToArrayBuffer(balance);
    return wasmx.instantiateAccount(codeId, msgbuf, balancebuf);
}

function instantiateAccount2(codeId, salt, initMsg, balance) {
    return wasmx.instantiateAccount2(codeId, hexStringToArrayBuffer(salt), hexStringToArrayBuffer(initMsg), hexStringToArrayBuffer(balance));
}

function store(key_, value_) {
    const key = stringToArrayBuffer(key_);
    const value = stringToArrayBuffer(value_);
    wasmx.storageStore(key, value);
}

function load(key_) {
    const key = stringToArrayBuffer(key_);
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

// const hexStringToArrayBuffer = hexString => new Uint8Array(hexString.match(/.{1,2}/g).map(byte => parseInt(byte, 16))).buffer;

const hexStringToArrayBuffer = hexString => new Uint8Array(hexString.match(/[\da-f]{2}/gi).map(function (h) {
    return parseInt(h, 16)
})).buffer;
