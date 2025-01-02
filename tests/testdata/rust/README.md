# wasmx-rust

## prerequisites

```bash
rustup target add wasm32-unknown-unknown
```

## build

```bash
cd simple_storage
cargo build --target wasm32-unknown-unknown --release
mv ./target/wasm32-unknown-unknown/release/simple_storage.wasm ../simple_storage.wasm
```

## test

```bash
cargo test
```

## new project

```bash
cargo new project
```
