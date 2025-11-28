package oauth2client

import (
	httpclient "github.com/loredanacirstea/wasmx-env-httpclient/lib"
)

type AuthUrlParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Endpoint struct {
	AuthURL       string `json:"auth_url"`
	DeviceAuthURL string `json:"device_auth_url"`
	TokenURL      string `json:"token_url"`
	AuthStyle     int    `json:"auth_style"`
}

type OAuth2Config struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Endpoint     Endpoint `json:"endpoint"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
}

type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Expiry       int64  `json:"expiry"`
}

type GetRedirectUrlRequest struct {
	Config        OAuth2Config   `json:"config"`
	RandomState   string         `json:"random_state"`
	AuthUrlParams []AuthUrlParam `json:"auth_url_params"`
}

type GetRedirectUrlResponse struct {
	Error string `json:"error"`
	Url   string `json:"url"`
}

type ExchangeCodeForTokenRequest struct {
	Config            OAuth2Config `json:"config"`
	AuthorizationCode string       `json:"authorization_code"`
}

type ExchangeCodeForTokenResponse struct {
	Error string `json:"error"`
	Token *Token `json:"token"`
}

type RefreshTokenRequest struct {
	Config       OAuth2Config `json:"config"`
	RefreshToken string       `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	Error string `json:"error"`
	Token *Token `json:"token"`
}

type Oauth2ClientGetRequest struct {
	Config     OAuth2Config `json:"config"`
	Token      *Token       `json:"token"`
	RequestUri string       `json:"request_uri"`
}

type Oauth2ClientGetResponse struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

type Oauth2ClientDoRequest struct {
	Config          OAuth2Config                `json:"config"`
	Token           *Token                      `json:"token"`
	Request         httpclient.HttpRequest      `json:"request"`
	ResponseHandler httpclient.ResponseHandler  `json:"response_handler"`
}

type Oauth2ClientPostRequest struct {
	Config          OAuth2Config                `json:"config"`
	Token           *Token                      `json:"token"`
	RequestUri      string                      `json:"request_uri"`
	ContentType     string                      `json:"content_type"`
	Data            []byte                      `json:"data"`
	ResponseHandler httpclient.ResponseHandler  `json:"response_handler"`
}
