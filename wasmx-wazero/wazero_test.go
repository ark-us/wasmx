package runtime_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/binary"
	"fmt"
	"os"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	"github.com/loredanacirstea/wasmx/v1/x/wasmx/types"
	wasmxvm "github.com/loredanacirstea/wasmx/v1/x/wasmx/vm"
	utils "github.com/loredanacirstea/wasmx/v1/x/wasmx/vm/utils"
)

var (
	//go:embed testdata/simple_storage.wasm
	wasmxSimpleStorage []byte
)

const AS_PTR_LENGHT_OFFSET = int32(4)
const AS_ARRAY_BUFFER_TYPE = int32(1)

func ReadMemFromPtr(mem api.Memory, ptr int32) ([]byte, error) {
	lengthbz, success := mem.Read(uint32(ptr-AS_PTR_LENGHT_OFFSET), uint32(AS_PTR_LENGHT_OFFSET))
	if !success {
		return nil, fmt.Errorf("memory failed to read")
	}
	length := binary.LittleEndian.Uint32(lengthbz)
	data, success := mem.Read(uint32(ptr), uint32(length))
	if !success {
		return nil, fmt.Errorf("memory failed to read")
	}
	return data, nil
}

func AllocWriteMem(ctx context.Context, m api.Module, data []byte) (uint32, error) {
	mem := m.Memory()
	size := len(data)
	result, err := m.ExportedFunction(types.MEMORY_EXPORT_AS).Call(ctx, uint64(size), uint64(AS_ARRAY_BUFFER_TYPE))
	if err != nil {
		return 0, err
	}
	if len(result) == 0 {
		return 0, fmt.Errorf("memory allocation failed")
	}
	ptr := result[0]
	success := mem.Write(uint32(ptr), data)
	if !success {
		return 0, fmt.Errorf("memory write failed")
	}
	return uint32(ptr), nil
}

func buildWasmxEnv(ctx sdk.Context, r wazero.Runtime) error {
	storageMap := map[string][]byte{}
	wasmxEnv := r.NewHostModuleBuilder("wasmx")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		vmctx, ok := ctx.Value("vmctx").(*wasmxvm.Context)
		if !ok {
			panic("vmctx not found")
		}
		mem := m.Memory()

		size := len(vmctx.Env.CurrentCall.CallData)
		result, err := m.ExportedFunction(types.MEMORY_EXPORT_AS).Call(ctx, uint64(size), uint64(AS_ARRAY_BUFFER_TYPE))
		if err != nil {
			panic(err)
		}

		ptr := result[0]
		success := mem.Write(uint32(ptr), vmctx.Env.CurrentCall.CallData)
		if !success {
			panic("mem write failed")
		}
		stack[0] = api.EncodeI32(int32(ptr))
	}), []api.ValueType{}, []api.ValueType{api.ValueTypeI32}).Export("getCallData")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		mem := m.Memory()
		keyptr := api.DecodeI32(stack[0])
		valueptr := api.DecodeI32(stack[1])
		keybz, err := ReadMemFromPtr(mem, keyptr)
		if err != nil {
			panic(err)
		}
		valuebz, err := ReadMemFromPtr(mem, valueptr)
		if err != nil {
			panic(err)
		}
		storageMap[string(keybz)] = valuebz
	}), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).Export("storageStore")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		mem := m.Memory()
		keyptr := api.DecodeI32(stack[0])
		keybz, err := ReadMemFromPtr(mem, keyptr)
		if err != nil {
			panic(err)
		}
		ptr, err := AllocWriteMem(ctx, m, storageMap[string(keybz)])
		if err != nil {
			panic(err)
		}
		stack[0] = uint64(ptr)
	}), []api.ValueType{api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).Export("storageLoad")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		vmctx, ok := ctx.Value("vmctx").(*wasmxvm.Context)
		if !ok {
			panic("no vmctx found")
		}
		mem := m.Memory()
		ptr := api.DecodeI32(stack[0])
		valuebz, err := ReadMemFromPtr(mem, ptr)
		if err != nil {
			panic(err)
		}
		vmctx.FinishData = valuebz
	}), []api.ValueType{api.ValueTypeI32}, []api.ValueType{}).Export("finish")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		vmctx, ok := ctx.Value("vmctx").(*wasmxvm.Context)
		if !ok {
			panic("no vmctx found")
		}
		mem := m.Memory()
		size := len(vmctx.Env.CurrentCall.CallData)
		AS_ARRAY_BUFFER_TYPE := uint64(1)
		result, err := m.ExportedFunction(types.MEMORY_EXPORT_AS).Call(ctx, uint64(size), AS_ARRAY_BUFFER_TYPE)
		if err != nil {
			panic(err)
		}
		ptr := result[0]

		success := mem.Write(uint32(ptr), vmctx.Env.CurrentCall.CallData)
		if !success {
			panic("memory write failed")
		}
		panic("revert fn")
	}), []api.ValueType{api.ValueTypeI32}, []api.ValueType{}).Export("revert")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		mem := m.Memory()
		ptr := api.DecodeI32(stack[0])
		valuebz, err := ReadMemFromPtr(mem, ptr)
		if err != nil {
			panic(err)
		}
		fmt.Println("--log--", string(valuebz))
	}), []api.ValueType{api.ValueTypeI32}, []api.ValueType{}).Export("log")

	_, err := wasmxEnv.Instantiate(ctx)
	return err
}

