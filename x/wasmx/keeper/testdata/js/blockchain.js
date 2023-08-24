import * as wasmx from 'wasmx';

export function instantiate(dataObj) {
    return store(...dataObj);
}

export function main(dataObj) {
    // console.log("-***--main-dataObj-", dataObj)
    console.log("-***--main-dataObj-");
    console.log("-***--main-dataObj-keys-", Object.keys(dataObj));
    if (dataObj.getEnv) return getEnv();
    if (dataObj.getCallData) return getCallData();
    if (dataObj.store) return store(...dataObj.store);
    if (dataObj.load) return load();
    if (dataObj.getBlockHash) return getBlockHash(...dataObj.getBlockHash);
    if (dataObj.getAccount) return getAccount(...dataObj.getAccount);
    if (dataObj.getBalance) return getBalance(...dataObj.getBalance);
    if (dataObj.keccak256) return keccak256(...dataObj.keccak256);
    if (dataObj.createAccount) return createAccount(...dataObj.createAccount);
    if (dataObj.createAccount2) return createAccount2(...dataObj.createAccount2);
    throw new Error("no valid function");
}

function getEnv() {
    console.log("-*****-getEnv");
    const envbuf = wasmx.getEnv();
    console.log("--envbuf", envbuf);
    const env = JSON.parse(arrayBufferToString(envbuf));
    console.log("--env", env);
    return stringToArrayBuffer(JSON.parse(env));
}

function getCallData() {
    const calldbuf = wasmx.getCallData();
    console.log("--calldbuf", calldbuf);
    const data = JSON.parse(arrayBufferToString(envbuf));
    console.log("--data", data);
    return stringToArrayBuffer(JSON.parse(data));
}

function getBlockHash(address) {

}

function getAccount(address) {

}

function getBalance(address) {
    const bz = wasmx.bech32StringToBytes(address);
    const balance = wasmx.getBalance(bz);
    return balance;
}

function keccak256(data) {

}

function createAccount() {

}

function createAccount2() {

}

function store(key_, value_) {
    console.log("-store-dataObj-key_-", key_)
    console.log("-store-dataObj-value_-", value_)
    const key = new Uint8Array(key_);
    const value = new Uint8Array(value_);
    console.log("-store-dataObj-key-", key)
    console.log("-store-dataObj-value-", value)
    wasmx.storageStore(key.buffer, value.buffer);
}

function load(key_) {
    const key = new Uint8Array(key_);
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

const hexStringToArrayBuffer = hexString => new Uint8Array(hexString.match(/.{1,2}/g).map(byte => parseInt(byte, 16))).buffer;
