package imap

// #include <stdlib.h>
import "C"

import (
	"encoding/json"

	utils "github.com/loredanacirstea/wasmx-env-utils"
)

//go:wasm-module imap
//export wasmx_imap_i64_1
func wasmx_imap_i64_1() {}

//go:wasmimport imap Connect
func Connect_(reqPtr int64) int64

//go:wasmimport imap Close
func Close_(reqPtr int64) int64

//go:wasmimport imap Listen
func Listen_(reqPtr int64) int64

//go:wasmimport imap Count
func Count_(reqPtr int64) int64

//go:wasmimport imap UIDSearch
func UIDSearch_(reqPtr int64) int64

//go:wasmimport imap ListMailboxes
func ListMailboxes_(reqPtr int64) int64

//go:wasmimport imap Fetch
func Fetch_(reqPtr int64) int64

//go:wasmimport imap CreateFolder
func CreateFolder_(reqPtr int64) int64

//go:wasmimport imap ServerStart
func ServerStart_(reqPtr int64) int64

//go:wasmimport imap ServerClose
func ServerClose_(reqPtr int64) int64

func Connect(req *ImapConnectionRequest) ImapConnectionResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Connect_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp ImapConnectionResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Close(req *ImapCloseRequest) ImapCloseResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Close_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp ImapCloseResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Listen(req *ImapListenRequest) ImapListenResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Listen_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp ImapListenResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Count(req *ImapCountRequest) ImapCountResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Count_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp ImapCountResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func UIDSearch(req *ImapUIDSearchRequest) ImapUIDSearchResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := UIDSearch_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp ImapUIDSearchResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func ListMailboxes(req *ListMailboxesRequest) ListMailboxesResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := ListMailboxes_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp ListMailboxesResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Fetch(req *ImapFetchRequest) ImapFetchResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Fetch_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp ImapFetchResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func CreateFolder(req *ImapCreateFolderRequest) ImapCreateFolderResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := CreateFolder_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp ImapCreateFolderResponse
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
