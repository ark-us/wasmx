package vmhttpserver

import (
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func StartWebServerMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &StartWebServerResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func SetRouteHandlerMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SetRouteHandlerResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func RemoveRouteHandlerMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &RemoveRouteHandlerResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func CloseMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &CloseResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func GenerateJWTMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &GenerateJWTResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func VerifyJWTMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &VerifyJWTResponse{Error: ""}
	return prepareResponse(rnh, response)
}
