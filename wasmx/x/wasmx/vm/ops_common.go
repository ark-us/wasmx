package vm

import (
	"fmt"
	"slices"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm/types"
)

func GetContractDependency(ctx *Context, addr mcodec.AccAddressPrefixed) *types.ContractDependency {
	depContext, ok := ctx.ContractRouter[addr.String()]
	if ok {
		return depContext.ContractInfo
	}
	dep, err := ctx.CosmosHandler.GetContractDependency(ctx.Ctx, addr)
	if err != nil {
		return nil
	}
	// cache it
	ctx.ContractRouter[addr.String()] = &Context{
		ContractInfo: &dep,
	}
	return &dep
}

func BankGetBalance(ctx *Context, addr mcodec.AccAddressPrefixed, denom string) (sdk.Coin, error) {
	alias, found := ctx.CosmosHandler.GetAlias(addr)
	if found {
		addr = alias
	}
	msg := &banktypes.QueryBalanceRequest{Address: addr.String(), Denom: denom}
	bankmsgbz, err := ctx.CosmosHandler.JSONCodec().MarshalJSON(msg)
	if err != nil {
		return sdk.Coin{}, err
	}
	msgbz := []byte(fmt.Sprintf(`{"GetBalance":%s}`, string(bankmsgbz)))
	res, err := BankCall(ctx, msgbz, true)
	if err != nil {
		return sdk.Coin{}, err
	}
	var response banktypes.QueryBalanceResponse
	err = ctx.CosmosHandler.JSONCodec().UnmarshalJSON(res, &response)
	if err != nil {
		return sdk.Coin{}, err
	}

	return *response.Balance, nil
}

func BankSendCoin(ctx *Context, from mcodec.AccAddressPrefixed, to mcodec.AccAddressPrefixed, amount sdk.Coins) error {
	aliasFrom, found := ctx.CosmosHandler.GetAlias(from)
	if found {
		from = aliasFrom
	}
	aliasTo, found := ctx.CosmosHandler.GetAlias(to)
	if found {
		to = aliasTo
	}
	msg := &banktypes.MsgSend{
		FromAddress: from.String(),
		ToAddress:   to.String(),
		Amount:      amount,
	}
	bankmsgbz, err := ctx.CosmosHandler.JSONCodec().MarshalJSON(msg)
	if err != nil {
		return err
	}
	msgbz := []byte(fmt.Sprintf(`{"SendCoins":%s}`, string(bankmsgbz)))
	_, err = BankCall(ctx, msgbz, false)
	return err
}

func BankCall(ctx *Context, msgbz []byte, isQuery bool) ([]byte, error) {
	// initMsg, err := json.Marshal(types.WasmxExecutionMessage{Data: []byte{}})
	// if err != nil {
	// 	return returns, err
	// }

	bankAddress, err := ctx.GetCosmosHandler().GetAddressOrRole(ctx.Ctx, types.ROLE_BANK)
	if err != nil {
		return nil, err
	}
	req := vmtypes.CallRequestCommon{
		To:       bankAddress,
		From:     bankAddress, // TODO wasmx?
		Calldata: msgbz,
		IsQuery:  isQuery,
	}
	success, data := WasmxCall(ctx, req)
	if success > 0 {
		return nil, fmt.Errorf("bank call failed: %s", string(data))
	}
	return data, nil
}

