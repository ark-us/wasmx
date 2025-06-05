package vmkv

import (
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func ConnectMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvConnectionResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func CloseMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvCloseResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func GetMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvGetResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func HasMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvHasResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func SetMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvSetResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func DeleteMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvDeleteResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func IteratorMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvIteratorResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func NewBatchMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvNewBatchResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func PrintMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	return returns, nil
}

func StatsMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := map[string]string{}
	return prepareResponse(rnh, response)
}
