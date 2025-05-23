package vmoauth2client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	vmhttpclient "github.com/loredanacirstea/wasmx/x/vmhttpclient"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
	"golang.org/x/oauth2"
)

func GetRedirectUrl(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &GetRedirectUrlResponse{Error: ""}
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req GetRedirectUrlRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	if err = req.Validate(); err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	opts := make([]oauth2.AuthCodeOption, 0)
	for _, param := range req.AuthUrlParams {
		opts = append(opts, oauth2.SetAuthURLParam(param.Key, param.Value))
	}
	config := req.Config.toConfig()
	url := config.AuthCodeURL(req.RandomState, opts...)
	response.Url = url

	return prepareResponse(rnh, response)
}

func ExchangeCodeForToken(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &ExchangeCodeForTokenResponse{Error: ""}
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req ExchangeCodeForTokenRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	if err = req.Validate(); err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	config := req.Config.toConfig()
	token, err := config.Exchange(ctx.Ctx, req.AuthorizationCode)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	response.Token = token
	return prepareResponse(rnh, response)
}

func RefreshToken(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &RefreshTokenResponse{Error: ""}
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req RefreshTokenRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	if err = req.Validate(); err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	token := &oauth2.Token{
		RefreshToken: req.RefreshToken,
	}

	newToken, err := req.Config.toConfig().TokenSource(ctx.Ctx, token).Token()
	if err != nil {
		response.Error = fmt.Errorf("Failed to refresh token: %s", err).Error()
		return prepareResponse(rnh, response)
	}
	response.Token = newToken
	return prepareResponse(rnh, response)
}

func Oauth2ClientGet(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &Oauth2ClientGetResponse{Error: ""}
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req Oauth2ClientGetRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	if err = req.Validate(); err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	client := req.Config.toConfig().Client(ctx.Ctx, req.Token)
	resp, err := client.Get(req.RequestUri)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Error = "cannot read response"
		return prepareResponse(rnh, response)
	}
	response.Data = body
	return prepareResponse(rnh, response)
}

func Oauth2ClientDo(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &vmhttpclient.HttpResponseWrap{Error: ""}
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req Oauth2ClientDoRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	if err = req.Validate(); err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	client := req.Config.toConfig().Client(ctx.Ctx, req.Token)

	httpreq, err := vmhttpclient.BuildHttpRequest(req.Request)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	resp, err := client.Do(httpreq)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	r, err := vmhttpclient.BuildHttpResponse(resp, req.ResponseHandler)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	response.Data = *r
	return prepareResponse(rnh, response)
}

func Oauth2ClientPost(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &vmhttpclient.HttpResponseWrap{Error: ""}
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req Oauth2ClientPostRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	if err = req.Validate(); err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	client := req.Config.toConfig().Client(ctx.Ctx, req.Token)
	body := bytes.NewReader(req.Data)
	resp, err := client.Post(req.RequestUri, req.ContentType, body)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	r, err := vmhttpclient.BuildHttpResponse(resp, req.ResponseHandler)
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

func BuildWasmxOAuth2Client(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("GetRedirectUrl", GetRedirectUrl, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ExchangeCodeForToken", ExchangeCodeForToken, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("RefreshToken", RefreshToken, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Get", Oauth2ClientGet, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Do", Oauth2ClientDo, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Post", Oauth2ClientPost, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "oauth2client", context, fndefs)
}
