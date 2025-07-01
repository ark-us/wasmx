package main

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env"
	_ "github.com/loredanacirstea/wasmx-env-httpclient"
	vmimap "github.com/loredanacirstea/wasmx-env-imap"
	vmsmtp "github.com/loredanacirstea/wasmx-env-smtp"
)

//go:wasm-module emailprover
//export instantiate
func Instantiate() {}

func main() {
	databz := wasmx.GetCallData()
	calld := &Calldata{}
	err := json.Unmarshal(databz, calld)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	response := []byte{}

	if calld.ConnectWithPassword != nil {
		resp := vmimap.ConnectWithPassword(&vmimap.ImapConnectionSimpleRequest{
			Id:            calld.ConnectWithPassword.Id,
			ImapServerUrl: calld.ConnectWithPassword.ImapServerUrl,
			Username:      calld.ConnectWithPassword.Username,
			Password:      calld.ConnectWithPassword.Password,
		})
		if resp.Error == "" {
			resp2 := vmsmtp.ConnectWithPassword(&vmsmtp.SmtpConnectionSimpleRequest{
				Id:                    calld.ConnectWithPassword.Id,
				SmtpServerUrlSTARTTLS: calld.ConnectWithPassword.SmtpServerUrlSTARTTLS,
				SmtpServerUrlTLS:      calld.ConnectWithPassword.SmtpServerUrlTLS,
				Username:              calld.ConnectWithPassword.Username,
				Password:              calld.ConnectWithPassword.Password,
			})
			if resp2.Error != "" {
				resp.Error = resp2.Error
			}
		}
		response, _ = json.Marshal(&resp)
	} else if calld.ConnectOAuth2 != nil {
		resp := vmimap.ConnectOAuth2(&vmimap.ImapConnectionOauth2Request{
			Id:            calld.ConnectOAuth2.Id,
			ImapServerUrl: calld.ConnectOAuth2.ImapServerUrl,
			Username:      calld.ConnectOAuth2.Username,
			AccessToken:   calld.ConnectOAuth2.AccessToken,
		})
		if resp.Error == "" {
			resp2 := vmsmtp.ConnectOAuth2(&vmsmtp.SmtpConnectionOauth2Request{
				Id:                    calld.ConnectOAuth2.Id,
				SmtpServerUrlSTARTTLS: calld.ConnectOAuth2.SmtpServerUrlSTARTTLS,
				SmtpServerUrlTLS:      calld.ConnectOAuth2.SmtpServerUrlTLS,
				Username:              calld.ConnectOAuth2.Username,
				AccessToken:           calld.ConnectOAuth2.AccessToken,
			})
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
	} else if calld.StartServer != nil {
		StartServer()
	} else {
		wasmx.Revert([]byte(`invalid function call data: ` + string(databz)))
	}
	wasmx.SetFinishData(response)
}
