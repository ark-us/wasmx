package oauth2client

// #include <stdlib.h>
import "C"

import (
	"encoding/json"

	httpclient "github.com/loredanacirstea/wasmx-env-httpclient/lib"
	utils "github.com/loredanacirstea/wasmx-env-utils"
)

//go:wasmimport oauth2client GetRedirectUrl
func GetRedirectUrl_(reqPtr int64) int64

//go:wasmimport oauth2client ExchangeCodeForToken
func ExchangeCodeForToken_(reqPtr int64) int64

//go:wasmimport oauth2client RefreshToken
func RefreshToken_(reqPtr int64) int64

//go:wasmimport oauth2client Get
func Get_(reqPtr int64) int64

//go:wasmimport oauth2client Do
func Do_(reqPtr int64) int64

//go:wasmimport oauth2client Post
func Post_(reqPtr int64) int64

func GetRedirectUrl(req *GetRedirectUrlRequest) GetRedirectUrlResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := GetRedirectUrl_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp GetRedirectUrlResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func ExchangeCodeForToken(req *ExchangeCodeForTokenRequest) ExchangeCodeForTokenResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := ExchangeCodeForToken_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp ExchangeCodeForTokenResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func RefreshToken(req *RefreshTokenRequest) RefreshTokenResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := RefreshToken_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp RefreshTokenResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Get(req *Oauth2ClientGetRequest) Oauth2ClientGetResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Get_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp Oauth2ClientGetResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Do(req *Oauth2ClientDoRequest) httpclient.HttpResponseWrap {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Do_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp httpclient.HttpResponseWrap
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Post(req *Oauth2ClientPostRequest) httpclient.HttpResponseWrap {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Post_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp httpclient.HttpResponseWrap
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}
