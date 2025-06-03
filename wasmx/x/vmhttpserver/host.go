package vmhttpserver

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	cfg "github.com/loredanacirstea/wasmx/config"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func StartWebServer(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &StartWebServerResponse{Error: ""}
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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

func GenerateJWT(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &GenerateJWTResponse{Error: ""}
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req GenerateJWTRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	claims := buildClaims(req.Claims, req.AdditionalClaim)

	var method jwt.SigningMethod
	switch req.SigningMethod {
	case "HS256":
		method = jwt.SigningMethodHS384
	case "HS384":
		method = jwt.SigningMethodHS384
	case "HS512":
		method = jwt.SigningMethodHS512
	case "":
		method = jwt.SigningMethodHS384
	default:
		response.Error = "invalid signing method"
		return prepareResponse(rnh, response)
	}

	token := jwt.NewWithClaims(method, claims)
	signed, err := token.SignedString(req.Secret)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	response.Token = signed
	return prepareResponse(rnh, response)
}

func VerifyJWT(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &VerifyJWTResponse{Error: ""}
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req VerifyJWTRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	claims := buildClaims(req.Claims, req.AdditionalClaim)

	token, err := jwt.ParseWithClaims(req.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return req.Secret, nil
	})
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	response.Valid = token.Valid
	return prepareResponse(rnh, response)
}

func NewExpirationTime(expirationMs int64) time.Time {
	return time.Now().Add(time.Duration(expirationMs) * time.Millisecond)
}

func prepareResponse(rnh memc.RuntimeHandler, response interface{}) ([]interface{}, error) {
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(responsebz)
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

		// temporary, these should be provided by a smart contract
		vm.BuildFn("GenerateJWT", GenerateJWT, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("VerifyJWT", VerifyJWT, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "httpserver", context, fndefs)
}
