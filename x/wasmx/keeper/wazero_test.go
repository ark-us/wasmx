package keeper_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/binary"
	"fmt"
	"os"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	"mythos/v1/x/wasmx/types"
	wasmxvm "mythos/v1/x/wasmx/vm"
	utils "mythos/v1/x/wasmx/vm/utils"
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

func (suite *KeeperTestSuite) TestWazeroWasmxSimpleStorage() {
	var err error
	wasmbin := wasmxSimpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	storageMap := map[string][]byte{}
	cache := wazero.NewCompilationCache()
	s.Require().NoError(err)
	defer cache.Close(appA.Context())
	config := wazero.NewRuntimeConfigCompiler().WithCompilationCache(cache)
	r := wazero.NewRuntimeWithConfig(appA.Context(), config)

	defer r.Close(appA.Context())

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

	envEnv := r.NewHostModuleBuilder("env")
	envEnv = envEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		panic("abort")
	}), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).Export("abort")

	ctx := appA.Context()

	calldata := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	vmCtx := &wasmxvm.Context{
		Env: &types.Env{
			CurrentCall: types.MessageInfo{
				CallData: calldata,
			},
		},
	}

	ctx = ctx.WithValue("vmctx", vmCtx)

	_, err = wasmxEnv.Instantiate(ctx)
	s.Require().NoError(err)

	_, err = envEnv.Instantiate(ctx)
	s.Require().NoError(err)

	mod, err := r.Instantiate(ctx, wasmbin)
	s.Require().NoError(err)

	_, err = mod.ExportedFunction("instantiate").Call(ctx)
	s.Require().NoError(err)
	start := time.Now()

	resp, err := mod.ExportedFunction("main").Call(ctx)
	s.Require().NoError(err)
	fmt.Println("---resp---", resp, vmCtx.FinishData, string(vmCtx.FinishData))

	calldata = []byte(`{"get":{"key":"hello"}}`)
	vmCtx.Env.CurrentCall.CallData = calldata
	resp, err = mod.ExportedFunction("main").Call(ctx)
	s.Require().NoError(err)

	elapsed := time.Since(start)
	fmt.Printf("Elapsed time: %s\n", elapsed)

	fmt.Println("---resp---", resp, vmCtx.FinishData, string(vmCtx.FinishData))

	return
}

