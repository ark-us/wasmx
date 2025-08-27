package vmmc

import (
	"encoding/json"

	mcfg "github.com/loredanacirstea/wasmx/config"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

// InitSubChain(*InitSubChainMsg) (*abci.ResponseInitChain, error)
func InitSubChain(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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
	return rnh.AllocateWriteMem(responsebz)
}

// StartSubChain(StartSubChainMsg): void
func StartSubChain(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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
	return prepareResponse(rnh, response)
}

func GetSubChainIds(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	multichainapp, err := mcfg.GetMultiChainApp(ctx.GoContextParent)
	if err != nil {
		return nil, err
	}
	return prepareResponse(rnh, multichainapp.ChainIds)
}

// this is what we use to statesync subchains
// StartStateSyncRequest(StateSyncRequestMsg): void
func StartStateSyncRequest(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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
	return prepareResponse(rnh, response)
}

func prepareResponse(rnh memc.RuntimeHandler, response interface{}) ([]interface{}, error) {
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(responsebz)
}
