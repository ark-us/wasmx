package smtp

// #include <stdlib.h>
import "C"

import (
	"encoding/json"
	"fmt"

	utils "github.com/loredanacirstea/wasmx-utils"
)

//go:wasm-module smtp
//export wasmx_smtp_i64_1
func wasmx_smtp_i64_1() {}

//go:wasmimport smtp ClientConnect
func ClientConnect_(reqPtr int64) int64

//go:wasmimport smtp Close
func Close_(reqPtr int64) int64

//go:wasmimport smtp Quit
func Quit_(reqPtr int64) int64

//go:wasmimport smtp Extension
func Extension_(reqPtr int64) int64

//go:wasmimport smtp Noop
func Noop_(reqPtr int64) int64

//go:wasmimport smtp Hello
func Hello_(reqPtr int64) int64

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

//go:wasmimport smtp ServerStart
func ServerStart_(reqPtr int64) int64

//go:wasmimport smtp ServerClose
func ServerClose_(reqPtr int64) int64

//go:wasmimport smtp ServerShutdown
func ServerShutdown_(reqPtr int64) int64

func ClientConnect(req *SmtpConnectionRequest) SmtpConnectionResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := ClientConnect_(reqPtr)
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

func Hello(req *SmtpHelloRequest) SmtpHelloResponse {
	fmt.Println("--tinygo.Hello----")
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	fmt.Println("--tinygo.Hello----", string(reqbz))
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Hello_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp SmtpHelloResponse
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

func ServerStart(req *ServerStartRequest) ServerStartResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := ServerStart_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp ServerStartResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func ServerClose(req *ServerCloseRequest) ServerCloseResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := ServerClose_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp ServerCloseResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func ServerShutdown(req *ServerShutdownRequest) ServerShutdownResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := ServerShutdown_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp ServerShutdownResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}
