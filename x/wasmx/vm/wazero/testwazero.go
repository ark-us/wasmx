package wazero

// import (
// 	"github.com/tetratelabs/wazero"
// )

// func TestWazero() {
// 	r := wazero.NewRuntime(ctx)
// 	defer r.Close(ctx)
// 	mod, _ := r.Instantiate(ctx, wasmAdd)
// 	res, _ := mod.ExportedFunction("add").Call(ctx, 1, 2)

// 	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfigInterpreter())

// 	Runtime.CompileModule

// 	wazero.NewRuntimeConfigInterpreter()
// 	wazero.NewRuntimeConfigCompiler()

// 	r := wazero.NewRuntime(ctx)
// 	defer r.Close(ctx) // This closes everything this Runtime created.

// 	// Instantiate WASI, which implements host functions needed for TinyGo to
// 	// implement `panic`.
// 	wasi_snapshot_preview1.MustInstantiate(ctx, r)

// 	// Instantiate the guest Wasm into the same runtime. It exports the `add`
// 	// function, implemented in WebAssembly.
// 	mod, err := r.Instantiate(ctx, addWasm)
// 	if err != nil {
// 		log.Panicf("failed to instantiate module: %v", err)
// 	}

// 	// Read two args to add.
// 	x, y, err := readTwoArgs(flag.Arg(0), flag.Arg(1))
// 	if err != nil {
// 		log.Panicf("failed to read arguments: %v", err)
// 	}

// 	// Call the `add` function and print the results to the console.
// 	add := mod.ExportedFunction("add")
// 	results, err := add.Call(ctx, x, y)
// 	if err != nil {
// 		log.Panicf("failed to call add: %v", err)
// 	}

// }