// All WasmX, eWasm calls must go through here
// Returns 0 on success, 1 on failure and 2 on revert
func WasmxCall(ctx *Context, req vmtypes.CallRequestCommon) (int32, []byte) {
	if types.IsSystemAddress(req.To.Bytes()) && !ctx.CosmosHandler.CanCallSystemContract(ctx.Ctx, req.From) {
		return int32(1), []byte(`wasmxcall: cannot call system contract`)
	}
	depContractInfo := GetContractDependency(ctx, req.To)
	// ! we return success here in case the contract does not exist
	// an empty transaction to any account should succeed (evm way)
	// even with value 0 & no calldata
	if depContractInfo == nil {
		return int32(0), []byte(`wasmxcall: cannot get contract context`)
	}

	fromstr := req.From.String()
	tostr := req.To.String()

	// deterministic contracts cannot transact with or query non-deterministic contracts
	sourceContractInfo := GetContractDependency(ctx, req.From)
	if sourceContractInfo != nil {
		if depContractInfo.Role != "" && sourceContractInfo.Role == "" {
			errmsg := "no-role contract tried to execute role contract"
			ctx.Ctx.Logger().Debug(errmsg, "from", fromstr, "from_role", sourceContractInfo.Role, "to", tostr, "to_role", depContractInfo.Role)
			errmsg += fmt.Sprintf(": from %s, to %s - %s", fromstr, tostr, depContractInfo.Role)
			return int32(1), []byte(errmsg)
		}

		fromStorageType := sourceContractInfo.StorageType
		toStorageType := depContractInfo.StorageType
		if fromStorageType == types.ContractStorageType_CoreConsensus && toStorageType != types.ContractStorageType_CoreConsensus {
			// deterministic contracts can read & write from/to metaconsensus contracts
			if toStorageType != types.ContractStorageType_MetaConsensus {
				errmsg := "deterministic contract tried to execute non-deterministic contract"
				ctx.Ctx.Logger().Debug(errmsg, "from", fromstr, "to", tostr)
				errmsg += fmt.Sprintf(": from %s, to %s", fromstr, tostr)
				return int32(1), []byte(errmsg)
			}
		}
		// TODO execution should be stopped, queries are ok
		// if fromStorageType == types.ContractStorageType_SingleConsensus && toStorageType == types.ContractStorageType_CoreConsensus {
		// 	errmsg := "non-deterministic contract tried to execute deterministic (core consensus) contract"
		// 	ctx.Ctx.Logger().Debug(errmsg, "from", fromstr, "to", tostr)
		// 	errmsg += fmt.Sprintf(": from %s, to %s", fromstr, tostr)
		// 	return int32(1), []byte(errmsg)
		// }

		// right now single consensus contracts can call metaconsensus contracts
		// and this is used to update chain information for a simple node statesync when the node is not a validator
	}
	callContext := types.MessageInfo{
		Origin:   ctx.Env.CurrentCall.Origin,
		Sender:   req.From,
		Funds:    req.Value,
		CallData: req.Calldata,
		GasLimit: req.GasLimit,
	}

	to := req.To
	tostr2 := tostr
	systemDeps := req.SystemDeps
	// clone router
	newrouter := cloneContractRouter(ctx.ContractRouter)
	// TODO req.To or to?
	routerAddress := tostr

	if depContractInfo.Role == types.ROLE_LIBRARY {
		// use the sender contract if the call is to a library
		to = req.From
		tostr2 = fromstr
		// TODO
		// newrouter[tostr2].ContractInfo.
		// TODO inherit execution depepndencies comming from roles
		sysdeps := newrouter[tostr].ContractInfo.SystemDeps
		for _, dep := range sourceContractInfo.SystemDeps {
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
		newrouter[routerAddress].ContractInfo = &types.ContractDependency{
			Address:       ci.Address,
			Role:          ci.Role,
			Label:         ci.Label,
			StoreKey:      ci.StoreKey,
			CodeFilePath:  ci.CodeFilePath,
			AotFilePath:   ci.AotFilePath,
			Bytecode:      ci.Bytecode,
			CodeHash:      ci.CodeHash,
			CodeId:        ci.CodeId,
			StorageType:   ci.StorageType,
			Pinned:        ci.Pinned,
			SystemDepsRaw: systemDeps,
			SystemDeps:    sysdeps,
		}
		newrouter[routerAddress].ContractInfo.SystemDepsRaw = systemDeps
		newrouter[routerAddress].ContractInfo.SystemDeps = sysdeps
	}
	tempCtx, commit := ctx.Ctx.CacheContext()
	contractStore := ctx.CosmosHandler.ContractStore(tempCtx, ctx.ContractRouter[tostr2].ContractInfo.StorageType, ctx.ContractRouter[tostr2].ContractInfo.StoreKey)

	// for authorizing cosmos messages sent by the contract, we check the sender/signer is the contract
	// so we initialize the cosmos handler with the target contract
	newCosmosHandler := ctx.CosmosHandler.WithNewAddress(to)
	sysDeps := newrouter[routerAddress].ContractInfo.SystemDeps
	pinned := newrouter[routerAddress].ContractInfo.Pinned
	rnh := getRuntimeHandler(ctx.newIVmFn, tempCtx, sysDeps, pinned)
	newctx := &Context{
		GoRoutineGroup:  ctx.GoRoutineGroup,
		GoContextParent: ctx.GoContextParent,
		Ctx:             tempCtx,
		Logger:          GetVmLogger(ctx.Logger, ctx.Env.Chain.ChainIdFull, to.String()),
		GasMeter:        ctx.GasMeter,
		ContractStore:   contractStore,
		CosmosHandler:   newCosmosHandler,
		ContractRouter:  newrouter,
		App:             ctx.App,
		NativeHandler:   ctx.NativeHandler,
		dbIterators:     map[int32]types.Iterator{},
		RuntimeHandler:  rnh,
		newIVmFn:        ctx.newIVmFn,
		ContractInfo:    newrouter[routerAddress].ContractInfo,
		Env: &types.Env{
			Block:       ctx.Env.Block,
			Transaction: ctx.Env.Transaction,
			Chain:       ctx.Env.Chain,
			Contract: types.EnvContractInfo{
				Address:    to,
				CodeHash:   req.CodeHash,
				Bytecode:   req.Bytecode,
				CodeId:     req.CodeId,
				SystemDeps: systemDeps,
			},
			CurrentCall: callContext,
		},
	}
	_, err := newctx.Execute()
	var success int32
	// Returns 0 on success, 1 on failure and 2 on revert
	if err != nil {
		success = int32(2)
		// note, just log the error here, because it may contain non-deterministic data added by the WASM runtime
		newctx.Logger(newctx.Ctx).Debug(err.Error())
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

// we only need to forward ContractInfo for nested calls
func cloneContractRouter(router map[string]*Context) map[string]*Context {
	newrouter := make(map[string]*Context, 0)
	for k := range router {
		if router[k].ContractInfo == nil {
			newrouter[k] = &Context{}
		} else {
			newrouter[k] = &Context{
				ContractInfo: router[k].ContractInfo.Clone(),
			}
		}
	}
	return newrouter
}
