package keeper_test

import (
	"context"
	_ "embed"
	"encoding/binary"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	"mythos/v1/x/wasmx/types"
	wasmxvm "mythos/v1/x/wasmx/vm"
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
	fmt.Println("--getCallData allocate--", err, result)
	if err != nil {
		return 0, err
	}
	if len(result) == 0 {
		return 0, fmt.Errorf("memory allocation failed")
	}
	ptr := result[0]

	fmt.Println("--getCallData write mem--", ptr, string(data))
	success := mem.Write(uint32(ptr), data)
	fmt.Println("--getCallData success--", success)
	return uint32(ptr), nil
}

func (suite *KeeperTestSuite) TestWazeroWasmxSimpleStorage() {
	wasmbin := wasmxSimpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)

	fmt.Println("--TestWazeroWasmxSimpleStorage--")

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	fmt.Println("--TestWazeroWasmxSimpleStorage2--")

	storageMap := map[string][]byte{}

	r := wazero.NewRuntime(appA.Context())
	fmt.Println("--TestWazeroWasmxSimpleStorage r--", r)
	defer r.Close(appA.Context())

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

		// x, y := api.DecodeI32(stack[0]), api.DecodeI32(stack[1])
		// sum := x + y
		// stack[0] = api.EncodeI32(sum)
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

		// x, y := api.DecodeI32(stack[0]), api.DecodeI32(stack[1])
		// sum := x + y
		// stack[0] = api.EncodeI32(sum)
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

	_, err := wasmxEnv.Instantiate(ctx)
	s.Require().NoError(err)

	_, err = envEnv.Instantiate(ctx)
	s.Require().NoError(err)

	mod, err := r.Instantiate(ctx, wasmbin)
	s.Require().NoError(err)

	respi, err := mod.ExportedFunction("instantiate").Call(ctx)
	s.Require().NoError(err)
	fmt.Println("---respi---", respi)

	resp, _ := mod.ExportedFunction("main").Call(ctx)
	s.Require().NoError(err)
	fmt.Println("---resp---", resp, vmCtx.FinishData, string(vmCtx.FinishData))

	calldata = []byte(`{"get":{"key":"hello"}}`)
	vmCtx.Env.CurrentCall.CallData = calldata
	resp, _ = mod.ExportedFunction("main").Call(ctx)
	s.Require().NoError(err)

	fmt.Println("---resp---", resp, vmCtx.FinishData, string(vmCtx.FinishData))

	return

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

	data := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	wasmlogs := appA.GetWasmxEvents(res.GetEvents())
	emptyDataLogs := appA.GetEventsByAttribute(wasmlogs, "data", "0x")
	topicLogs := appA.GetEventsByAttribute(wasmlogs, "topic", "0x68656c6c6f000000000000000000000000000000000000000000000000000000")
	s.Require().Equal(1, len(wasmlogs), res.GetEvents())
	s.Require().Equal(1, len(emptyDataLogs), res.GetEvents())
	s.Require().Equal(1, len(topicLogs), res.GetEvents())

	initvalue := "sammy"
	keybz := []byte("hello")
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, string(queryres))

	data = []byte(`{"get":{"key":"hello"}}`)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(string(qres), "sammy")
}
