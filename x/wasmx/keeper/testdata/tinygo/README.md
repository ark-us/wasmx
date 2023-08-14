# wasi examples

## compile with tinygo

- with WASI
```sh
tinygo build -o add.wasm -target=wasi ./add

cd simple_storage && tinygo build -o ../simple_storage.wasm -target=wasi . && cd ..
```

- for web
```sh
tinygo build -o add2.wasm -target wasm ./add.go
```
