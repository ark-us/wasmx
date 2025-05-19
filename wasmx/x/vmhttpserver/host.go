package vmhttpserver

import (
	"encoding/json"
	"fmt"

	cfg "github.com/loredanacirstea/wasmx/config"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func StartWebServer(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &StartWebServerResponse{Error: ""}
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req StartWebServerRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	hctx, err := GetHttpServerContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	if hctx.Server != nil {
		return prepareResponse(rnh, response)
	}
	mapp, ok := ctx.App.(cfg.MythosApp)
	if !ok {
		response.Error = "error App interface from multichainapp"
		return prepareResponse(rnh, response)
	}

	httpSrv, websrvServer, httpSrvDone, err := StartWebsrv(ctx.GetCosmosHandler(), ctx.GoContextParent, ctx.GoRoutineGroup, ctx.Ctx.Logger(), &req.Config, mapp.GetActionExecutor(), ctx.Env.Contract.Address.String())
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	hctx.Server = httpSrv
	hctx.WebsrvServer = websrvServer
	hctx.ServerChannelDone = httpSrvDone

	// close server when context is closing
	ctx.GoRoutineGroup.Go(func() error {
		select {
		case <-ctx.GoContextParent.Done():
			ctx.Ctx.Logger().Info("parent context was closed, closing webserver")
			hctx, err := GetHttpServerContext(ctx.Context.GoContextParent)
			if err != nil {
				ctx.Ctx.Logger().Error(fmt.Sprintf(`webserver close error: cannot find server instance %v`, err))
			}
			err = hctx.Server.Close()
			if err != nil {
				ctx.Ctx.Logger().Error(fmt.Sprintf(`webserver close error: %v`, err))
			}
			close(httpSrvDone)
			return nil
		case <-httpSrvDone:
			// when close signal is received from Close() API
			// webserver is already closed, so we exit this goroutine
			ctx.Ctx.Logger().Info("webserver connection closed")
			return nil
		}
	})

	return prepareResponse(rnh, response)
}

func SetRouteHandler(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SetRouteHandlerResponse{Error: ""}
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req SetRouteHandlerRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	hctx, err := GetHttpServerContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	if hctx.Server == nil || hctx.WebsrvServer == nil {
		response.Error = "server not started"
		return prepareResponse(rnh, response)
	}
	hctx.WebsrvServer.cfg.RouteToContractAddress[req.Route] = req.ContractAddress
	return prepareResponse(rnh, response)
}

func RemoveRouteHandler(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &RemoveRouteHandlerResponse{Error: ""}
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req RemoveRouteHandlerRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	hctx, err := GetHttpServerContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	if hctx.Server == nil || hctx.WebsrvServer == nil {
		response.Error = "server not started"
		return prepareResponse(rnh, response)
	}
	delete(hctx.WebsrvServer.cfg.RouteToContractAddress, req.Route)

	return prepareResponse(rnh, response)
}

func Close(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &CloseResponse{Error: ""}
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req CloseRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	hctx, err := GetHttpServerContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	if hctx.Server == nil || hctx.WebsrvServer == nil {
		return prepareResponse(rnh, response)
	}
	err = hctx.Server.Close()
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

func prepareResponse(rnh memc.RuntimeHandler, response interface{}) ([]interface{}, error) {
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

// only one contract should have the role to handle http server routes
// this contract has an internal registry and other contracts
func BuildWasmxHttpServer(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("StartWebServer", StartWebServer, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("SetRouteHandler", SetRouteHandler, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("RemoveRouteHandler", RemoveRouteHandler, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Close", Close, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "httpserver", context, fndefs)
}
