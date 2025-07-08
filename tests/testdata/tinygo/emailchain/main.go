package main

import (
	"encoding/json"

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
	databz := wasmx.GetCallData()
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
	} else if calld.SendEmail != nil {
		resp := SendEmail(calld.SendEmail)
		response, _ = json.Marshal(&resp)
	} else if calld.StartServer != nil {
		StartServer()
	} else if calld.IncomingEmail != nil {
		IncomingEmail(calld.IncomingEmail)
	} else if calld.RoleChanged != nil {
		// TODO
		// utils.OnlyRole(MODULE_NAME, roles.ROLE_ROLES, "RoleChanged")
		// roleChanged(calld.RoleChanged);
		InitializeTables(ConnectionId)
	} else {
		wasmx.Revert([]byte(`invalid function call data: ` + string(databz)))
	}
	wasmx.SetFinishData(response)
}

//go:wasm-module emailprover
//export smtp_update
func SmtpUpdate() {
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
