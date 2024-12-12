package vmmc

import (
	"encoding/json"

	mcfg "github.com/loredanacirstea/wasmx/v1/config"
	vmtypes "github.com/loredanacirstea/wasmx/v1/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/v1/x/wasmx/vm/memory/common"
)

// InitSubChain(*InitSubChainMsg) (*abci.ResponseInitChain, error)
func InitSubChain(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req InitSubChainMsg
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	response, err := InitApp(ctx, &req)
	if err != nil {
		ctx.Logger(ctx.Ctx).Error("could not initiate subchain app", "error", err.Error())
		return nil, err
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	ptr, err := rnh.AllocateWriteMem(responsebz)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, nil
}

// StartSubChain(StartSubChainMsg): void
func StartSubChain(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req StartSubChainMsg
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	err = StartApp(ctx, &req)
	response := &StartSubChainResponse{Error: ""}
	if err != nil {
		ctx.Logger(ctx.Ctx).Error("could not start subchain app", "error", err.Error())
		response.Error = err.Error()
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	ptr, err := rnh.AllocateWriteMem(responsebz)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, nil
}

func GetSubChainIds(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	multichainapp, err := mcfg.GetMultiChainApp(ctx.GoContextParent)
	if err != nil {
		return nil, err
	}
	responsebz, err := json.Marshal(multichainapp.ChainIds)
	if err != nil {
		return nil, err
	}
	ptr, err := rnh.AllocateWriteMem(responsebz)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, nil
}

// this is what we use to statesync subchains
// StartStateSyncRequest(StateSyncRequestMsg): void
func StartStateSyncRequest(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req StateSyncRequestMsg
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	err = StartStateSyncWithChainId(ctx, req)
	response := &StartSubChainResponse{Error: ""}
	if err != nil {
		ctx.Logger(ctx.Ctx).Error("could not start subchain app", "error", err.Error())
		response.Error = err.Error()
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	ptr, err := rnh.AllocateWriteMem(responsebz)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, nil
}

func BuildWasmxMultichainJson1(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("InitSubChain", InitSubChain, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("StartSubChain", StartSubChain, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("GetSubChainIds", GetSubChainIds, []interface{}{}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("StartStateSync", StartStateSyncRequest, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "multichain", context, fndefs)
}
