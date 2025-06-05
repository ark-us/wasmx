package vmsql

import (
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func ConnectMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SqlConnectionResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func CloseMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SqlCloseResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func PingMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SqlPingResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func ExecuteMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SqlExecuteResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func BatchAtomicMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SqlExecuteResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func QueryMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &SqlQueryResponse{Error: ""}
	return prepareResponse(rnh, response)
}
