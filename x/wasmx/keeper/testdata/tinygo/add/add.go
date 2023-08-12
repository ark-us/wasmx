package main

//go:wasm-module myadd
//export add
func add(x, y uint32) uint32 {
	return x + y
}

// main is required for the `wasi` target, even if it isn't used.
func main() {
	println("Hello TinyGo from WasmEdge!")
}

//go:wasm-module myadd
//export instantiate
func instantiate() {
	println("instantiate")
}
