# JS smart contracts

Use https://github.com/bytecodealliance/javy to compile JS contracts.

```sh
javy compile simple_storage.js -o simple_storage.wasm
javy compile simple_storage.js --wit simple_storage.wit -n index-world -o simple_storage.wasm
```
