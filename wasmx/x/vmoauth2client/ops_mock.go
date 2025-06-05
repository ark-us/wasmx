package vmoauth2client

import (
	vmhttpclient "github.com/loredanacirstea/wasmx/x/vmhttpclient"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func GetRedirectUrlMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &GetRedirectUrlResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func ExchangeCodeForTokenMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &ExchangeCodeForTokenResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func RefreshTokenMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &RefreshTokenResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func Oauth2ClientGetMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &Oauth2ClientGetResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func Oauth2ClientDoMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &vmhttpclient.HttpResponseWrap{Error: ""}
	return prepareResponse(rnh, response)
}

func Oauth2ClientPostMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &vmhttpclient.HttpResponseWrap{Error: ""}
	return prepareResponse(rnh, response)
}
