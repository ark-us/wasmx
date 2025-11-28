package httpserver

// #include <stdlib.h>
import "C"

import (
	"encoding/json"

	utils "github.com/loredanacirstea/wasmx-env-utils"
)

//go:wasmimport httpserver StartWebServer
func StartWebServer_(reqPtr int64) int64

//go:wasmimport httpserver SetRouteHandler
func SetRouteHandler_(reqPtr int64) int64

//go:wasmimport httpserver RemoveRouteHandler
func RemoveRouteHandler_(reqPtr int64) int64

//go:wasmimport httpserver Close
func Close_(reqPtr int64) int64

//go:wasmimport httpserver GenerateJWT
func GenerateJWT_(reqPtr int64) int64

//go:wasmimport httpserver VerifyJWT
func VerifyJWT_(reqPtr int64) int64

func StartWebServer(req *StartWebServerRequest) StartWebServerResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := StartWebServer_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp StartWebServerResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func SetRouteHandler(req *SetRouteHandlerRequest) SetRouteHandlerResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := SetRouteHandler_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp SetRouteHandlerResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func RemoveRouteHandler(req *RemoveRouteHandlerRequest) RemoveRouteHandlerResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := RemoveRouteHandler_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp RemoveRouteHandlerResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Close(req *CloseRequest) CloseResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Close_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp CloseResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func GenerateJWT(req *GenerateJWTRequest) GenerateJWTResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := GenerateJWT_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp GenerateJWTResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func VerifyJWT(req *VerifyJWTRequest) VerifyJWTResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := VerifyJWT_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp VerifyJWTResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}