func buildEnvEnv(ctx sdk.Context, r wazero.Runtime) error {
	envEnv := r.NewHostModuleBuilder("env")
	envEnv = envEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		panic("abort")
	}), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).Export("abort")

	_, err := envEnv.Instantiate(ctx)
	return err
}

func TestWazeroWasmxSimpleStorage(t *testing.T) {
	var err error
	wasmbin := wasmxSimpleStorage

	ctx := sdk.Context{}
	ctx = ctx.WithContext(context.Background())

	cache := wazero.NewCompilationCache()
	require.NoError(t, err)
	defer cache.Close(ctx)
	config := wazero.NewRuntimeConfigCompiler().WithCompilationCache(cache)
	r := wazero.NewRuntimeWithConfig(ctx, config)

	defer r.Close(ctx)

	vmCtx := &wasmxvm.Context{
		Env: &types.Env{
			CurrentCall: types.MessageInfo{
				CallData: []byte{},
			},
		},
	}
	ctx = ctx.WithValue("vmctx", vmCtx)

	err = buildWasmxEnv(ctx, r)
	require.NoError(t, err)
	err = buildEnvEnv(ctx, r)
	require.NoError(t, err)

	mod, err := r.Instantiate(ctx, wasmbin)
	require.NoError(t, err)

	_, err = mod.ExportedFunction("instantiate").Call(ctx)
	require.NoError(t, err)
	start := time.Now()

	calldata := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	_, err = mod.ExportedFunction("main").Call(ctx)
	require.NoError(t, err)

	calldata = []byte(`{"get":{"key":"hello"}}`)
	vmCtx.Env.CurrentCall.CallData = calldata
	_, err = mod.ExportedFunction("main").Call(ctx)
	require.NoError(t, err)

	elapsed := time.Since(start)
	fmt.Printf("Elapsed time: %s\n", elapsed)
	require.True(t, bytes.Equal(vmCtx.FinishData, []byte("sammy")))
}

