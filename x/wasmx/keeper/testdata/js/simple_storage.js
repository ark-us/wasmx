import * as wasmx from 'wasmx';

function fib(n)
{
    if (n <= 0)
        return 0;
    else if (n == 1)
        return 1;
    else
        return fib(n - 1) + fib(n - 2);
}

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

function store(value) {
    wasmx.storageStore("jsstore", value);
}

function load() {
    return wasmx.storageLoad("jsstore");
}
