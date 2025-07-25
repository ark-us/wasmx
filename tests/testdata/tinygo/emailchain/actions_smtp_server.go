package main

import "github.com/loredanacirstea/wasmx-env"

// TODO implement me
func HandleSmtpLogin(req *LoginRequest) ([]byte, error) {
	if req.Password != "123456" {
		wasmx.Revert([]byte(`unauthorized`))
	}
	return nil, nil
}

func HandleSmtpLogout(req *LogoutRequest) ([]byte, error) {
	return nil, nil
}
