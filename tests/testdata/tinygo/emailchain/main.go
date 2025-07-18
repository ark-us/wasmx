package main

import (
	"encoding/json"
	"fmt"

	wasmx "github.com/loredanacirstea/wasmx-env"
	_ "github.com/loredanacirstea/wasmx-env-httpclient"
	vmimap "github.com/loredanacirstea/wasmx-env-imap"
	vmsmtp "github.com/loredanacirstea/wasmx-env-smtp"
)

//go:wasm-module emailchain
//export instantiate
func Instantiate() {
	InitializeTables(ConnectionId)
}

func main() {
	fmt.Println("---main----!!!!!!")
	databz := wasmx.GetCallData()
	fmt.Println("---databz----!!!!!!", string(databz))
	calld := &Calldata{}
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
		resp := SignDKIM(calld.SignDKIM)
		response, _ = json.Marshal(&resp)
	} else if calld.VerifyDKIM != nil {
		resp := VerifyDKIM(calld.VerifyDKIM)
		response, _ = json.Marshal(&resp)
	} else if calld.VerifyARC != nil {
		resp := VerifyARC(calld.VerifyARC)
		response, _ = json.Marshal(&resp)
	} else if calld.SignARC != nil {
		resp := SignARC(calld.SignARC)
		response, _ = json.Marshal(&resp)
	} else if calld.ForwardEmail != nil {
		resp := ForwardEmail(calld.ForwardEmail)
		response, _ = json.Marshal(&resp)
	} else if calld.CreateAccount != nil {
		CreateAccount(calld.CreateAccount)
	} else if calld.SendEmail != nil {
		resp := SendEmail(calld.SendEmail)
		response, _ = json.Marshal(&resp)
	} else if calld.StartServer != nil {
		StartServer(calld.StartServer)
	} else if calld.IncomingEmail != nil {
		IncomingEmail(calld.IncomingEmail)
	} else if calld.RoleChanged != nil {
		// TODO
		// utils.OnlyRole(MODULE_NAME, roles.ROLE_ROLES, "RoleChanged")
		// roleChanged(calld.RoleChanged);
		InitializeTables(ConnectionId)
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

//go:wasm-module emailchain
//export smtp_update
func SmtpUpdate() {
	fmt.Println("---SmtpUpdate----!!!!!!")
	databz := wasmx.GetCallData()
	calld := &ReentryCalldata{}
	err := json.Unmarshal(databz, calld)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	if calld.IncomingEmail != nil {
		IncomingEmail(calld.IncomingEmail)
	}
}

type Response struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

func prepareResponse(data []byte, err error) []byte {
	resp := &Response{Data: data}
	if err != nil {
		resp.Error = err.Error()
	}
	bz, _ := json.Marshal(resp)
	return bz
}

func ImapServerRequest(calld *Calldata) bool {
	// var res interface{}
	var res []byte

	switch {
	case calld.Login != nil:
		data, err := HandleLogin(calld.Login)
		res = prepareResponse(data, err)
	case calld.Logout != nil:
		data, err := HandleLogout(calld.Logout)
		res = prepareResponse(data, err)
	case calld.Create != nil:
		data, err := HandleCreate(calld.Create)
		res = prepareResponse(data, err)
	case calld.Delete != nil:
		data, err := HandleDelete(calld.Delete)
		res = prepareResponse(data, err)
	case calld.Rename != nil:
		data, err := HandleRename(calld.Rename)
		res = prepareResponse(data, err)
	case calld.Select != nil:
		data, err := HandleSelect(calld.Select)
		res = prepareResponse(data, err)
	case calld.List != nil:
		data, err := HandleList(calld.List)
		res = prepareResponse(data, err)
	case calld.Status != nil:
		data, err := HandleStatus(calld.Status)
		res = prepareResponse(data, err)
	case calld.Append != nil:
		data, err := HandleAppend(calld.Append)
		res = prepareResponse(data, err)
	case calld.Expunge != nil:
		data, err := HandleExpunge(calld.Expunge)
		res = prepareResponse(data, err)
	case calld.Search != nil:
		data, err := HandleSearch(calld.Search)
		res = prepareResponse(data, err)
	case calld.Fetch != nil:
		data, err := HandleFetch(calld.Fetch)
		res = prepareResponse(data, err)
	case calld.Store != nil:
		data, err := HandleStore(calld.Store)
		res = prepareResponse(data, err)
	case calld.Copy != nil:
		data, err := HandleCopy(calld.Copy)
		res = prepareResponse(data, err)
	default:
		return false
	}
	wasmx.SetFinishData(res)
	return true
}

func SmtpServerRequest(calld *Calldata) bool {
	var res []byte
	switch {
	case calld.Login != nil:
		data, err := HandleSmtpLogin(calld.Login)
		res = prepareResponse(data, err)
	case calld.Logout != nil:
		data, err := HandleSmtpLogout(calld.Logout)
		res = prepareResponse(data, err)
	default:
		return false
	}
	wasmx.SetFinishData(res)
	return true
}
