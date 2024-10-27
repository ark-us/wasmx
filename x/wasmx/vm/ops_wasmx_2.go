package vm

import (
	"bytes"
	"encoding/json"
	"math/big"

	sdkmath "cosmossdk.io/math"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ed25519 "github.com/cometbft/cometbft/crypto/ed25519"
	merkle "github.com/cometbft/cometbft/crypto/merkle"
	"github.com/cometbft/cometbft/crypto/tmhash"

	"github.com/second-state/WasmEdge-go/wasmedge"

	mcodec "mythos/v1/codec"
	networktypes "mythos/v1/x/network/types"
	"mythos/v1/x/wasmx/types"
	mem "mythos/v1/x/wasmx/vm/memory/common"
	vmtypes "mythos/v1/x/wasmx/vm/types"
)

func sha256(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	data, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	hashbz := tmhash.Sum(data)
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, hashbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// getEnv(): ArrayBuffer
func getEnv(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	envbz, err := json.Marshal(ctx.Env)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, envbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxGetChainId(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, []byte(ctx.Env.Chain.ChainIdFull))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// address -> account
func getAccount(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	addr, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	address := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(vmtypes.CleanupAddress(addr))
	codeInfo := ctx.CosmosHandler.GetCodeInfo(address.Bytes())
	code := types.EnvContractInfo{
		Address:  address,
		CodeHash: codeInfo.CodeHash,
		Bytecode: codeInfo.InterpretedBytecodeRuntime,
	}

	codebz, err := json.Marshal(code)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, codebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func keccak256Util(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	data, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	if ctx.ContractRouter["keccak256"] == nil {
		return nil, wasmedge.Result_Fail
	}
	keccakVm := ctx.ContractRouter["keccak256"].Vm
	input_offset := int32(0)
	input_length := int32(len(data))
	output_offset := input_length
	context_offset := output_offset + int32(32)

	keccakMem := keccakVm.GetActiveModule().FindMemory("memory")
	if keccakMem == nil {
		return nil, wasmedge.Result_Fail
	}
	err = keccakMem.SetData(data, uint(input_offset), uint(input_length))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	_, err = keccakVm.Execute("keccak", context_offset, input_offset, input_length, output_offset)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	result, err := keccakMem.GetData(uint(output_offset), uint(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, result)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// TODO move this to a restricted host
// call request -> call response
func externalCall(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req vmtypes.CallRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Ctx.Logger().Debug("unmarshalling CallRequest failed", "error", err)
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	var success int32
	var returnData []byte

	to := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(req.To)
	from := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(req.From)
	commonReq := req.ToCommon(from, to)

	// Send funds
	if req.Value.BitLen() > 0 {
		err = BankSendCoin(ctx, ctx.Env.Contract.Address, to, sdk.NewCoins(sdk.NewCoin(ctx.Env.Chain.Denom, sdkmath.NewIntFromBigInt(req.Value))))
	}
	if err != nil {
		success = int32(2)
	} else {
		success, returnData = WasmxCall(ctx, commonReq)
	}

	response := vmtypes.CallResponse{
		Success: uint8(success),
		Data:    returnData,
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxCall(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req vmtypes.SimpleCallRequestRaw
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Ctx.Logger().Debug("unmarshalling CallRequest failed", "error", err)
		return nil, wasmedge.Result_Fail
	}
	// TODO have this resolver for any internal call
	// it should be a smart contract
	to, err := ctx.CosmosHandler.GetAddressOrRole(ctx.Ctx, req.To)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	var success int32
	var returnData []byte

	// Send funds
	if req.Value.BigInt().BitLen() > 0 {
		err = BankSendCoin(ctx, ctx.Env.Contract.Address, to, sdk.NewCoins(sdk.NewCoin(ctx.Env.Chain.Denom, sdkmath.NewIntFromBigInt(req.Value.BigInt()))))
	}
	if err != nil {
		success = int32(2)
	} else {
		contractContext := GetContractContext(ctx, to.Bytes())
		if contractContext == nil {
			// ! we return success here in case the contract does not exist
			success = int32(0)
		} else {
			gasLimit := req.GasLimit
			if gasLimit == nil {
				// TODO: gas remaining!!
			}
			req := vmtypes.CallRequestCommon{
				To:       to,
				From:     ctx.Env.Contract.Address,
				Value:    req.Value.BigInt(),
				GasLimit: gasLimit,
				Calldata: req.Calldata,
				Bytecode: contractContext.ContractInfo.Bytecode,
				CodeHash: contractContext.ContractInfo.CodeHash,
				IsQuery:  req.IsQuery,
			}
			success, returnData = WasmxCall(ctx, req)
			ctx.ReturnData = returnData
		}
	}

	response := vmtypes.CallResponse{
		Success: uint8(success),
		Data:    returnData,
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		ctx.Ctx.Logger().Error("wasmxCall allocate memory", "err", err.Error())
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxGetBalance(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	addr, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	address := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(vmtypes.CleanupAddress(addr))
	balance, err := BankGetBalance(ctx, address, ctx.Env.Chain.Denom)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, balance.Amount.BigInt().FillBytes(make([]byte, 32)))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxGetBlockHash(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	bz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	blockNumber := big.NewInt(0).SetBytes(bz)
	data := ctx.CosmosHandler.GetBlockHash(blockNumber.Uint64())
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxGetCurrentBlock(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	bz, err := json.Marshal(ctx.Env.Block)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, bz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxGetAddressByRole(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	rolebz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	addr, err := ctx.CosmosHandler.GetAddressOrRole(ctx.Ctx, string(rolebz))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, addr.Bytes())
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxGetRoleByAddress(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	addrbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	addr := sdk.AccAddress(addrbz)
	role := ctx.CosmosHandler.GetRoleByContractAddress(ctx.Ctx, addr)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, []byte(role))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxExecuteCosmosMsg(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	reqbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var msg cdctypes.Any
	ctx.CosmosHandler.JSONCodec().UnmarshalJSON(reqbz, &msg)

	evs, _, err := ctx.CosmosHandler.ExecuteCosmosMsgAny(&msg)
	errmsg := ""
	success := 0
	if err != nil {
		errmsg = err.Error()
		success = 1
	} else {
		ctx.Ctx.EventManager().EmitEvents(evs)
	}
	response := vmtypes.CallResponse{
		Success: uint8(success),
		Data:    []byte(errmsg),
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxDecodeCosmosTxToJson(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	reqbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	jsonbz, err := ctx.CosmosHandler.DecodeCosmosTx(reqbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, jsonbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxVerifyCosmosTx(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	reqbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	valid, err := ctx.CosmosHandler.VerifyCosmosTx(reqbz)
	resp := VerifyCosmosTxResponse{Valid: valid, Error: ""}
	if err != nil {
		resp.Error = err.Error()
	}
	respbz, err := json.Marshal(&resp)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxCreateAccountInterpreted(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)

	requestbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req vmtypes.CreateAccountInterpretedRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// TODO metadata should come from the request
	metadata := types.CodeMetadata{}
	// TODO info from provenance ?
	initMsg, err := json.Marshal(types.WasmxExecutionMessage{Data: []byte{}})
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	var sdeps []string

	addrstr := ctx.Env.Contract.Address.String()

	for _, dep := range ctx.ContractRouter[addrstr].ContractInfo.SystemDeps {
		sdeps = append(sdeps, dep.Label)
	}
	_, _, contractAddress, err := ctx.CosmosHandler.Deploy(
		req.Bytecode,
		&ctx.Env.CurrentCall.Origin,
		&ctx.Env.Contract.Address,
		initMsg,
		req.Balance,
		sdeps,
		metadata,
		"", // TODO label?
		[]byte{},
	)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}

	contractbz := mem.PaddLeftTo32(contractAddress.Bytes())
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, contractbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxCreate2AccountInterpreted(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)

	requestbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req vmtypes.Create2AccountInterpretedRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	metadata := types.CodeMetadata{}
	// TODO info from provenance ?
	initMsg, err := json.Marshal(types.WasmxExecutionMessage{Data: []byte{}})
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	var sdeps []string

	addrstr := ctx.Env.Contract.Address.String()

	for _, dep := range ctx.ContractRouter[addrstr].ContractInfo.SystemDeps {
		sdeps = append(sdeps, dep.Label)
	}

	_, _, contractAddress, err := ctx.CosmosHandler.Deploy(
		req.Bytecode,
		&ctx.Env.CurrentCall.Origin,
		&ctx.Env.Contract.Address,
		initMsg,
		req.Balance,
		sdeps,
		metadata,
		"", // TODO label?
		req.Salt.FillBytes(make([]byte, 32)),
	)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}

	contractbz := mem.PaddLeftTo32(contractAddress.Bytes())
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, contractbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxCreateAccount(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)

	requestbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req vmtypes.InstantiateAccountRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Ctx.Logger().Debug("unmarshaling InstantiateAccountRequest", "error", err)
		return nil, wasmedge.Result_Fail
	}
	// TODO info from provenance ?
	initMsg, err := json.Marshal(types.WasmxExecutionMessage{Data: req.Msg})
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	contractAddress, err := ctx.CosmosHandler.Create(
		req.CodeId,
		ctx.Env.Contract.Address,
		initMsg,
		req.Label,
		nil,
		req.Funds,
	)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}

	response := vmtypes.InstantiateAccountResponse{Address: *contractAddress}
	respbz, err := json.Marshal(response)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxCreate2Account(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)

	requestbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req vmtypes.Instantiate2AccountRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	// TODO info from provenance ?
	initMsg, err := json.Marshal(types.WasmxExecutionMessage{Data: []byte{}})
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	contractAddress, err := ctx.CosmosHandler.Create2(
		req.CodeId,
		ctx.Env.Contract.Address,
		initMsg,
		req.Salt,
		req.Label,
		nil,
		req.Funds,
	)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}

	response := vmtypes.Instantiate2AccountResponse{Address: *contractAddress}
	respbz, err := json.Marshal(response)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

type GrpcRequest struct {
	IpAddress string `json:"ip_address"`
	Contract  []byte `json:"contract"`
	Data      []byte `json:"data"` // should be []byte (base64 encoded)
}

type GrpcResponse struct {
	Data  []byte `json:"data"`
	Error string `json:"error"`
}

func wasmxGrpcRequest(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	databz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var data GrpcRequest
	err = json.Unmarshal(databz, &data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	contractAddress := sdk.AccAddress(vmtypes.CleanupAddress(data.Contract))
	contractAddressStr, err := ctx.CosmosHandler.AddressCodec().BytesToString(contractAddress)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	msg := &networktypes.MsgGrpcSendRequest{
		IpAddress: data.IpAddress,
		Contract:  contractAddressStr,
		Data:      []byte(data.Data),
		Sender:    ctx.Env.Contract.Address.String(),
	}
	evs, res, err := ctx.CosmosHandler.ExecuteCosmosMsg(msg)
	errmsg := ""
	if err != nil {
		errmsg = err.Error()
	} else {
		ctx.Ctx.EventManager().EmitEvents(evs)
	}
	rres := networktypes.MsgGrpcSendRequestResponse{Data: make([]byte, 0)}
	if res != nil {
		err = rres.Unmarshal(res)
		if err != nil {
			return nil, wasmedge.Result_Fail
		}
	}
	resp := GrpcResponse{
		Data:  rres.Data,
		Error: errmsg,
	}
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

type StartTimeoutRequest struct {
	Id       string `json:"id"`
	Contract string `json:"contract"`
	Delay    int64  `json:"delay"`
	Args     []byte `json:"args"`
}

type CancelTimeoutRequest struct {
	Id string `json:"id"`
}

type StartBackgroundProcessRequest struct {
	Contract string `json:"contract"`
	Args     []byte `json:"args"`
}

type StartBackgroundProcessResponse struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

type WriteToBackgroundProcessRequest struct {
	Contract string `json:"contract"`
	Data     []byte `json:"data"`
	PtrFunc  string `json:"ptrFunc"`
}

type WriteToBackgroundProcessResponse struct {
	Error string `json:"error"`
}

type ReadFromBackgroundProcessRequest struct {
	Contract string `json:"contract"`
	PtrFunc  string `json:"ptrFunc"`
	LenFunc  string `json:"lenFunc"`
}

type ReadFromBackgroundProcessResponse struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

// TODO move this to a restricted role
// startTimeout(req: ArrayBuffer): i32
func wasmxStartTimeout(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	reqbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req StartTimeoutRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	msgtosend := &networktypes.MsgStartTimeoutRequest{
		Id:       req.Id,
		Sender:   ctx.Env.Contract.Address.String(),
		Contract: req.Contract,
		Delay:    req.Delay,
		Args:     req.Args,
	}
	evs, res, err := ctx.CosmosHandler.ExecuteCosmosMsg(msgtosend)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error())
		return nil, wasmedge.Result_Fail
	}
	ctx.Ctx.EventManager().EmitEvents(evs)
	var resp networktypes.MsgStartTimeoutResponse
	err = resp.Unmarshal(res)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// TODO move this to a restricted role
func wasmxCancelTimeout(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	reqbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req CancelTimeoutRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	msgtosend := &networktypes.MsgCancelTimeoutRequest{
		Id:     req.Id,
		Sender: ctx.Env.Contract.Address.String(),
	}
	evs, _, err := ctx.CosmosHandler.ExecuteCosmosMsg(msgtosend)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error())
		return nil, wasmedge.Result_Fail
	}
	ctx.Ctx.EventManager().EmitEvents(evs)
	return returns, wasmedge.Result_Success
}

// startBackgroundProcess(ArrayBuffer): ArrayBuffer
func wasmxStartBackgroundProcess(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	reqbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req StartBackgroundProcessRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	msgtosend := &networktypes.MsgStartBackgroundProcessRequest{
		Sender:   ctx.Env.Contract.Address.String(), // TODO wasmx?
		Contract: req.Contract,
		Args:     req.Args,
	}
	evs, res, err := ctx.CosmosHandler.ExecuteCosmosMsg(msgtosend)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error())
		return nil, wasmedge.Result_Fail
	}
	ctx.Ctx.EventManager().EmitEvents(evs)
	var resp networktypes.MsgStartBackgroundProcessResponse
	err = resp.Unmarshal(res)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// writeToBackgroundProcess(ArrayBuffer): ArrayBuffer
func wasmxWriteToBackgroundProcess(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	reqbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req WriteToBackgroundProcessRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	resp := WriteToBackgroundProcessResponse{
		Error: "",
	}

	contractAddress, err := ctx.CosmosHandler.GetAddressOrRole(ctx.Ctx, req.Contract)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	procc, err := types.GetBackgroundProcesses(ctx.GoContextParent)
	if err != nil {
		resp.Error = err.Error()
		ptr, err := prepareResponse(ctx, callframe, &resp)
		if err != nil {
			return nil, wasmedge.Result_Fail
		}
		returns[0] = ptr
		return returns, wasmedge.Result_Success
	}
	proc, ok := procc.Processes[contractAddress.String()]
	if !ok {
		resp.Error = "process not existent"
		ptr, err := prepareResponse(ctx, callframe, &resp)
		if err != nil {
			return nil, wasmedge.Result_Fail
		}
		returns[0] = ptr
		return returns, wasmedge.Result_Success
	}

	mod := proc.ContractVm.GetActiveModule()
	ptrGlobal := mod.FindGlobal(req.PtrFunc).GetValue()
	dataptr := uint(ptrGlobal.(int32))

	activeMemory := mod.FindMemory("memory")
	err = activeMemory.SetData(req.Data, dataptr, uint(len(req.Data)))
	if err != nil {
		resp.Error = err.Error()
	}

	ptr, err := prepareResponse(ctx, callframe, &resp)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// readFromBackgroundProcess(ArrayBuffer): ArrayBuffer
func wasmxReadFromBackgroundProcess(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	reqbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req ReadFromBackgroundProcessRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	resp := ReadFromBackgroundProcessResponse{
		Data:  []byte{},
		Error: "",
	}

	contractAddress, err := ctx.CosmosHandler.GetAddressOrRole(ctx.Ctx, req.Contract)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	procc, err := types.GetBackgroundProcesses(ctx.GoContextParent)
	if err != nil {
		resp.Error = err.Error()
		ptr, err := prepareResponse(ctx, callframe, &resp)
		if err != nil {
			return nil, wasmedge.Result_Fail
		}
		returns[0] = ptr
		return returns, wasmedge.Result_Success
	}
	proc, ok := procc.Processes[contractAddress.String()]
	if !ok {
		resp.Error = "process not existent"
		ptr, err := prepareResponse(ctx, callframe, &resp)
		if err != nil {
			return nil, wasmedge.Result_Fail
		}
		returns[0] = ptr
		return returns, wasmedge.Result_Success
	}
	mod := proc.ContractVm.GetActiveModule()
	lengthGlobal := mod.FindGlobal(req.LenFunc).GetValue()
	reslen := uint(lengthGlobal.(int32))

	ptrGlobal := mod.FindGlobal(req.PtrFunc).GetValue()
	resptr := uint(ptrGlobal.(int32))

	activeMemory := mod.FindMemory("memory")
	byteArray, err := activeMemory.GetData(resptr, reslen)
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Data = byteArray
	}
	ptr, err := prepareResponse(ctx, callframe, &resp)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func prepareResponse(ctx *Context, callframe *wasmedge.CallingFrame, resp interface{}) (int32, error) {
	respbz, err := json.Marshal(&resp)
	if err != nil {
		return 0, nil
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, respbz)
	if err != nil {
		return 0, err
	}
	return ptr, nil
}

type MerkleSlices struct {
	Slices [][]byte `json:"slices"`
}

type FinalizeBlockWrap struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

func merkleHash(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	data, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var val MerkleSlices
	err = json.Unmarshal(data, &val)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	hashbz := merkle.HashFromByteSlices(val.Slices)
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, hashbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func ed25519Sign(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	privbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	msgbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	privKey := ed25519.PrivKey(privbz)
	signature, err := privKey.Sign(msgbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, signature)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func ed25519Verify(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	pubkeybz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	signaturebz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	msgbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[2])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	pubKey := ed25519.PubKey(pubkeybz)
	isSigner := pubKey.VerifySignature(msgbz, signaturebz)
	returns := make([]interface{}, 1)
	returns[0] = 0
	if isSigner {
		returns[0] = 1
	}
	return returns, wasmedge.Result_Success
}

// TODO replace this with sha256.Sum256(bz)[:20]
func ed25519PubToHex(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	pubkeybz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	pubKey := ed25519.PubKey(pubkeybz)
	hexAddr := pubKey.Address()
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, hexAddr)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// addr_canonicalize(string) -> ArrayBuffer;
func wasmxCanonicalize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addrStrBz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	addr, err := mcodec.AccAddressPrefixedFromBech32(string(addrStrBz))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, addr.Bytes())
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxAddrEquivalent(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	addr1StrBz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	addr2StrBz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	addr1, err := mcodec.AccAddressPrefixedFromBech32(string(addr1StrBz))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	addr2, err := mcodec.AccAddressPrefixedFromBech32(string(addr2StrBz))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	same := bytes.Equal(addr1.Bytes(), addr2.Bytes())
	returns := make([]interface{}, 1)
	returns[0] = 0
	if same {
		returns[0] = 1
	}
	return returns, wasmedge.Result_Success
}

// addr_humanize(ArrayBuffer) -> string;
func wasmxHumanize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addrBz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrBz))
	addrstr, err := ctx.CosmosHandler.AddressCodec().BytesToString(addr)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, []byte(addrstr))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// addr_humanize_mc(ArrayBuffer) -> string;
func wasmxHumanizeMultiChain(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addrBz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	prefixBz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrBz))
	prefix := string(prefixBz)

	addrCodec := mcodec.NewBech32Codec(prefix, mcodec.NewAddressPrefixedFromAcc)

	addrstr, err := addrCodec.BytesToString(addr)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, []byte(addrstr))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// addr_canonicalize_mc(string) -> ArrayBuffer;
func wasmxCanonicalizeMultiChain(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addrStrBz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	addr, err := mcodec.AccAddressPrefixedFromBech32(string(addrStrBz))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	respbz, err := json.Marshal(&vmtypes.PrefixedAddress{Bz: addr.Bytes(), Prefix: addr.Prefix()})
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := ctx.MemoryHandler.AllocateWriteMem(ctx.MustGetVmFromContext(), callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

type LoggerLog struct {
	Msg   string   `json:"msg"`
	Parts []string `json:"parts"`
}

func getLoggerData(ctx *Context, callframe *wasmedge.CallingFrame, params []interface{}) (string, []any, error) {
	message, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return "", nil, err
	}
	var data LoggerLog
	err = json.Unmarshal(message, &data)
	if err != nil {
		return "", nil, err
	}
	parts := make([]any, len(data.Parts))
	for i, part := range data.Parts {
		parts[i] = part
	}
	return data.Msg, parts, nil
}

func wasmxLoggerInfo(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	msg, parts, err := getLoggerData(ctx, callframe, params)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.Logger(ctx.Ctx).Info(msg, parts...)
	// if strings.Contains(msg, "start block proposal") {
	// 	panic("000")
	// }
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func wasmxLoggerError(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	msg, parts, err := getLoggerData(ctx, callframe, params)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.Logger(ctx.Ctx).Error(msg, parts...)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func wasmxLoggerDebug(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	msg, parts, err := getLoggerData(ctx, callframe, params)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.Logger(ctx.Ctx).Debug(msg, parts...)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func wasmxLoggerDebugExtended(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	msg, parts, err := getLoggerData(ctx, callframe, params)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	newmodule := GetVmLoggerExtended(ctx.Logger, ctx.Env.Chain.ChainIdFull, ctx.Env.Contract.Address.String())
	newmodule(ctx.Ctx).Debug(msg, parts...)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func wasmxEmitCosmosEvents(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	evsbz, err := ctx.MemoryHandler.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// var data []types.Event
	var data []sdk.Event
	err = json.Unmarshal(evsbz, &data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// ctx.CosmosEvents = append(ctx.CosmosEvents, data...)
	ctx.Ctx.EventManager().EmitEvents(data)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func BuildWasmxEnv2(context *Context) *wasmedge.Module {
	env := wasmedge.NewModule("wasmx")
	functype_i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype__i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)

	env.AddFunction("sha256", wasmedge.NewFunction(functype_i32_i32, sha256, context, 0))

	env.AddFunction("getCallData", wasmedge.NewFunction(functype__i32, getCallData, context, 0))
	env.AddFunction("getEnv", wasmedge.NewFunction(functype__i32, getEnv, context, 0))
	env.AddFunction("getChainId", wasmedge.NewFunction(functype__i32, wasmxGetChainId, context, 0))
	env.AddFunction("getCaller", wasmedge.NewFunction(functype__i32, wasmxGetCaller, context, 0))
	env.AddFunction("getAddress", wasmedge.NewFunction(functype__i32, wasmxGetAddress, context, 0))
	env.AddFunction("storageLoad", wasmedge.NewFunction(functype_i32_i32, wasmxStorageLoad, context, 0))
	env.AddFunction("storageStore", wasmedge.NewFunction(functype_i32i32_, wasmxStorageStore, context, 0))
	env.AddFunction("storageLoadRange", wasmedge.NewFunction(functype_i32_i32, wasmxStorageLoadRange, context, 0))
	env.AddFunction("storageLoadRangePairs", wasmedge.NewFunction(functype_i32_i32, wasmxStorageLoadRangePairs, context, 0))
	env.AddFunction("log", wasmedge.NewFunction(functype_i32_, wasmxLog, context, 0))
	env.AddFunction("emitCosmosEvents", wasmedge.NewFunction(functype_i32_, wasmxEmitCosmosEvents, context, 0))
	env.AddFunction("getReturnData", wasmedge.NewFunction(functype__i32, wasmxGetReturnData, context, 0))
	env.AddFunction("getFinishData", wasmedge.NewFunction(functype__i32, wasmxGetFinishData, context, 0))
	env.AddFunction("setFinishData", wasmedge.NewFunction(functype_i32_, wasmxSetFinishData, context, 0))
	// TODO some precompiles use setReturnData instead of setFinishData
	env.AddFunction("setReturnData", wasmedge.NewFunction(functype_i32_, wasmxSetFinishData, context, 0))
	env.AddFunction("finish", wasmedge.NewFunction(functype_i32_, wasmxFinish, context, 0))
	env.AddFunction("revert", wasmedge.NewFunction(functype_i32_, wasmxRevert, context, 0))
	env.AddFunction("getBlockHash", wasmedge.NewFunction(functype_i32_i32, wasmxGetBlockHash, context, 0))
	env.AddFunction("getCurrentBlock", wasmedge.NewFunction(functype__i32, wasmxGetCurrentBlock, context, 0))
	env.AddFunction("getAccount", wasmedge.NewFunction(functype_i32_i32, getAccount, context, 0))
	env.AddFunction("getBalance", wasmedge.NewFunction(functype_i32_i32, wasmxGetBalance, context, 0))
	env.AddFunction("call", wasmedge.NewFunction(functype_i32_i32, wasmxCall, context, 0))
	env.AddFunction("keccak256", wasmedge.NewFunction(functype_i32_i32, keccak256Util, context, 0))

	env.AddFunction("createAccountInterpreted", wasmedge.NewFunction(functype_i32_i32, wasmxCreateAccountInterpreted, context, 0))
	env.AddFunction("create2AccountInterpreted", wasmedge.NewFunction(functype_i32_i32, wasmxCreate2AccountInterpreted, context, 0))

	env.AddFunction("createAccount", wasmedge.NewFunction(functype_i32_i32, wasmxCreateAccount, context, 0))
	env.AddFunction("create2Account", wasmedge.NewFunction(functype_i32_i32, wasmxCreate2Account, context, 0))

	env.AddFunction("MerkleHash", wasmedge.NewFunction(functype_i32_i32, merkleHash, context, 0))
	// TODO
	// env.AddFunction("ProtoMarshal", wasmedge.NewFunction(functype__i32, ProtoMarshal, context, 0))
	// env.AddFunction("ProtoUnmarshal", wasmedge.NewFunction(functype__i32, ProtoUnmarshal, context, 0))

	env.AddFunction("LoggerInfo", wasmedge.NewFunction(functype_i32_, wasmxLoggerInfo, context, 0))
	env.AddFunction("LoggerError", wasmedge.NewFunction(functype_i32_, wasmxLoggerError, context, 0))
	env.AddFunction("LoggerDebug", wasmedge.NewFunction(functype_i32_, wasmxLoggerDebug, context, 0))
	env.AddFunction("LoggerDebugExtended", wasmedge.NewFunction(functype_i32_, wasmxLoggerDebugExtended, context, 0))

	env.AddFunction("ed25519Sign", wasmedge.NewFunction(functype_i32i32_i32, ed25519Sign, context, 0))
	env.AddFunction("ed25519Verify", wasmedge.NewFunction(functype_i32i32i32_i32, ed25519Verify, context, 0))
	env.AddFunction("ed25519PubToHex", wasmedge.NewFunction(functype_i32_i32, ed25519PubToHex, context, 0))

	env.AddFunction("addr_humanize", wasmedge.NewFunction(functype_i32_i32, wasmxHumanize, context, 0))
	env.AddFunction("addr_canonicalize", wasmedge.NewFunction(functype_i32_i32, wasmxCanonicalize, context, 0))
	env.AddFunction("addr_equivalent", wasmedge.NewFunction(functype_i32i32_i32, wasmxAddrEquivalent, context, 0))
	env.AddFunction("addr_humanize_mc", wasmedge.NewFunction(functype_i32i32_i32, wasmxHumanizeMultiChain, context, 0))
	env.AddFunction("addr_canonicalize_mc", wasmedge.NewFunction(functype_i32_i32, wasmxCanonicalizeMultiChain, context, 0))

	env.AddFunction("getAddressByRole", wasmedge.NewFunction(functype_i32_i32, wasmxGetAddressByRole, context, 0))
	env.AddFunction("getRoleByAddress", wasmedge.NewFunction(functype_i32_i32, wasmxGetRoleByAddress, context, 0))

	env.AddFunction("executeCosmosMsg", wasmedge.NewFunction(functype_i32_i32, wasmxExecuteCosmosMsg, context, 0))
	env.AddFunction("decodeCosmosTxToJson", wasmedge.NewFunction(functype_i32_i32, wasmxDecodeCosmosTxToJson, context, 0))
	env.AddFunction("verifyCosmosTx", wasmedge.NewFunction(functype_i32_i32, wasmxVerifyCosmosTx, context, 0))

	// TODO move externalCall, grpcRequest, startTimeout to only system API
	// move them to the network module: vmnetwork
	env.AddFunction("externalCall", wasmedge.NewFunction(functype_i32_i32, externalCall, context, 0))
	env.AddFunction("grpcRequest", wasmedge.NewFunction(functype_i32_i32, wasmxGrpcRequest, context, 0))
	env.AddFunction("startTimeout", wasmedge.NewFunction(functype_i32_, wasmxStartTimeout, context, 0))
	env.AddFunction("cancelTimeout", wasmedge.NewFunction(functype_i32_, wasmxCancelTimeout, context, 0))
	env.AddFunction("startBackgroundProcess", wasmedge.NewFunction(functype_i32_, wasmxStartBackgroundProcess, context, 0))
	env.AddFunction("writeToBackgroundProcess", wasmedge.NewFunction(functype_i32_i32, wasmxWriteToBackgroundProcess, context, 0))
	env.AddFunction("readFromBackgroundProcess", wasmedge.NewFunction(functype_i32_i32, wasmxReadFromBackgroundProcess, context, 0))

	// TODO
	// env.AddFunction("endBackgroundProcess", wasmedge.NewFunction(functype_i32_, wasmxEndBackgroundProcess, context, 0))

	return env
}
