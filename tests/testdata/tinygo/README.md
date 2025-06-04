# wasi examples

## compile with tinygo

* tested with

```
tinygo version
tinygo version 0.37.0 darwin/arm64 (using go version go1.23.2 and LLVM version 19.1.2)
```

- with WASI
```sh

tinygo build -o add.wasm -no-debug -panic=trap -scheduler=none -gc=leaking -target=wasi ./add

cd simple_storage && GOWORK=off tinygo build -o ../simple_storage.wasm -no-debug -panic=trap -scheduler=none -gc=leaking -target=wasi . && cd ..

cd forward && GOWORK=off tinygo build -o ../forward.wasm -no-debug -panic=trap -scheduler=none -gc=leaking -target=wasi . && cd ..

cd emailchain && GOWORK=off tinygo build -o ../emailchain.wasm -no-debug -panic=trap -scheduler=none -gc=leaking -target=wasi . && cd ..


```

- for web
```sh
tinygo build -o add2.wasm -target wasm ./add.go
```

* optimizations

```

tinygo build -o test.wasm -no-debug -panic=trap -scheduler=none -gc=leaking examples/serial && ls -lh test.wasm

# strip utility

-no-debug
-scheduler=none
-panic=trap
-gc=leaking
-opt=s
-opt=2
```
