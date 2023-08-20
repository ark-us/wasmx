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

function wrapStore(address, value) {
    let calldata = "0x60fe47b10000000000000000000000000000000000000000000000000000000000000007"
    return wasmx.call(1000000, address, 0, calldata)
}

function wrapLoad(address) {
    let calldata = "0x6d4ce63c"
    let res = wasmx.callStatic(1000000, address, calldata)
    console.log("---wrapLoad-res", typeof res, res)
    console.log("---jjjj", JSON.parse('{"success":0,"data":[115,116,114,49]}'))
    let response = JSON.parse(res)
    console.log("-wrapLoad-response", response)
    let data = Object.values(response.data)
    console.log("-wrapLoad-data", typeof data, data)
    let datastr = bin2String(data)
    console.log("-wrapLoad-datastr", datastr)
    return datastr
}

function bin2String(array) {
    var result = "";
    for (var i = 0; i < array.length; i++) {
      result += String.fromCharCode(array[i]);
    }
    return result;
}
