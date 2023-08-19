package vm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tm-db"

	"mythos/v1/x/wasmx/types"
	vmtypes "mythos/v1/x/wasmx/vm/types"
)

// Returns nil if there is no contract
func GetContractContext(ctx *Context, addr sdk.AccAddress) *ContractContext {
	depContext, ok := ctx.ContractRouter[addr.String()]
	if ok {
		return depContext
	}
	dep, err := ctx.CosmosHandler.GetContractDependency(ctx.Ctx, addr)
	if err != nil {
		return nil
	}
	depContext = buildExecutionContextClassic(dep.FilePath, dep.Bytecode, dep.CodeHash, dep.StoreKey, dep.SystemDeps, nil)
	ctx.ContractRouter[addr.String()] = depContext
	return depContext
}

// All WasmX, eWasm calls must go through here
// Returns 0 on success, 1 on failure and 2 on revert
func WasmxCall(ctx *Context, req vmtypes.CallRequest) (int32, []byte) {
	if types.IsSystemAddress(req.To) && !ctx.CosmosHandler.CanCallSystemContract(ctx.Ctx, req.From) {
		return int32(1), nil
	}
	depContext := GetContractContext(ctx, req.To)
	// ! we return success here in case the contract does not exist
	// an empty transaction to any account should succeed (evm way)
	// even with value 0 & no calldata
	if depContext == nil {
		return int32(0), nil
	}

	callContext := types.MessageInfo{
		Origin:   ctx.Env.CurrentCall.Origin,
		Sender:   req.From,
		Funds:    req.Value,
		CallData: req.Calldata,
		GasLimit: req.GasLimit,
	}

	tempCtx, commit := ctx.Ctx.CacheContext()
	contractStore := ctx.CosmosHandler.ContractStore(tempCtx, ctx.ContractRouter[req.To.String()].ContractStoreKey)

	newctx := &Context{
		Ctx:            tempCtx,
		GasMeter:       ctx.GasMeter,
		ContractStore:  contractStore,
		CosmosHandler:  ctx.CosmosHandler,
		ContractRouter: ctx.ContractRouter,
		NativeHandler:  ctx.NativeHandler,
		dbIterators:    map[int32]dbm.Iterator{},
		Env: &types.Env{
			Block:       ctx.Env.Block,
			Transaction: ctx.Env.Transaction,
			Chain:       ctx.Env.Chain,
			Contract: types.EnvContractInfo{
				Address:  req.To,
				CodeHash: req.CodeHash,
				Bytecode: req.Bytecode,
				FilePath: req.FilePath,
			},
			CurrentCall: callContext,
		},
	}
	_, err := newctx.ContractRouter[req.To.String()].Execute(newctx)
	var success int32
	// Returns 0 on success, 1 on failure and 2 on revert
	if err != nil {
		success = int32(2)
		newctx.GetContext().Logger().Debug(err.Error())
	} else {
		success = int32(0)
		if !req.IsQuery {
			commit()
			// Write events
			ctx.Ctx.EventManager().EmitEvents(tempCtx.EventManager().Events())
			ctx.Logs = append(ctx.Logs, newctx.Logs...)
		}
	}
	ctx.ReturnData = newctx.ReturnData
	return success, newctx.ReturnData
}
