# wasmx-rust


## prerequisites

```
rustup target add wasm32-unknown-unknown
```

## build

```
cargo new project
cargo build --target wasm32-unknown-unknown --release

```

## test

```
cargo test
```

valgrind
wasm-memcheck
