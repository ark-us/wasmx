package vm

import (
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func BuildWasmxEnvi64(context *Context, rnh memc.RuntimeHandler) (interface{}, error) {
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("sha256", sha256, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getCallData", getCallData, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getEnv", getEnv, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getChainId", wasmxGetChainId, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getCaller", wasmxGetCaller, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getAddress", wasmxGetAddress, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("storageLoad", wasmxStorageLoad, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("storageStore", wasmxStorageStore, []interface{}{vm.ValType_I64(), vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("storageDelete", wasmxStorageDelete, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("storageDeleteRange", wasmxStorageDeleteRange, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("storageLoadRange", wasmxStorageLoadRange, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("storageLoadRangePairs", wasmxStorageLoadRangePairs, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("log", wasmxLog, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("emitCosmosEvents", wasmxEmitCosmosEvents, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("getReturnData", wasmxGetReturnData, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getFinishData", wasmxGetFinishData, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("setFinishData", wasmxSetFinishData, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		// TODO some precompiles use setReturnData instead of setFinishData
		vm.BuildFn("setReturnData", wasmxSetFinishData, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("finish", wasmxFinish, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("revert", wasmxRevert, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("getBlockHash", wasmxGetBlockHash, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getCurrentBlock", wasmxGetCurrentBlock, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getAccount", getAccount, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getBalance", wasmxGetBalance, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("call", wasmxCall, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("keccak256", keccak256Util, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("createAccountInterpreted", wasmxCreateAccountInterpreted, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("create2AccountInterpreted", wasmxCreate2AccountInterpreted, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("createAccount", wasmxCreateAccount, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("create2Account", wasmxCreate2Account, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("MerkleHash", merkleHash, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("LoggerInfo", wasmxLoggerInfo, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("LoggerError", wasmxLoggerError, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("LoggerDebug", wasmxLoggerDebug, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("LoggerDebugExtended", wasmxLoggerDebugExtended, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("ed25519Sign", ed25519Sign, []interface{}{vm.ValType_I64(), vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("ed25519Verify", ed25519Verify, []interface{}{vm.ValType_I64(), vm.ValType_I64(), vm.ValType_I64()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ed25519PubToHex", ed25519PubToHex, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("validate_bech32_address", wasmxValidateBech32Addr, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("addr_humanize", wasmxHumanize, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("addr_canonicalize", wasmxCanonicalize, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("addr_equivalent", wasmxAddrEquivalent, []interface{}{vm.ValType_I64(), vm.ValType_I64()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("addr_humanize_mc", wasmxHumanizeMultiChain, []interface{}{vm.ValType_I64(), vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("addr_canonicalize_mc", wasmxCanonicalizeMultiChain, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),

		vm.BuildFn("getAddressByRole", wasmxGetAddressByRole, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getRoleByAddress", wasmxGetRoleByAddress, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),

		vm.BuildFn("executeCosmosMsg", wasmxExecuteCosmosMsg, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("decodeCosmosTxToJson", wasmxDecodeCosmosTxToJson, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("verifyCosmosTx", wasmxVerifyCosmosTx, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),

		// TODO
		// env.AddFunction("ProtoMarshal", NewFunction(functype__i32, ProtoMarshal, context, 0))
		// env.AddFunction("ProtoUnmarshal", NewFunction(functype__i32, ProtoUnmarshal, context, 0))
	}

	return vm.BuildModule(rnh, "wasmx", context, fndefs)
}
