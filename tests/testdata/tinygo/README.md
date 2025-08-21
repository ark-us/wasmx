# wasi examples

## compile with tinygo

* tested with

```
tinygo version
tinygo version 0.37.0 darwin/arm64 (using go version go1.23.2 and LLVM version 19.1.2)
```

- with WASI
```sh

cd wasmx-gov && GOWORK=off tinygo build -o ../../../../wasmx/x/wasmx/vm/precompiles/35.gov_0.0.1.wasm -no-debug -scheduler=none -gc=leaking -target=wasi . && cd ..

cd wasmx-gov-continuous && GOWORK=off tinygo build -o ../../../../wasmx/x/wasmx/vm/precompiles/37.gov_cont_0.0.1.wasm -no-debug -scheduler=none -gc=leaking -target=wasi . && cd ..

tinygo build -o add.wasm -no-debug -scheduler=none -gc=leaking -target=wasi ./add

cd simple_storage && GOWORK=off tinygo build -o ../simple_storage.wasm -no-debug -scheduler=none -gc=leaking -target=wasi . && cd ..

cd forward && GOWORK=off tinygo build -o ../forward.wasm -no-debug -scheduler=none -gc=leaking -target=wasi . && cd ..

cd imaptest && GOWORK=off tinygo build -o ../imaptest.wasm -no-debug -scheduler=none -gc=leaking -target=wasi . && cd ..

cd smtptest && GOWORK=off tinygo build -o ../smtptest.wasm -no-debug -scheduler=none -gc=leaking -target=wasi . && cd ..

cd emailchain && GOWORK=off tinygo build -o ../emailchain.wasm -no-debug -scheduler=none -gc=leaking -target=wasi . && cd ..

cd mailsrv && GOWORK=off tinygo build -o ../mailsrv.wasm -no-debug -scheduler=none -gc=leaking -target=wasi . && cd ..

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
-panic=trap # any unimplemented feature or panic results in unreachable without error
-gc=leaking
-opt=s
-opt=2
```
