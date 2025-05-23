package vmhttpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	httpreq, err := BuildHttpRequest(reqw.Request)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(httpreq)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	defer resp.Body.Close()

	r, err := BuildHttpResponse(resp, reqw.ResponseHandler)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	response.Data = *r
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

func BuildHttpRequest(req HttpRequest) (*http.Request, error) {
	httpreq, err := http.NewRequest(req.Method, req.Url, bytes.NewBuffer(req.Data))
	if err != nil {
		return nil, err
	}

	for key, values := range req.Header {
		for _, value := range values {
			httpreq.Header.Add(key, value)
		}
	}
	// TODO multipart uploads
	return httpreq, nil
}

func BuildHttpResponse(resp *http.Response, resph ResponseHandler) (*HttpResponse, error) {
	response := &HttpResponse{
		Status:        resp.Status,
		StatusCode:    resp.StatusCode,
		ContentLength: resp.ContentLength,
		Uncompressed:  resp.Uncompressed,
		Header:        resp.Header,
	}

	if resph.FilePath != "" {
		// TODO download to file & set the apropriate extension based on Content-Type
		return nil, fmt.Errorf("not implemented")
	}
	if resph.MaxSize > 0 && resph.MaxSize < resp.ContentLength {
		return nil, fmt.Errorf("http response body exceeds max length")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	response.Data = body
	return response, nil
}

func BuildWasmxHttpClient(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("Request", Request, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "httpclient", context, fndefs)
}
