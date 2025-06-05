package main

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env"
	vmimap "github.com/loredanacirstea/wasmx-env-imap"
	vmsmtp "github.com/loredanacirstea/wasmx-env-smtp"
)

//go:wasm-module smtptest
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
		resp := vmsmtp.ConnectWithPassword(calld.ConnectWithPassword)
		response, err = json.Marshal(&resp)
	} else if calld.ConnectOAuth2 != nil {
		resp := vmsmtp.ConnectOAuth2(calld.ConnectOAuth2)
		response, err = json.Marshal(&resp)
	} else if calld.Close != nil {
		resp := vmsmtp.Close(calld.Close)
		response, err = json.Marshal(&resp)
	} else if calld.Quit != nil {
		resp := vmsmtp.Quit(calld.Quit)
		response, err = json.Marshal(&resp)
	} else if calld.Extension != nil {
		resp := vmsmtp.Extension(calld.Extension)
		response, err = json.Marshal(&resp)
	} else if calld.Noop != nil {
		resp := vmsmtp.Noop(calld.Noop)
		response, err = json.Marshal(&resp)
	} else if calld.SendMail != nil {
		resp := vmsmtp.SendMail(calld.SendMail)
		response, err = json.Marshal(&resp)
	} else if calld.Verify != nil {
		resp := vmsmtp.Verify(calld.Verify)
		response, err = json.Marshal(&resp)
	} else if calld.SupportsAuth != nil {
		resp := vmsmtp.SupportsAuth(calld.SupportsAuth)
		response, err = json.Marshal(&resp)
	} else if calld.MaxMessageSize != nil {
		resp := vmsmtp.MaxMessageSize(calld.MaxMessageSize)
		response, err = json.Marshal(&resp)
	} else if calld.BuildAndSend != nil {
		resp := vmsmtp.SendMail(BuildEmail(calld.BuildAndSend))
		response, err = json.Marshal(&resp)
	} else {
		wasmx.Revert([]byte(`invalid function call data: ` + string(databz)))
	}
	wasmx.SetFinishData(response)
}

func BuildEmail(req *BuildAndSendMailRequest) *vmsmtp.SmtpSendMailRequest {
	sendReq, err := BuildEmailInternal(req)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	return sendReq
}

func BuildEmailInternal(req *BuildAndSendMailRequest) (*vmsmtp.SmtpSendMailRequest, error) {
	to := make([]vmimap.Address, len(req.To))
	for i, acc := range req.To {
		to[i] = vmimap.AddressFromString(acc, "")
	}
	header := make(map[string][]string, 0)
	header["X-PROVABLE-ID"] = []string{"myuniqueprovableid"}

	email := vmimap.Email{
		Envelope: &vmimap.Envelope{
			Subject: req.Subject,
			From: []vmimap.Address{
				vmimap.AddressFromString(req.From, ""),
			},
			To: to,
		},
		Header:      header,
		Body:        string(req.Body),
		Attachments: []vmimap.Attachment{},
	}
	emailstr, err := BuildRawEmail(email)
	if err != nil {
		return nil, err
	}
	return &vmsmtp.SmtpSendMailRequest{
		Id:    req.Id,
		From:  req.From,
		To:    req.To,
		Email: []byte(emailstr),
	}, nil
}
