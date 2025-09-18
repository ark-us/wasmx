package vm

import (
	"encoding/hex"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm/types"
)

func GetContractDependency(ctx *Context, addr mcodec.AccAddressPrefixed) *types.ContractDependency {
	depContext, ok := ctx.ContractRouter[addr.String()]
	if ok {
		if depContext.ContractInfo.Role != types.ROLE_LIBRARY {
			return depContext.ContractInfo
		}
	}
	dep, err := ctx.CosmosHandler.GetContractDependency(ctx.Ctx, addr)
	if err != nil {
		return nil
	}
	if dep == nil {
		return nil
	}
	// cache it
	ctx.ContractRouter[addr.String()] = &Context{
		WasmxAuthority: ctx.WasmxAuthority,
		ContractInfo:   dep,
	}
	return dep
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
	bankAddress, err := ctx.GetCosmosHandler().GetAddressOrRole(ctx.Ctx, types.ROLE_BANK)
	if err != nil {
		return nil, err
	}
	contractInfo, err := ctx.GetCosmosHandler().GetContractInfo(bankAddress)
	if err != nil {
		return nil, err
	}
	codeInfo := ctx.GetCosmosHandler().GetCodeInfo(contractInfo.CodeId)
	if codeInfo == nil {
		return nil, fmt.Errorf("BankCall: code info null")
	}
	remaining := ctx.GasMeter.GasRemaining()
	req := vmtypes.CallRequestCommon{
		To:          bankAddress,
		From:        bankAddress, // TODO wasmx?
		Calldata:    msgbz,
		Value:       sdkmath.ZeroInt(),
		GasLimit:    sdkmath.NewIntFromUint64(remaining),
		IsQuery:     isQuery,
		Bytecode:    codeInfo.InterpretedBytecodeRuntime,
		CodeHash:    codeInfo.CodeHash,
		SystemDeps:  codeInfo.Deps,
		Pinned:      codeInfo.Pinned,
		MeteringOff: codeInfo.MeteringOff,
		CodeId:      contractInfo.CodeId,
		StorageType: contractInfo.StorageType.String(),
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
	// reset previous return data from previous internal calls
	ctx.ReturnData = []byte{}

	appWithHooks, appWithHooksEnabled := ctx.App.(AppWithSubCallHook)
	if types.IsSystemAddress(req.To.Bytes()) && !ctx.CosmosHandler.CanCallSystemContract(ctx.Ctx, req.From) {
		return int32(1), []byte(`wasmxcall: cannot call system contract`)
	}
	storageType, ok := types.ContractStorageType_value[req.StorageType]
	if !ok {
		errmsg := fmt.Sprintf("invalid contract storage type: %s", req.StorageType)
		return int32(1), []byte(errmsg)
	}
	toStorageType := types.ContractStorageType(storageType)

	storeKey := []byte{}
	if req.StorageAddress == nil || req.StorageAddress.Empty() {
		storeKey = types.GetContractStorePrefix(req.To.Bytes())
	} else {
		storeKey = types.GetContractStorePrefix(req.StorageAddress.Bytes())
	}
	contractAddress := req.To
	if req.ContractAddress != nil && !req.ContractAddress.Empty() {
		contractAddress = *req.ContractAddress
	}

	// ! we return success here in case the contract does not exist
	// an empty transaction to any account should succeed (evm way)
	// even with value 0 & no calldata
	// if req.CodeId == 0 {
	// 	return int32(0), []byte(`wasmxcall: cannot get contract context`)
	// }

	fromstr := req.From.String()
	tostr := req.To.String()

	// deterministic contracts cannot transact with or query non-deterministic contracts
	// TODO eventually move this rule in contracts
	sourceContractInfo := GetContractDependency(ctx, req.From)
	if sourceContractInfo != nil {
		fromStorageType := sourceContractInfo.StorageType
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
		Funds:    req.Value.BigInt(),
		CallData: req.Calldata,
		GasLimit: req.GasLimit.BigInt(),
	}
	systemDeps := req.SystemDeps
	// clone router
	newrouter := cloneContractRouter(ctx.ContractRouter)
	tempCtx, commit := ctx.Ctx.CacheContext()
	contractStore := ctx.CosmosHandler.ContractStore(tempCtx, toStorageType, storeKey)

	// for authorizing cosmos messages sent by the contract, we check the sender/signer is the contract
	// so we initialize the cosmos handler with the target contract
	newCosmosHandler := ctx.CosmosHandler.WithNewAddress(req.To)
	sysDeps := ctx.CosmosHandler.SystemDepsFromCodeDeps(ctx.Ctx, req.SystemDeps)
	pinned := req.Pinned
	destContractInfo := &types.ContractDependency{
		Address:       req.To,
		Role:          "",
		RoleLabel:     "",
		Label:         "",
		SystemDeps:    sysDeps,
		Bytecode:      req.Bytecode,
		CodeHash:      req.CodeHash,
		CodeId:        req.CodeId,
		SystemDepsRaw: systemDeps,
		Pinned:        req.Pinned,
		MeteringOff:   req.MeteringOff,
		StorageType:   toStorageType,
		StoreKey:      storeKey,
	}


	role := ctx.CosmosHandler.GetRoleByContractAddress(ctx.Ctx, req.To)
	if role != nil {
		destContractInfo.Role = role.Role
		destContractInfo.RoleLabel = role.Labels[role.Primary]
	}

	destContractInfo.CodeFilePath = ctx.CosmosHandler.GetCodeFilePath(req.CodeHash, systemDeps, req.Bytecode)
	if req.Pinned {
		destContractInfo.AotFilePath = ctx.CosmosHandler.GetAotFilePath(req.CodeHash)
	}

	rnh := getRuntimeHandler(ctx.newIVmFn, tempCtx, sysDeps, pinned)
	// increase current call count at this level
	ctx.CurrentSubCallLevelCount += 1

	newctx := &Context{
		WasmxAuthority:           ctx.WasmxAuthority,
		GoRoutineGroup:           ctx.GoRoutineGroup,
		GoContextParent:          ctx.GoContextParent,
		Ctx:                      tempCtx,
		Logger:                   GetVmLogger(ctx.Logger, ctx.Env.Chain.ChainIdFull, req.To.String()),
		GasMeter:                 ctx.GasMeter,
		ContractStore:            contractStore,
		CosmosHandler:            newCosmosHandler,
		ContractRouter:           newrouter,
		App:                      ctx.App,
		NativeHandler:            ctx.NativeHandler,
		dbIterators:              map[int32]types.Iterator{},
		RuntimeHandler:           rnh,
		newIVmFn:                 ctx.newIVmFn,
		ContractInfo:             destContractInfo,
		CurrentSubCallLevel:      ctx.CurrentSubCallLevel + 1,
		CurrentSubCallId:         ctx.CurrentSubCallLevelCount,
		CurrentSubCallLevelCount: 0, // new level, reset to 0
		Env: &types.Env{
			Block:       ctx.Env.Block,
			Transaction: ctx.Env.Transaction,
			Chain:       ctx.Env.Chain,
			Contract: types.EnvContractInfo{
				Address:    contractAddress,
				CodeHash:   req.CodeHash,
				Bytecode:   req.Bytecode,
				CodeId:     req.CodeId,
				SystemDeps: systemDeps,
			},
			CurrentCall: callContext,
		},
	}
	newrouter[tostr] = newctx

	if appWithHooksEnabled {
		err := appWithHooks.BeginSubCall(newctx.Ctx, newctx.CurrentSubCallLevel, newctx.CurrentSubCallId, req.IsQuery)
		if err != nil {
			errmsg := fmt.Sprintf("BeginSubCall error: %s", err.Error())
			return int32(1), []byte(errmsg)
		}
	}
	_, err := newctx.Execute()
	var success int32
	returnData := newctx.ReturnData
	// Returns 0 on success, 1 on failure and 2 on revert
	if err != nil {
		success = int32(2)
		// note, just log the error here, because it may contain non-deterministic data added by the WASM runtime
		newctx.Logger(newctx.Ctx).Debug("internal call failed", "error", err.Error(), "data", string(newctx.ReturnData), "from", newctx.Env.CurrentCall.Sender.String(), "to", newctx.Env.Contract.Address.String())
		LoggerExtended(newctx).Debug("internal call failed", "to", newctx.Env.Contract.Address.String(), "calldata", hex.EncodeToString(newctx.Env.CurrentCall.CallData))
	} else {
		success = int32(0)
		if !req.IsQuery {
			commit()
			// Write events
			ctx.Ctx.EventManager().EmitEvents(tempCtx.EventManager().Events())
			ctx.Logs = append(ctx.Logs, newctx.Logs...)
		}
	}
	if appWithHooksEnabled {
		err2 := appWithHooks.EndSubCall(newctx.Ctx, newctx.CurrentSubCallLevel, newctx.CurrentSubCallId, req.IsQuery, err)
		if err2 != nil {
			newctx.Logger(newctx.Ctx).Debug("EndSubCall error: "+err2.Error(), "data", string(newctx.ReturnData))
			errmsg := fmt.Sprintf("BeginSubCall error: %s", err2.Error())
			return int32(1), []byte(errmsg)
		}
	}
	return success, returnData
}

// we only need to forward ContractInfo for nested calls
func cloneContractRouter(router map[string]*Context) map[string]*Context {
	newrouter := make(map[string]*Context, 0)
	for k := range router {
		if router[k].ContractInfo == nil {
			newrouter[k] = &Context{WasmxAuthority: router[k].WasmxAuthority}
		} else {
			newrouter[k] = &Context{
				WasmxAuthority: router[k].WasmxAuthority,
				ContractInfo:   router[k].ContractInfo.Clone(),
			}
		}
	}
	return newrouter
}
