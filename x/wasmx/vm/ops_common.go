package vm

import (
	"slices"
	"strings"

	dbm "github.com/cometbft/cometbft-db"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
	depContext = buildExecutionContextClassic(dep)
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

	// deterministic contracts cannot transact with or query non-deterministic contracts
	sourceContract := GetContractContext(ctx, req.From)
	if sourceContract != nil {
		fromStorageType := sourceContract.ContractInfo.StorageType
		toStorageType := depContext.ContractInfo.StorageType
		if fromStorageType == types.ContractStorageType_CoreConsensus && toStorageType != types.ContractStorageType_CoreConsensus {
			// deterministic contracts can read from metaconsensus
			if toStorageType == types.ContractStorageType_MetaConsensus && !req.IsQuery {
				ctx.Ctx.Logger().Debug("deterministic contract tried to execute meta consensus contract", "from", req.From.String(), "to", req.To.String())
				return int32(1), nil
			}
			if toStorageType != types.ContractStorageType_MetaConsensus {
				ctx.Ctx.Logger().Debug("deterministic contract tried to execute non-deterministic contract", "from", req.From.String(), "to", req.To.String())
				return int32(1), nil
			}
		}
	}

	callContext := types.MessageInfo{
		Origin:   ctx.Env.CurrentCall.Origin,
		Sender:   req.From,
		Funds:    req.Value,
		CallData: req.Calldata,
		GasLimit: req.GasLimit,
	}

	to := req.To
	systemDeps := req.SystemDeps
	// clone router
	newrouter := cloneContractRouter(ctx.ContractRouter)
	// TODO req.To or to?
	routerAddress := req.To.String()

	if depContext.ContractInfo.Role == types.ROLE_LIBRARY {
		// use the sender contract if the call is to a library
		to = req.From
		// TODO
		// newrouter[to.String()].ContractInfo.
		// TODO inherit execution depepndencies comming from roles
		sysdeps := newrouter[req.To.String()].ContractInfo.SystemDeps
		for _, dep := range sourceContract.ContractInfo.SystemDeps {
			if strings.Contains(dep.Role, "consensus") {
				if !slices.Contains(systemDeps, dep.Role) {
					systemDeps = append(systemDeps, dep.Role)
				}
				found := slices.ContainsFunc(sysdeps, func(n types.SystemDep) bool {
					return n.Role == dep.Role
				})
				if !found {
					sysdeps = append(sysdeps, dep)
				}
			}
			for _, subdep := range dep.Deps {
				if strings.Contains(subdep.Role, "consensus") {
					if !slices.Contains(systemDeps, subdep.Role) {
						systemDeps = append(systemDeps, subdep.Role)
					}
					found := slices.ContainsFunc(sysdeps, func(n types.SystemDep) bool {
						return n.Role == subdep.Role
					})
					if !found {
						sysdeps = append(sysdeps, subdep)
					}
				}
			}
		}
		ci := newrouter[routerAddress].ContractInfo
		newrouter[routerAddress].ContractInfo = types.ContractDependency{
			Address:       ci.Address,
			Role:          ci.Role,
			Label:         ci.Label,
			StoreKey:      ci.StoreKey,
			FilePath:      ci.FilePath,
			Bytecode:      ci.Bytecode,
			CodeHash:      ci.CodeHash,
			CodeId:        ci.CodeId,
			StorageType:   ci.StorageType,
			SystemDepsRaw: systemDeps,
			SystemDeps:    sysdeps,
		}
		newrouter[routerAddress].ContractInfo.SystemDepsRaw = systemDeps
		newrouter[routerAddress].ContractInfo.SystemDeps = sysdeps
	}
	tempCtx, commit := ctx.Ctx.CacheContext()
	contractStore := ctx.CosmosHandler.ContractStore(tempCtx, ctx.ContractRouter[to.String()].ContractInfo.StorageType, ctx.ContractRouter[to.String()].ContractInfo.StoreKey)

	// for authorizing cosmos messages sent by the contract, we check the sender/signer is the contract
	// so we initialize the cosmos handler with the target contract
	newCosmosHandler := ctx.CosmosHandler.WithNewAddress(to)

	newctx := &Context{
		Ctx:            tempCtx,
		GasMeter:       ctx.GasMeter,
		ContractStore:  contractStore,
		CosmosHandler:  newCosmosHandler,
		ContractRouter: newrouter,
		App:            ctx.App,
		NativeHandler:  ctx.NativeHandler,
		dbIterators:    map[int32]dbm.Iterator{},
		Env: &types.Env{
			Block:       ctx.Env.Block,
			Transaction: ctx.Env.Transaction,
			Chain:       ctx.Env.Chain,
			Contract: types.EnvContractInfo{
				Address:    to,
				CodeHash:   req.CodeHash,
				Bytecode:   req.Bytecode,
				FilePath:   req.FilePath,
				CodeId:     req.CodeId,
				SystemDeps: systemDeps,
			},
			CurrentCall: callContext,
		},
	}

	_, err := newrouter[routerAddress].Execute(newctx)
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
	return success, newctx.ReturnData
}

func cloneContractRouter(router map[string]*ContractContext) map[string]*ContractContext {
	newrouter := make(map[string]*ContractContext, 0)
	for k := range router {
		newrouter[k] = router[k].CloneShallow()
	}
	return newrouter
}
