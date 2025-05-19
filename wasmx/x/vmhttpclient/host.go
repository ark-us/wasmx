package vmhttpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func Request(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &HttpResponseWrap{Error: ""}
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var reqw HttpRequestWrap
	err = json.Unmarshal(requestbz, &reqw)
	if err != nil {
		return nil, err
	}
	req := reqw.Request

	httpreq, err := http.NewRequest(req.Method, req.Url, bytes.NewBuffer(req.Data))
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	for key, values := range req.Header {
		for _, value := range values {
			httpreq.Header.Add(key, value)
		}
	}

	// TODO multipart uploads

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(httpreq)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	defer resp.Body.Close()

	response.Data = HttpResponse{
		Status:        resp.Status,
		StatusCode:    resp.StatusCode,
		ContentLength: resp.ContentLength,
		Uncompressed:  resp.Uncompressed,
		Header:        resp.Header,
	}

	if reqw.ResponseHandler.FilePath != "" {
		// TODO download to file & set the apropriate extension based on Content-Type
		response.Error = "not implemented"
		return prepareResponse(rnh, response)
	}
	if reqw.ResponseHandler.MaxSize > 0 && reqw.ResponseHandler.MaxSize < resp.ContentLength {
		response.Error = "http response body exceeds max length"
		return prepareResponse(rnh, response)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	response.Data.Data = body
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

func BuildWasmxHttpClient(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("Request", Request, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "httpclient", context, fndefs)
}
