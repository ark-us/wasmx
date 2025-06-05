package vmhttpclient

import (
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func RequestMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &HttpResponseWrap{Error: ""}
	return prepareResponse(rnh, response)
}