func TestWazeroWasmxSimpleStorage2(t *testing.T) {
	fmt.Println("--TestWazeroWasmxSimpleStorage2 r--")
	var err error
	wasmbin := wasmxSimpleStorage

	ctx := sdk.Context{}
	ctx = ctx.WithContext(context.Background())

	storageMap := map[string][]byte{}
	// cache := wazero.NewCompilationCache()
	cache, err := wazero.NewCompilationCacheWithDir("/Users/user/dev/blockchain/wasmx/build/simplestorage")
	require.NoError(t, err)
	defer cache.Close(ctx)
	// config := wazero.NewRuntimeConfigInterpreter().WithCompilationCache(cache)
	config := wazero.NewRuntimeConfigCompiler().WithCompilationCache(cache)

	// r := wazero.NewRuntime(ctx)
	r := wazero.NewRuntimeWithConfig(ctx, config)
	// r2 := wazero.NewRuntimeWithConfig(ctx, config)

	fmt.Println("--TestWazeroWasmxSimpleStorage r--", r)
	defer r.Close(ctx)

	wasmxEnv := r.NewHostModuleBuilder("wasmx")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		fmt.Println("--getCallData--")
		vmctx, ok := ctx.Value("vmctx").(*wasmxvm.Context)
		fmt.Println("--getCallData vmctx--", ok)
		mem := m.Memory()

		size := len(vmctx.Env.CurrentCall.CallData)
		result, err := m.ExportedFunction(types.MEMORY_EXPORT_AS).Call(ctx, uint64(size), uint64(AS_ARRAY_BUFFER_TYPE))
		fmt.Println("--getCallData allocate--", err, result)

		ptr := result[0]

		fmt.Println("--getCallData write mem--", ptr, string(vmctx.Env.CurrentCall.CallData))
		success := mem.Write(uint32(ptr), vmctx.Env.CurrentCall.CallData)
		fmt.Println("--getCallData success--", success)

		stack[0] = api.EncodeI32(int32(ptr))

		// x, y := api.DecodeI32(stack[0]), api.DecodeI32(stack[1])
		// sum := x + y
		// stack[0] = api.EncodeI32(sum)
	}), []api.ValueType{}, []api.ValueType{api.ValueTypeI32}).Export("getCallData")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		fmt.Println("--storageStore--")
		mem := m.Memory()
		keyptr := api.DecodeI32(stack[0])
		valueptr := api.DecodeI32(stack[1])
		keybz, err := ReadMemFromPtr(mem, keyptr)
		fmt.Println("--storageStore key, value--", err, string(keybz))
		valuebz, err := ReadMemFromPtr(mem, valueptr)
		fmt.Println("--storageStore key, value--", err, string(valuebz))
		storageMap[string(keybz)] = valuebz
	}), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).Export("storageStore")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		fmt.Println("--storageLoad--")
		mem := m.Memory()
		keyptr := api.DecodeI32(stack[0])
		keybz, err := ReadMemFromPtr(mem, keyptr)
		fmt.Println("--storageLoad key--", err, string(keybz))
		ptr, err := AllocWriteMem(ctx, m, storageMap[string(keybz)])
		fmt.Println("--storageLoad value ptr--", err, ptr)
		stack[0] = uint64(ptr)
	}), []api.ValueType{api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).Export("storageLoad")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		fmt.Println("--finish ctx--")
		vmctx, ok := ctx.Value("vmctx").(*wasmxvm.Context)
		fmt.Println("--finish vmctx--", ok)
		mem := m.Memory()
		ptr := api.DecodeI32(stack[0])
		valuebz, err := ReadMemFromPtr(mem, ptr)
		fmt.Println("--finish value--", err, string(valuebz))
		vmctx.FinishData = valuebz
	}), []api.ValueType{api.ValueTypeI32}, []api.ValueType{}).Export("finish")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		fmt.Println("--revert ctx--")
		vmctx, ok := ctx.Value("vmctx").(*wasmxvm.Context)
		fmt.Println("--revert vmctx--", ok)
		mem := m.Memory()

		size := len(vmctx.Env.CurrentCall.CallData)
		AS_ARRAY_BUFFER_TYPE := uint64(1)
		result, err := m.ExportedFunction(types.MEMORY_EXPORT_AS).Call(ctx, uint64(size), AS_ARRAY_BUFFER_TYPE)
		fmt.Println("--revert allocate--", err, result)
		ptr := result[0]

		success := mem.Write(uint32(ptr), vmctx.Env.CurrentCall.CallData)
		fmt.Println("--revert success--", success)

		fmt.Println("--pre revert--")
		panic("revert fn")
		fmt.Println("--post revert--")
	}), []api.ValueType{api.ValueTypeI32}, []api.ValueType{}).Export("revert")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		fmt.Println("--log--")
		// vmctx, ok := ctx.Value("vmctx").(*wasmxvm.Context)
		// fmt.Println("--log vmctx--", ok)
		mem := m.Memory()
		ptr := api.DecodeI32(stack[0])
		valuebz, err := ReadMemFromPtr(mem, ptr)
		fmt.Println("--log value--", err, string(valuebz))
	}), []api.ValueType{api.ValueTypeI32}, []api.ValueType{}).Export("log")

	envEnv := r.NewHostModuleBuilder("env")
	envEnv = envEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		fmt.Println("--abort ctx--")
	}), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).Export("abort")

	calldata := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	vmCtx := &wasmxvm.Context{
		Env: &types.Env{
			CurrentCall: types.MessageInfo{
				CallData: calldata,
			},
		},
	}

	ctx = ctx.WithValue("vmctx", vmCtx)

	_, err = wasmxEnv.Instantiate(ctx)
	require.NoError(t, err)

	_, err = envEnv.Instantiate(ctx)
	require.NoError(t, err)

	mod, err := r.Instantiate(ctx, wasmbin)
	require.NoError(t, err)

	respi, err := mod.ExportedFunction("instantiate").Call(ctx)
	require.NoError(t, err)
	fmt.Println("---respi---", respi)

	start := time.Now()

	resp, err := mod.ExportedFunction("main").Call(ctx)
	require.NoError(t, err)
	fmt.Println("---resp---", resp, vmCtx.FinishData, string(vmCtx.FinishData))

	calldata = []byte(`{"get":{"key":"hello"}}`)
	vmCtx.Env.CurrentCall.CallData = calldata
	resp, err = mod.ExportedFunction("main").Call(ctx)
	require.NoError(t, err)

	elapsed := time.Since(start)
	fmt.Printf("Elapsed time: %s\n", elapsed)

	fmt.Println("---resp---", resp, vmCtx.FinishData, string(vmCtx.FinishData))
}

