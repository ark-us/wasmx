package main

import (
	"encoding/json"
	"os"

	lib "github.com/loredanacirstea/emailchain/lib"
	_ "github.com/loredanacirstea/wasmx-env-httpclient/lib"
	vmimap "github.com/loredanacirstea/wasmx-env-imap/lib"
	vmsmtp "github.com/loredanacirstea/wasmx-env-smtp/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

//go:wasm-module wasmxcore
//export wasmx_nondeterministic_1
func Wasmx_nondeterministic_1() {}

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Memory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func Wasmx_env_i64_2() {}

//go:wasm-module smtp
//export wasmx_smtp_i64_1
func Wasmx_smtp_i64_1() {}

//go:wasm-module imap
//export wasmx_imap_i64_1
func Wasmx_imap_i64_1() {}

//go:wasm-module httpclient
//export wasmx_httpclient_i64_1
func Wasmx_httpclient_i64_1() {}

//go:wasm-module sql
//export wasmx_sql_i64_1
func Wasmx_sql_i64_1() {}

func main() {
	entrypoint := os.Getenv("ENTRY_POINT")
	// these entrypoints are internal by default
	switch entrypoint {
	case "smtp_update":
		SmtpUpdate()
		return
	case "instantiate":
		lib.InitializeTables(lib.ConnectionId)
		return
	}

	databz := wasmx.GetCallData()
	calld := &lib.Calldata{}
	err := json.Unmarshal(databz, calld)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	response := []byte{}

	if calld.Connect != nil {
		resp := vmimap.Connect(&vmimap.ImapConnectionRequest{
			Id:            calld.Connect.Id,
			ImapServerUrl: calld.Connect.ImapServerUrl,
			Auth:          vmimap.ConnectionAuth(*calld.Connect.SmtpRequest.Auth),
		})
		if resp.Error == "" {
			resp2 := vmsmtp.ClientConnect(&calld.Connect.SmtpRequest)
			if resp2.Error != "" {
				resp.Error = resp2.Error
			}
		}
		response, _ = json.Marshal(&resp)
	} else if calld.Close != nil {
		resp := vmimap.Close(&vmimap.ImapCloseRequest{Id: calld.Close.Id})
		resp2 := vmsmtp.Close(&vmsmtp.SmtpCloseRequest{Id: calld.Close.Id})
		if resp2.Error != "" {
			resp.Error = resp2.Error
		}
		response, _ = json.Marshal(&resp)
	} else if calld.SignDKIM != nil {
		resp := lib.SignDKIM(calld.SignDKIM)
		response, _ = json.Marshal(&resp)
	} else if calld.VerifyDKIM != nil {
		resp := lib.VerifyDKIM(calld.VerifyDKIM)
		response, _ = json.Marshal(&resp)
	} else if calld.VerifyARC != nil {
		resp := lib.VerifyARC(calld.VerifyARC)
		response, _ = json.Marshal(&resp)
	} else if calld.SignARC != nil {
		resp := lib.SignARC(calld.SignARC)
		response, _ = json.Marshal(&resp)
	} else if calld.ForwardEmail != nil {
		resp := lib.ForwardEmail(calld.ForwardEmail)
		response, _ = json.Marshal(&resp)
	} else if calld.CreateAccount != nil {
		lib.CreateAccount(calld.CreateAccount)
	} else if calld.SendEmail != nil {
		resp := lib.SendEmail(calld.SendEmail)
		response, _ = json.Marshal(&resp)
	} else if calld.BuildAndSend != nil {
		resp := lib.BuildAndSend(calld.BuildAndSend)
		response, _ = json.Marshal(&resp)
	} else if calld.StartServer != nil {
		lib.StartServer(calld.StartServer)
	} else if calld.IncomingEmail != nil {
		lib.IncomingEmail(calld.IncomingEmail)
	} else if calld.RoleChanged != nil {
		wasmx.OnlyRole(lib.MODULE_NAME, wasmx.ROLE_ROLES, "RoleChanged")
		lib.InitializeTables(lib.ConnectionId)
	} else {
		handled := ImapServerRequest(calld)
		if handled {
			return
		}
		handled = SmtpServerRequest(calld)
		if handled {
			return
		}
		wasmx.Revert([]byte(`invalid function call data: ` + string(databz)))
	}
	wasmx.SetFinishData(response)
}

func SmtpUpdate() {
	databz := wasmx.GetCallData()
	calld := &lib.ReentryCalldata{}
	err := json.Unmarshal(databz, calld)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	if calld.IncomingEmail != nil {
		lib.IncomingEmail(calld.IncomingEmail)
	}
}

type Response struct {
	ImapError *vmimap.Error `json:"imap_error"`
	Error     string        `json:"error"`
	Data      []byte        `json:"data"`
}

func prepareResponse(data []byte, ierr *vmimap.Error, err error) []byte {
	resp := &Response{Data: data, ImapError: ierr}
	if err != nil {
		resp.Error = err.Error()
	}
	bz, _ := json.Marshal(resp)
	return bz
}

func ImapServerRequest(calld *lib.Calldata) bool {
	// var res interface{}
	var res []byte

	switch {
	case calld.Login != nil:
		data, err := lib.HandleLogin(calld.Login)
		res = prepareResponse(data, nil, err)
	case calld.Logout != nil:
		data, err := lib.HandleLogout(calld.Logout)
		res = prepareResponse(data, nil, err)
	case calld.Create != nil:
		data, ierr, err := lib.HandleCreate(calld.Create)
		res = prepareResponse(data, ierr, err)
	case calld.Delete != nil:
		data, err := lib.HandleDelete(calld.Delete)
		res = prepareResponse(data, nil, err)
	case calld.Rename != nil:
		data, err := lib.HandleRename(calld.Rename)
		res = prepareResponse(data, nil, err)
	case calld.Select != nil:
		data, err := lib.HandleSelect(calld.Select)
		res = prepareResponse(data, nil, err)
	case calld.List != nil:
		data, err := lib.HandleList(calld.List)
		res = prepareResponse(data, nil, err)
	case calld.Status != nil:
		data, err := lib.HandleStatus(calld.Status)
		res = prepareResponse(data, nil, err)
	case calld.Append != nil:
		data, err := lib.HandleAppend(calld.Append)
		res = prepareResponse(data, nil, err)
	case calld.Expunge != nil:
		data, err := lib.HandleExpunge(calld.Expunge)
		res = prepareResponse(data, nil, err)
	case calld.Search != nil:
		data, err := lib.HandleSearch(calld.Search)
		res = prepareResponse(data, nil, err)
	case calld.Fetch != nil:
		data, err := lib.HandleFetch(calld.Fetch)
		res = prepareResponse(data, nil, err)
	case calld.Store != nil:
		data, err := lib.HandleStore(calld.Store)
		res = prepareResponse(data, nil, err)
	case calld.Copy != nil:
		data, err := lib.HandleCopy(calld.Copy)
		res = prepareResponse(data, nil, err)
	default:
		return false
	}
	wasmx.SetFinishData(res)
	return true
}

func SmtpServerRequest(calld *lib.Calldata) bool {
	var res []byte
	switch {
	case calld.Login != nil:
		data, err := lib.HandleSmtpLogin(calld.Login)
		res = prepareResponse(data, nil, err)
	case calld.Logout != nil:
		data, err := lib.HandleSmtpLogout(calld.Logout)
		res = prepareResponse(data, nil, err)
	default:
		return false
	}
	wasmx.SetFinishData(res)
	return true
}
