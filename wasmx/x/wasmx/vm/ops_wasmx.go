package vm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ed25519 "github.com/cometbft/cometbft/crypto/ed25519"
	merkle "github.com/cometbft/cometbft/crypto/merkle"
	"github.com/cometbft/cometbft/crypto/tmhash"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm/types"
)

func sha256(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	data, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	hashbz := tmhash.Sum(data)
	return rnh.AllocateWriteMem(hashbz)
}

// getEnv(): ArrayBuffer
func getEnv(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	envbz, err := json.Marshal(ctx.Env)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(envbz)
}

func wasmxGetChainId(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	return rnh.AllocateWriteMem([]byte(ctx.Env.Chain.ChainIdFull))
}

// address -> account
func getAccount(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	addr, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	address := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(vmtypes.CleanupAddress(addr))
	_, codeInfo, _, err := ctx.CosmosHandler.GetContractInstance(address)
	if err != nil {
		return nil, err
	}
	code := types.EnvContractInfo{
		Address:  address,
		CodeHash: codeInfo.CodeHash,
		Bytecode: codeInfo.InterpretedBytecodeRuntime,
	}

	codebz, err := json.Marshal(code)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(codebz)
}

func keccak256Util(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	data, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	if ctx.ContractRouter["keccak256"] == nil {
		return nil, fmt.Errorf("missing keccak256 wasm")
	}
	keccakRnh := ctx.ContractRouter["keccak256"].RuntimeHandler
	input_offset := int32(0)
	input_length := int32(len(data))
	output_offset := input_length
	context_offset := output_offset + int32(32)

	keccakVm := keccakRnh.GetVm()
	keccakMem, err := keccakVm.GetMemory()
	if err != nil {
		return nil, err
	}
	err = keccakMem.Write(input_offset, data)
	if err != nil {
		return nil, err
	}

	_, err = keccakVm.Call("keccak", []interface{}{context_offset, input_offset, input_length, output_offset}, ctx.GasMeter)
	if err != nil {
		return nil, err
	}
	result, err := keccakMem.Read(output_offset, 32)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(result)
}

func wasmxCall(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req vmtypes.SimpleCallRequestRaw
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Ctx.Logger().Debug("unmarshalling CallRequest failed", "error", err)
		return nil, err
	}
	// TODO have this resolver for any internal call
	// it should be a smart contract
	to, err := ctx.CosmosHandler.GetAddressOrRole(ctx.Ctx, req.To)
	if err != nil {
		return nil, err
	}

	var success int32
	var returnData []byte

	// Send funds
	if req.Value.BigInt().BitLen() > 0 {
		err = BankSendCoin(ctx, ctx.Env.Contract.Address, to, sdk.NewCoins(sdk.NewCoin(ctx.Env.Chain.Denom, sdkmath.NewIntFromBigInt(req.Value.BigInt()))))
	}
	if err != nil {
		success = int32(2)
	} else {
		contractInfo := GetContractDependency(ctx, to)
		if contractInfo == nil {
			// ! we return success here in case the contract does not exist
			success = int32(0)
		} else {
			gasLimit := req.GasLimit
			if gasLimit == nil {
				// TODO: gas remaining!!
			}
			req := vmtypes.CallRequestCommon{
				To:           to,
				From:         ctx.Env.Contract.Address,
				Value:        req.Value.BigInt(),
				GasLimit:     gasLimit,
				Calldata:     req.Calldata,
				Bytecode:     contractInfo.Bytecode,
				CodeHash:     contractInfo.CodeHash,
				CodeFilePath: contractInfo.CodeFilePath,
				AotFilePath:  contractInfo.AotFilePath,
				IsQuery:      req.IsQuery,
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
		return nil, err
	}
	return rnh.AllocateWriteMem(responsebz)
}

func wasmxGetBalance(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	addr, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	address := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(vmtypes.CleanupAddress(addr))
	balance, err := BankGetBalance(ctx, address, ctx.Env.Chain.Denom)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(balance.Amount.BigInt().FillBytes(make([]byte, 32)))
}

func wasmxGetBlockHash(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	bz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	blockNumber := big.NewInt(0).SetBytes(bz)
	data := ctx.CosmosHandler.GetBlockHash(blockNumber.Uint64())
	return rnh.AllocateWriteMem(data)
}

func wasmxGetCurrentBlock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	bz, err := json.Marshal(ctx.Env.Block)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(bz)
}

func wasmxGetAddressByRole(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	rolebz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	addr, err := ctx.CosmosHandler.GetAddressOrRole(ctx.Ctx, string(rolebz))
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(addr.Bytes())
}

func wasmxGetRoleByAddress(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	addrbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	addr := sdk.AccAddress(addrbz)
	contractAddr := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(addr)
	role := ctx.CosmosHandler.GetRoleByContractAddress(ctx.Ctx, contractAddr)
	return rnh.AllocateWriteMem([]byte(role))
}

