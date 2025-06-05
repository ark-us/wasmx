package smtp

// #include <stdlib.h>
import "C"

import (
	"encoding/json"

	utils "github.com/loredanacirstea/wasmx-utils"
)

//go:wasm-module smtp
//export wasmx_smtp_i64_1
func wasmx_smtp_i64_1() {}

//go:wasmimport smtp ConnectWithPassword
func ConnectWithPassword_(reqPtr int64) int64

//go:wasmimport smtp ConnectOAuth2
func ConnectOAuth2_(reqPtr int64) int64

//go:wasmimport smtp Close
func Close_(reqPtr int64) int64

//go:wasmimport smtp Quit
func Quit_(reqPtr int64) int64

//go:wasmimport smtp Extension
func Extension_(reqPtr int64) int64

//go:wasmimport smtp Noop
func Noop_(reqPtr int64) int64

//go:wasmimport smtp SendMail
func SendMail_(reqPtr int64) int64

//go:wasmimport smtp Verify
func Verify_(reqPtr int64) int64

//go:wasmimport smtp SupportsAuth
func SupportsAuth_(reqPtr int64) int64

//go:wasmimport smtp MaxMessageSize
func MaxMessageSize_(reqPtr int64) int64

//go:wasmimport smtp BuildMail
func BuildMail_(reqPtr int64) int64

func ConnectWithPassword(req *SmtpConnectionSimpleRequest) SmtpConnectionResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := ConnectWithPassword_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp SmtpConnectionResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func ConnectOAuth2(req *SmtpConnectionOauth2Request) SmtpConnectionResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := ConnectOAuth2_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp SmtpConnectionResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Close(req *SmtpCloseRequest) SmtpCloseResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Close_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp SmtpCloseResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Quit(req *SmtpQuitRequest) SmtpQuitResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Quit_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp SmtpQuitResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Extension(req *SmtpExtensionRequest) SmtpExtensionResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Extension_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp SmtpExtensionResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Noop(req *SmtpNoopRequest) SmtpNoopResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Noop_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp SmtpNoopResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Verify(req *SmtpVerifyRequest) SmtpVerifyResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Verify_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp SmtpVerifyResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func SupportsAuth(req *SmtpSupportsAuthRequest) SmtpSupportsAuthResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := SupportsAuth_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp SmtpSupportsAuthResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func MaxMessageSize(req *SmtpMaxMessageSizeRequest) SmtpMaxMessageSizeResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := MaxMessageSize_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp SmtpMaxMessageSizeResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func SendMail(req *SmtpSendMailRequest) SmtpSendMailResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := SendMail_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp SmtpSendMailResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}