func TestWazeroWasmxSimpleStorage2(t *testing.T) {
	var err error
	wasmbin := wasmxSimpleStorage

	ctx := sdk.Context{}
	ctx = ctx.WithContext(context.Background())
	cache, err := wazero.NewCompilationCacheWithDir("/Users/user/dev/blockchain/wasmx/build/simplestorage")
	require.NoError(t, err)
	defer cache.Close(ctx)
	config := wazero.NewRuntimeConfigCompiler().WithCompilationCache(cache)
	r := wazero.NewRuntimeWithConfig(ctx, config)
	defer r.Close(ctx)

	vmCtx := &wasmxvm.Context{
		Env: &types.Env{
			CurrentCall: types.MessageInfo{
				CallData: []byte{},
			},
		},
	}
	ctx = ctx.WithValue("vmctx", vmCtx)
	err = buildWasmxEnv(ctx, r)
	require.NoError(t, err)
	err = buildEnvEnv(ctx, r)
	require.NoError(t, err)

	mod, err := r.Instantiate(ctx, wasmbin)
	require.NoError(t, err)

	calldata := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	_, err = mod.ExportedFunction("instantiate").Call(ctx)
	require.NoError(t, err)

	_, err = mod.ExportedFunction("main").Call(ctx)
	require.NoError(t, err)

	calldata = []byte(`{"get":{"key":"hello"}}`)
	vmCtx.Env.CurrentCall.CallData = calldata
	_, err = mod.ExportedFunction("main").Call(ctx)
	require.NoError(t, err)
	require.True(t, bytes.Equal(vmCtx.FinishData, []byte("sammy")))
}

func TestWazeroWasmxSimpleStorage3(t *testing.T) {
	var err error
	wasmbin := wasmxSimpleStorage

	ctx := sdk.Context{}
	ctx = ctx.WithContext(context.Background())

	wcompiledPath := "/Users/user/dev/blockchain/wasmx/build/simplestorage/wazero-dev-arm64-darwin/69d4662ff5521acb600c42336368dc50cc5366a2c113061086358dfebf321688"

	wcompiledPathMe := "/Users/user/dev/blockchain/wasmx/build/simplestorage/aaa"

	config := wazero.NewRuntimeConfigCompiler()
	r := wazero.NewRuntimeWithConfig(ctx, config)
	_, reader, err := r.CompileModuleAndSerialize(ctx, wasmbin)
	require.NoError(t, err)
	err = utils.SafeWriteReader(wcompiledPathMe, reader)
	require.NoError(t, err)

	data1, err := os.ReadFile(wcompiledPath)
	require.NoError(t, err)

	data2, err := os.ReadFile(wcompiledPathMe)
	require.NoError(t, err)
	require.True(t, bytes.Equal(data1, data2))
}

func TestWazeroWasmxSimpleStorage4(t *testing.T) {
	var err error
	wasmbin := wasmxSimpleStorage

	ctx := sdk.Context{}
	ctx = ctx.WithContext(context.Background())

	wcompiledPath := "/Users/user/dev/blockchain/wasmx/build/simplestorage/wazero-dev-arm64-darwin/69d4662ff5521acb600c42336368dc50cc5366a2c113061086358dfebf321688"

	// wcompiledPathMe := "/Users/user/dev/blockchain/wasmx/build/simplestorage/aaa"
	config := wazero.NewRuntimeConfigCompiler() //.WithCompilationCache(cache)
	r := wazero.NewRuntimeWithConfig(ctx, config)
	defer r.Close(ctx)

	vmCtx := &wasmxvm.Context{
		Env: &types.Env{
			CurrentCall: types.MessageInfo{
				CallData: []byte{},
			},
		},
	}
	ctx = ctx.WithValue("vmctx", vmCtx)
	err = buildWasmxEnv(ctx, r)
	require.NoError(t, err)
	err = buildEnvEnv(ctx, r)
	require.NoError(t, err)

	content, err := os.Open(wcompiledPath)
	require.NoError(t, err)
	compiledmod, err := r.DeserializeCompiledModule(ctx, wasmbin, content)
	require.NoError(t, err)
	mod, err := r.InstantiateModule(ctx, compiledmod, wazero.NewModuleConfig())
	require.NoError(t, err)

	_, err = mod.ExportedFunction("instantiate").Call(ctx)
	require.NoError(t, err)

	calldata := []byte(`{"set":{"key":"hello","value":"sammy"}}`)

	_, err = mod.ExportedFunction("main").Call(ctx)
	require.NoError(t, err)

	calldata = []byte(`{"get":{"key":"hello"}}`)
	vmCtx.Env.CurrentCall.CallData = calldata
	_, err = mod.ExportedFunction("main").Call(ctx)
	require.NoError(t, err)
	require.True(t, bytes.Equal(vmCtx.FinishData, []byte("sammy")))
}