func wasmxExecuteCosmosMsg(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var msg cdctypes.Any
	ctx.CosmosHandler.JSONCodec().UnmarshalJSON(reqbz, &msg)

	// TODO ExecuteCosmosMsg and ExecuteCosmosQuery may trigger subcals
	// which mess up subcall hooks
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
		return nil, err
	}
	return rnh.AllocateWriteMem(responsebz)
}

func wasmxDecodeCosmosTxToJson(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	jsonbz, err := ctx.CosmosHandler.DecodeCosmosTx(reqbz)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(jsonbz)
}

func wasmxVerifyCosmosTx(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	valid, err := ctx.CosmosHandler.VerifyCosmosTx(reqbz)
	resp := VerifyCosmosTxResponse{Valid: valid, Error: ""}
	if err != nil {
		resp.Error = err.Error()
	}
	respbz, err := json.Marshal(&resp)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(respbz)
}

func wasmxCreateAccountInterpreted(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)

	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req vmtypes.CreateAccountInterpretedRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	// TODO metadata should come from the request
	metadata := types.CodeMetadata{}
	// TODO info from provenance ?
	initMsg, err := json.Marshal(types.WasmxExecutionMessage{Data: []byte{}})
	if err != nil {
		return returns, err
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
		return returns, err
	}

	contractbz := memc.PaddLeftTo32(contractAddress.Bytes())
	return rnh.AllocateWriteMem(contractbz)
}

func wasmxCreate2AccountInterpreted(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)

	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req vmtypes.Create2AccountInterpretedRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	metadata := types.CodeMetadata{}
	// TODO info from provenance ?
	initMsg, err := json.Marshal(types.WasmxExecutionMessage{Data: []byte{}})
	if err != nil {
		return returns, err
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
		return returns, err
	}

	contractbz := memc.PaddLeftTo32(contractAddress.Bytes())
	return rnh.AllocateWriteMem(contractbz)
}

func wasmxCreateAccount(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)

	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req vmtypes.InstantiateAccountRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Ctx.Logger().Debug("unmarshaling InstantiateAccountRequest", "error", err)
		return nil, err
	}
	// TODO info from provenance ?
	initMsg, err := json.Marshal(types.WasmxExecutionMessage{Data: req.Msg})
	if err != nil {
		return returns, err
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
		return returns, err
	}

	response := vmtypes.InstantiateAccountResponse{Address: *contractAddress}
	respbz, err := json.Marshal(response)
	if err != nil {
		return returns, err
	}
	return rnh.AllocateWriteMem(respbz)
}

func wasmxCreate2Account(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)

	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req vmtypes.Instantiate2AccountRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	// TODO info from provenance ?
	initMsg, err := json.Marshal(types.WasmxExecutionMessage{Data: []byte{}})
	if err != nil {
		return returns, err
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
		return returns, err
	}

	response := vmtypes.Instantiate2AccountResponse{Address: *contractAddress}
	respbz, err := json.Marshal(response)
	if err != nil {
		return returns, err
	}
	return rnh.AllocateWriteMem(respbz)
}

func prepareResponse(ctx *Context, rnh memc.RuntimeHandler, resp interface{}) (interface{}, error) {
	respbz, err := json.Marshal(&resp)
	if err != nil {
		return 0, nil
	}
	return rnh.AllocateWriteMem(respbz)
}

type MerkleSlices struct {
	Slices [][]byte `json:"slices"`
}

type FinalizeBlockWrap struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

func merkleHash(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	data, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var val MerkleSlices
	err = json.Unmarshal(data, &val)
	if err != nil {
		return nil, err
	}
	hashbz := merkle.HashFromByteSlices(val.Slices)
	return rnh.AllocateWriteMem(hashbz)
}

func ed25519Sign(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	keyptr, ndx := memc.GetPointerFromParams(rnh, params, 0)
	privbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	dataptr, _ := memc.GetPointerFromParams(rnh, params, ndx)
	msgbz, err := rnh.ReadMemFromPtr(dataptr)
	if err != nil {
		return nil, err
	}
	privKey := ed25519.PrivKey(privbz)
	signature, err := privKey.Sign(msgbz)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(signature)
}

func ed25519Verify(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	keyptr, ndx := memc.GetPointerFromParams(rnh, params, 0)
	pubkeybz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	sigptr, ndx := memc.GetPointerFromParams(rnh, params, ndx)
	signaturebz, err := rnh.ReadMemFromPtr(sigptr)
	if err != nil {
		return nil, err
	}
	msgptr, _ := memc.GetPointerFromParams(rnh, params, ndx)
	msgbz, err := rnh.ReadMemFromPtr(msgptr)
	if err != nil {
		return nil, err
	}
	pubKey := ed25519.PubKey(pubkeybz)
	isSigner := pubKey.VerifySignature(msgbz, signaturebz)
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	if isSigner {
		returns[0] = int32(1)
	}
	return returns, nil
}

// TODO replace this with sha256.Sum256(bz)[:20]
func ed25519PubToHex(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	pubkeybz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	pubKey := ed25519.PubKey(pubkeybz)
	hexAddr := pubKey.Address()
	return rnh.AllocateWriteMem(hexAddr)
}

