package lib

import (
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func HandleSmtpLogin(req *LoginRequest) ([]byte, error) {
	// Connect to database to check if username exists
	err := ConnectSql(ConnectionId)
	if err != nil {
		wasmx.Revert([]byte("HandleSmtpLogin: DB connection failed: " + err.Error()))
	}

	// Validate username exists in database
	owners, err := GetAccount(req.Username)
	if err != nil {
		wasmx.Revert([]byte("HandleSmtpLogin: failed to check account: " + err.Error()))
	}
	if len(owners) == 0 {
		wasmx.Revert([]byte("unauthorized: invalid username"))
	}

	// Validate password against stored server password
	storedPassword := LoadServerPassword()
	if storedPassword == "" {
		wasmx.Revert([]byte("unauthorized: server password not configured"))
	}
	if req.Password != storedPassword {
		wasmx.Revert([]byte("unauthorized: invalid password"))
	}

	return nil, nil
}

func HandleSmtpLogout(req *LogoutRequest) ([]byte, error) {
	return nil, nil
}
