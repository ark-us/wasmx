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

function instantiate(dataObj) {
    console.log("--instantiate--", dataObj);
    return store(dataObj);
}

function main(dataObj) {
    // console.log("Hello World mainnnn", dataObj);
    console.log("dataObj", dataObj["store"]);
    // console.log("-------**", "fib(10)=" + fib(10));
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