// addr_canonicalize(string) -> ArrayBuffer;
func wasmxCanonicalize(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	addrStrBz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	addr, err := mcodec.AccAddressPrefixedFromBech32(string(addrStrBz))
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(addr.Bytes())
}

func wasmxValidateBech32Addr(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	addrStrBz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	valid := int32(0)
	_, err = mcodec.AccAddressPrefixedFromBech32(string(addrStrBz))
	if err == nil {
		valid = int32(1)
	}
	returns := make([]interface{}, 1)
	returns[0] = valid
	return returns, nil
}

func wasmxAddrEquivalent(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	keyptr, ndx := memc.GetPointerFromParams(rnh, params, 0)
	addr1StrBz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	dataptr, _ := memc.GetPointerFromParams(rnh, params, ndx)
	addr2StrBz, err := rnh.ReadMemFromPtr(dataptr)
	if err != nil {
		return nil, err
	}
	addr1, err := mcodec.AccAddressPrefixedFromBech32(string(addr1StrBz))
	if err != nil {
		return nil, err
	}
	addr2, err := mcodec.AccAddressPrefixedFromBech32(string(addr2StrBz))
	if err != nil {
		return nil, err
	}

	same := bytes.Equal(addr1.Bytes(), addr2.Bytes())
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	if same {
		returns[0] = int32(1)
	}
	return returns, nil
}

// addr_humanize(ArrayBuffer) -> string;
func wasmxHumanize(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	addrBz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrBz))
	addrstr, err := ctx.CosmosHandler.AddressCodec().BytesToString(addr)
	if err != nil {
		return nil, err
	}

	return rnh.AllocateWriteMem([]byte(addrstr))
}

// addr_humanize_mc(ArrayBuffer) -> string;
func wasmxHumanizeMultiChain(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	keyptr, ndx := memc.GetPointerFromParams(rnh, params, 0)
	addrBz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	dataptr, _ := memc.GetPointerFromParams(rnh, params, ndx)
	prefixBz, err := rnh.ReadMemFromPtr(dataptr)
	if err != nil {
		return nil, err
	}
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrBz))
	prefix := string(prefixBz)

	addrCodec := mcodec.NewBech32Codec(prefix, mcodec.NewAddressPrefixedFromAcc)

	addrstr, err := addrCodec.BytesToString(addr)
	if err != nil {
		return nil, err
	}

	return rnh.AllocateWriteMem([]byte(addrstr))
}

// addr_canonicalize_mc(string) -> ArrayBuffer;
func wasmxCanonicalizeMultiChain(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	addrStrBz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	addr, err := mcodec.AccAddressPrefixedFromBech32(string(addrStrBz))
	if err != nil {
		return nil, err
	}
	respbz, err := json.Marshal(&vmtypes.PrefixedAddress{Bz: addr.Bytes(), Prefix: addr.Prefix()})
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(respbz)
}

type LoggerLog struct {
	Msg   string   `json:"msg"`
	Parts []string `json:"parts"`
}

func getLoggerData(ctx *Context, rnh memc.RuntimeHandler, params []interface{}) (string, []any, error) {
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	message, err := rnh.ReadMemFromPtr(keyptr)
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

func wasmxLoggerInfo(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	msg, parts, err := getLoggerData(ctx, rnh, params)
	if err != nil {
		return nil, err
	}
	ctx.Logger(ctx.Ctx).Info(msg, parts...)
	returns := make([]interface{}, 0)
	return returns, nil
}

func wasmxLoggerError(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	msg, parts, err := getLoggerData(ctx, rnh, params)
	if err != nil {
		return nil, err
	}
	ctx.Logger(ctx.Ctx).Error(msg, parts...)
	returns := make([]interface{}, 0)
	return returns, nil
}

func wasmxLoggerDebug(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	msg, parts, err := getLoggerData(ctx, rnh, params)
	if err != nil {
		return nil, err
	}
	ctx.Logger(ctx.Ctx).Debug(msg, parts...)
	returns := make([]interface{}, 0)
	return returns, nil
}

func wasmxLoggerDebugExtended(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	msg, parts, err := getLoggerData(ctx, rnh, params)
	if err != nil {
		return nil, err
	}
	newmodule := GetVmLoggerExtended(ctx.Logger, ctx.Env.Chain.ChainIdFull, ctx.Env.Contract.Address.String())
	newmodule(ctx.Ctx).Debug(msg, parts...)
	returns := make([]interface{}, 0)
	return returns, nil
}

func wasmxEmitCosmosEvents(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	evsbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	// var data []types.Event
	var data []sdk.Event
	err = json.Unmarshal(evsbz, &data)
	if err != nil {
		return nil, err
	}
	// ctx.CosmosEvents = append(ctx.CosmosEvents, data...)
	ctx.Ctx.EventManager().EmitEvents(data)
	returns := make([]interface{}, 0)
	return returns, nil
}