func TestWazeroWasmxSimpleStorage3(t *testing.T) {
	fmt.Println("--TestWazeroWasmxSimpleStorage2 r--")
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
	fmt.Println("--TestWazeroWasmxSimpleStorage4--")
	var err error
	wasmbin := wasmxSimpleStorage

	ctx := sdk.Context{}
	ctx = ctx.WithContext(context.Background())

	wcompiledPath := "/Users/user/dev/blockchain/wasmx/build/simplestorage/wazero-dev-arm64-darwin/69d4662ff5521acb600c42336368dc50cc5366a2c113061086358dfebf321688"

	// wcompiledPathMe := "/Users/user/dev/blockchain/wasmx/build/simplestorage/aaa"

	storageMap := map[string][]byte{}
	// cache := wazero.NewCompilationCache()
	// cache, err := wazero.NewCompilationCacheWithDir("/Users/user/dev/blockchain/wasmx/build/simplestorage")
	// require.NoError(t, err)
	// defer cache.Close(ctx)
	// config := wazero.NewRuntimeConfigInterpreter().WithCompilationCache(cache)
	config := wazero.NewRuntimeConfigCompiler() //.WithCompilationCache(cache)

	// r := wazero.NewRuntime(ctx)
	r := wazero.NewRuntimeWithConfig(ctx, config)
	// r2 := wazero.NewRuntimeWithConfig(ctx, config)

	fmt.Println("--TestWazeroWasmxSimpleStorage r--", r)
	defer r.Close(ctx)

	wasmxEnv := r.NewHostModuleBuilder("wasmx")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		fmt.Println("--getCallData--")
		vmctx, ok := ctx.Value("vmctx").(*wasmxvm.Context)
		fmt.Println("--getCallData vmctx--", ok)
		mem := m.Memory()

		size := len(vmctx.Env.CurrentCall.CallData)
		result, err := m.ExportedFunction(types.MEMORY_EXPORT_AS).Call(ctx, uint64(size), uint64(AS_ARRAY_BUFFER_TYPE))
		fmt.Println("--getCallData allocate--", err, result)

		ptr := result[0]

		fmt.Println("--getCallData write mem--", ptr, string(vmctx.Env.CurrentCall.CallData))
		success := mem.Write(uint32(ptr), vmctx.Env.CurrentCall.CallData)
		fmt.Println("--getCallData success--", success)

		stack[0] = api.EncodeI32(int32(ptr))

		// x, y := api.DecodeI32(stack[0]), api.DecodeI32(stack[1])
		// sum := x + y
		// stack[0] = api.EncodeI32(sum)
	}), []api.ValueType{}, []api.ValueType{api.ValueTypeI32}).Export("getCallData")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		fmt.Println("--storageStore--")
		mem := m.Memory()
		keyptr := api.DecodeI32(stack[0])
		valueptr := api.DecodeI32(stack[1])
		keybz, err := ReadMemFromPtr(mem, keyptr)
		fmt.Println("--storageStore key, value--", err, string(keybz))
		valuebz, err := ReadMemFromPtr(mem, valueptr)
		fmt.Println("--storageStore key, value--", err, string(valuebz))
		storageMap[string(keybz)] = valuebz
	}), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).Export("storageStore")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		fmt.Println("--storageLoad--")
		mem := m.Memory()
		keyptr := api.DecodeI32(stack[0])
		keybz, err := ReadMemFromPtr(mem, keyptr)
		fmt.Println("--storageLoad key--", err, string(keybz))
		ptr, err := AllocWriteMem(ctx, m, storageMap[string(keybz)])
		fmt.Println("--storageLoad value ptr--", err, ptr)
		stack[0] = uint64(ptr)
	}), []api.ValueType{api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).Export("storageLoad")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		fmt.Println("--finish ctx--")
		vmctx, ok := ctx.Value("vmctx").(*wasmxvm.Context)
		fmt.Println("--finish vmctx--", ok)
		mem := m.Memory()
		ptr := api.DecodeI32(stack[0])
		valuebz, err := ReadMemFromPtr(mem, ptr)
		fmt.Println("--finish value--", err, string(valuebz))
		vmctx.FinishData = valuebz
	}), []api.ValueType{api.ValueTypeI32}, []api.ValueType{}).Export("finish")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		fmt.Println("--revert ctx--")
		vmctx, ok := ctx.Value("vmctx").(*wasmxvm.Context)
		fmt.Println("--revert vmctx--", ok)
		mem := m.Memory()

		size := len(vmctx.Env.CurrentCall.CallData)
		AS_ARRAY_BUFFER_TYPE := uint64(1)
		result, err := m.ExportedFunction(types.MEMORY_EXPORT_AS).Call(ctx, uint64(size), AS_ARRAY_BUFFER_TYPE)
		fmt.Println("--revert allocate--", err, result)
		ptr := result[0]

		success := mem.Write(uint32(ptr), vmctx.Env.CurrentCall.CallData)
		fmt.Println("--revert success--", success)

		fmt.Println("--pre revert--")
		panic("revert fn")
		fmt.Println("--post revert--")
	}), []api.ValueType{api.ValueTypeI32}, []api.ValueType{}).Export("revert")

	wasmxEnv = wasmxEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		fmt.Println("--log--")
		// vmctx, ok := ctx.Value("vmctx").(*wasmxvm.Context)
		// fmt.Println("--log vmctx--", ok)
		mem := m.Memory()
		ptr := api.DecodeI32(stack[0])
		valuebz, err := ReadMemFromPtr(mem, ptr)
		fmt.Println("--log value--", err, string(valuebz))
	}), []api.ValueType{api.ValueTypeI32}, []api.ValueType{}).Export("log")

	envEnv := r.NewHostModuleBuilder("env")
	envEnv = envEnv.NewFunctionBuilder().WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
		fmt.Println("--abort ctx--")
	}), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).Export("abort")

	calldata := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	vmCtx := &wasmxvm.Context{
		Env: &types.Env{
			CurrentCall: types.MessageInfo{
				CallData: calldata,
			},
		},
	}

	ctx = ctx.WithValue("vmctx", vmCtx)

	_, err = wasmxEnv.Instantiate(ctx)
	require.NoError(t, err)

	_, err = envEnv.Instantiate(ctx)
	require.NoError(t, err)

	// mod, err := r.Instantiate(ctx, wasmbin)
	content, err := os.Open(wcompiledPath)
	require.NoError(t, err)
	compiledmod, err := r.DeserializeCompiledModule(ctx, wasmbin, content)
	require.NoError(t, err)
	mod, err := r.InstantiateModule(ctx, compiledmod, wazero.NewModuleConfig())
	require.NoError(t, err)

	respi, err := mod.ExportedFunction("instantiate").Call(ctx)
	require.NoError(t, err)
	fmt.Println("---respi---", respi)

	start := time.Now()

	resp, err := mod.ExportedFunction("main").Call(ctx)
	require.NoError(t, err)
	fmt.Println("---resp---", resp, vmCtx.FinishData, string(vmCtx.FinishData))

	calldata = []byte(`{"get":{"key":"hello"}}`)
	vmCtx.Env.CurrentCall.CallData = calldata
	resp, err = mod.ExportedFunction("main").Call(ctx)
	require.NoError(t, err)

	elapsed := time.Since(start)
	fmt.Printf("Elapsed time: %s\n", elapsed)

	fmt.Println("---resp---", resp, vmCtx.FinishData, string(vmCtx.FinishData))
}
