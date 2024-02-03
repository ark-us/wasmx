package vm

import (
	"encoding/json"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ed25519 "github.com/cometbft/cometbft/crypto/ed25519"
	merkle "github.com/cometbft/cometbft/crypto/merkle"
	"github.com/cometbft/cometbft/crypto/tmhash"

	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	vmtypes "mythos/v1/x/wasmx/vm/types"

	networktypes "mythos/v1/x/network/types"
)

func sha256(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	data, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	hashbz := tmhash.Sum(data)
	ptr, err := allocateWriteMem(ctx, callframe, hashbz)
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
	ptr, err := allocateWriteMem(ctx, callframe, envbz)
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
	addr, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	address := sdk.AccAddress(vmtypes.CleanupAddress(addr))
	codeInfo := ctx.CosmosHandler.GetCodeInfo(address)
	code := types.EnvContractInfo{
		Address:  address,
		CodeHash: codeInfo.CodeHash,
		Bytecode: codeInfo.InterpretedBytecodeRuntime,
	}

	codebz, err := json.Marshal(code)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, codebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func keccak256Util(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	data, err := readMemFromPtr(callframe, params[0])
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
	ptr, err := allocateWriteMem(ctx, callframe, result)
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
	requestbz, err := readMemFromPtr(callframe, params[0])
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

	// Send funds
	if req.Value.BitLen() > 0 {
		err = BankSendCoin(ctx, ctx.Env.Contract.Address, req.To, sdk.NewCoins(sdk.NewCoin(ctx.Env.Chain.Denom, sdkmath.NewIntFromBigInt(req.Value))))
	}
	if err != nil {
		success = int32(2)
	} else {
		success, returnData = WasmxCall(ctx, req)
	}

	response := vmtypes.CallResponse{
		Success: uint8(success),
		Data:    returnData,
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxCall(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := readMemFromPtr(callframe, params[0])
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
	if req.Value.BitLen() > 0 {
		err = BankSendCoin(ctx, ctx.Env.Contract.Address, to, sdk.NewCoins(sdk.NewCoin(ctx.Env.Chain.Denom, sdkmath.NewIntFromBigInt(req.Value))))
	}
	if err != nil {
		success = int32(2)
	} else {
		contractContext := GetContractContext(ctx, to)
		if contractContext == nil {
			// ! we return success here in case the contract does not exist
			success = int32(0)
		} else {
			gasLimit := req.GasLimit
			if gasLimit == nil {
				// TODO: gas remaining!!
			}
			req := vmtypes.CallRequest{
				To:       to,
				From:     ctx.Env.Contract.Address,
				Value:    req.Value,
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
	ptr, err := allocateWriteMem(ctx, callframe, responsebz)
	if err != nil {
		ctx.Ctx.Logger().Error("wasmxCall allocate memory", "err", err.Error())
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxGetBalance(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	addr, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	address := sdk.AccAddress(vmtypes.CleanupAddress(addr))
	balance, err := BankGetBalance(ctx, address, ctx.Env.Chain.Denom)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, balance.Amount.BigInt().FillBytes(make([]byte, 32)))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxGetBlockHash(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	bz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	blockNumber := big.NewInt(0).SetBytes(bz)
	data := ctx.CosmosHandler.GetBlockHash(blockNumber.Uint64())
	ptr, err := allocateWriteMem(ctx, callframe, data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxGetAddressByRole(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	rolebz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	addr, err := ctx.CosmosHandler.GetAddressOrRole(ctx.Ctx, string(rolebz))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, addr.Bytes())
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

	requestbz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req vmtypes.CreateAccountInterpretedRequest
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
	for _, dep := range ctx.ContractRouter[ctx.Env.Contract.Address.String()].ContractInfo.SystemDeps {
		sdeps = append(sdeps, dep.Label)
	}
	_, _, contractAddress, err := ctx.CosmosHandler.Deploy(
		req.Bytecode,
		ctx.Env.CurrentCall.Origin,
		ctx.Env.Contract.Address,
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

	contractbz := paddLeftTo32(contractAddress.Bytes())
	ptr, err := allocateWriteMem(ctx, callframe, contractbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxCreate2AccountInterpreted(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)

	requestbz, err := readMemFromPtr(callframe, params[0])
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
	for _, dep := range ctx.ContractRouter[ctx.Env.Contract.Address.String()].ContractInfo.SystemDeps {
		sdeps = append(sdeps, dep.Label)
	}

	_, _, contractAddress, err := ctx.CosmosHandler.Deploy(
		req.Bytecode,
		ctx.Env.CurrentCall.Origin,
		ctx.Env.Contract.Address,
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

	contractbz := paddLeftTo32(contractAddress.Bytes())
	ptr, err := allocateWriteMem(ctx, callframe, contractbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxCreateAccount(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)

	requestbz, err := readMemFromPtr(callframe, params[0])
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

	response := vmtypes.InstantiateAccountResponse{Address: contractAddress}
	respbz, err := json.Marshal(response)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func wasmxCreate2Account(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)

	requestbz, err := readMemFromPtr(callframe, params[0])
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

	response := vmtypes.Instantiate2AccountResponse{Address: contractAddress}
	respbz, err := json.Marshal(response)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, respbz)
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
	databz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var data GrpcRequest
	err = json.Unmarshal(databz, &data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	contractAddress := sdk.AccAddress(vmtypes.CleanupAddress(data.Contract))
	msg := &networktypes.MsgGrpcSendRequest{
		IpAddress: data.IpAddress,
		Contract:  contractAddress.String(),
		Data:      []byte(data.Data),
		Sender:    ctx.Env.Contract.Address.String(),
	}
	// TODO evs?
	_, res, err := ctx.CosmosHandler.ExecuteCosmosMsg(msg)
	errmsg := ""
	if err != nil {
		errmsg = err.Error()
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
	ptr, err := allocateWriteMem(ctx, callframe, respbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

type StartTimeoutRequest struct {
	Contract string `json:"contract"`
	Delay    int64  `json:"delay"`
	Args     []byte `json:"args"`
}

// TODO move this to a restricted role
// startTimeout(req: ArrayBuffer): i32
func wasmxStartTimeout(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	reqbz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req StartTimeoutRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	msgtosend := &networktypes.MsgStartTimeoutRequest{
		Sender:   ctx.Env.Contract.Address.String(),
		Contract: req.Contract,
		Delay:    req.Delay,
		Args:     req.Args,
	}
	_, res, err := ctx.CosmosHandler.ExecuteCosmosMsg(msgtosend)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error())
		return nil, wasmedge.Result_Fail
	}
	var resp networktypes.MsgStartTimeoutResponse
	err = resp.Unmarshal(res)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
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
	data, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var val MerkleSlices
	err = json.Unmarshal(data, &val)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	hashbz := merkle.HashFromByteSlices(val.Slices)
	ptr, err := allocateWriteMem(ctx, callframe, hashbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func ed25519Sign(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	privbz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	msgbz, err := readMemFromPtr(callframe, params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	privKey := ed25519.PrivKey(privbz)
	signature, err := privKey.Sign(msgbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, signature)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func ed25519Verify(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	pubkeybz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	signaturebz, err := readMemFromPtr(callframe, params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	msgbz, err := readMemFromPtr(callframe, params[2])
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

// addr_canonicalize(string) -> ArrayBuffer;
func wasmxCanonicalize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addrStrBz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	data, err := addr_canonicalize(addrStrBz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := allocateWriteMem(ctx, callframe, data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// addr_humanize(ArrayBuffer) -> string;
func wasmxHumanize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addrBz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrBz))
	ptr, err := allocateWriteMem(ctx, callframe, []byte(addr.String()))
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

func getLoggerData(callframe *wasmedge.CallingFrame, params []interface{}) (string, []any, error) {
	message, err := readMemFromPtr(callframe, params[0])
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
	msg, parts, err := getLoggerData(callframe, params)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.GetContext().Logger().Info(msg, parts...)
	// if strings.Contains(msg, "start block proposal") {
	// 	panic("000")
	// }
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func wasmxLoggerError(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	msg, parts, err := getLoggerData(callframe, params)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.GetContext().Logger().Error(msg, parts...)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func wasmxLoggerDebug(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	msg, parts, err := getLoggerData(callframe, params)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.GetContext().Logger().Info(msg, parts...)
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
	env.AddFunction("getCaller", wasmedge.NewFunction(functype__i32, wasmxGetCaller, context, 0))
	env.AddFunction("getAddress", wasmedge.NewFunction(functype__i32, wasmxGetAddress, context, 0))
	env.AddFunction("storageLoad", wasmedge.NewFunction(functype_i32_i32, wasmxStorageLoad, context, 0))
	env.AddFunction("storageStore", wasmedge.NewFunction(functype_i32i32_, wasmxStorageStore, context, 0))
	env.AddFunction("log", wasmedge.NewFunction(functype_i32_, wasmxLog, context, 0))
	env.AddFunction("getReturnData", wasmedge.NewFunction(functype__i32, wasmxGetReturnData, context, 0))
	env.AddFunction("getFinishData", wasmedge.NewFunction(functype__i32, wasmxGetFinishData, context, 0))
	env.AddFunction("setFinishData", wasmedge.NewFunction(functype_i32_, wasmxSetFinishData, context, 0))
	// TODO some precompiles use setReturnData instead of setFinishData
	env.AddFunction("setReturnData", wasmedge.NewFunction(functype_i32_, wasmxSetFinishData, context, 0))
	env.AddFunction("finish", wasmedge.NewFunction(functype_i32_, wasmxFinish, context, 0))
	env.AddFunction("revert", wasmedge.NewFunction(functype_i32_, wasmxRevert, context, 0))
	env.AddFunction("getBlockHash", wasmedge.NewFunction(functype_i32_i32, wasmxGetBlockHash, context, 0))
	env.AddFunction("getAccount", wasmedge.NewFunction(functype_i32_i32, getAccount, context, 0))
	env.AddFunction("getBalance", wasmedge.NewFunction(functype_i32_i32, wasmxGetBalance, context, 0))
	env.AddFunction("call", wasmedge.NewFunction(functype_i32_i32, wasmxCall, context, 0))
	env.AddFunction("keccak256", wasmedge.NewFunction(functype_i32_i32, keccak256Util, context, 0))

	env.AddFunction("createAccountInterpreted", wasmedge.NewFunction(functype_i32_i32, wasmxCreateAccountInterpreted, context, 0))
	env.AddFunction("create2AccountInterpreted", wasmedge.NewFunction(functype_i32_i32, wasmxCreate2AccountInterpreted, context, 0))

	env.AddFunction("createAccount", wasmedge.NewFunction(functype_i32_i32, wasmxCreateAccount, context, 0))
	env.AddFunction("create2Account", wasmedge.NewFunction(functype_i32_i32, wasmxCreate2Account, context, 0))

	env.AddFunction("MerkleHash", wasmedge.NewFunction(functype_i32_i32, merkleHash, context, 0))

	env.AddFunction("LoggerInfo", wasmedge.NewFunction(functype_i32_, wasmxLoggerInfo, context, 0))
	env.AddFunction("LoggerError", wasmedge.NewFunction(functype_i32_, wasmxLoggerError, context, 0))
	env.AddFunction("LoggerDebug", wasmedge.NewFunction(functype_i32_, wasmxLoggerDebug, context, 0))

	env.AddFunction("ed25519Sign", wasmedge.NewFunction(functype_i32i32_i32, ed25519Sign, context, 0))
	env.AddFunction("ed25519Verify", wasmedge.NewFunction(functype_i32i32i32_i32, ed25519Verify, context, 0))

	env.AddFunction("addr_humanize", wasmedge.NewFunction(functype_i32_i32, wasmxHumanize, context, 0))
	env.AddFunction("addr_canonicalize", wasmedge.NewFunction(functype_i32_i32, wasmxCanonicalize, context, 0))

	env.AddFunction("getAddressByRole", wasmedge.NewFunction(functype_i32_i32, wasmxGetAddressByRole, context, 0))

	// TODO move externalCall, grpcRequest, startTimeout to only system API
	env.AddFunction("externalCall", wasmedge.NewFunction(functype_i32_i32, externalCall, context, 0))
	env.AddFunction("grpcRequest", wasmedge.NewFunction(functype_i32_i32, wasmxGrpcRequest, context, 0))
	env.AddFunction("startTimeout", wasmedge.NewFunction(functype_i32_, wasmxStartTimeout, context, 0))

	return env
}
