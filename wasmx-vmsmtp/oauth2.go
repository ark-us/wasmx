package vmsmtp

import "fmt"

type OAuth2Authenticator struct {
	accessToken string
	username    string
}

func (a *OAuth2Authenticator) Start() (string, []byte, error) {
	resp := []byte(fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.username, a.accessToken))
	return "XOAUTH2", resp, nil
}

func (a *OAuth2Authenticator) Next(challenge []byte) ([]byte, error) {
	return nil, nil
}
