package vmoauth2client

import (
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/oauth2"

	vmhttpclient "github.com/loredanacirstea/wasmx/x/vmhttpclient"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
)

const (
	// ModuleName defines the module name
	ModuleName = "vmoauth2client"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

const HOST_WASMX_ENV_OAUTH2CLIENT_EXPORT = "wasmx_oauth2client_"

const HOST_WASMX_ENV_OAUTH2CLIENT_i32_VER1 = "wasmx_oauth2client_i32_1"
const HOST_WASMX_ENV_OAUTH2CLIENT_i64_VER1 = "wasmx_oauth2client_i64_1"

const HOST_WASMX_ENV_OAUTH2CLIENT = "oauth2client"

type ContextKey string

const OAuth2ClientContextKey ContextKey = "oauth2client-context"

type Context struct {
	*vmtypes.Context
}

type OAuth2ClientContext struct {
	mtx     sync.Mutex
	Clients map[string]*http.Client
}

func (p *OAuth2ClientContext) GetClient(id string) (*http.Client, bool) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	db, found := p.Clients[id]
	return db, found
}

func (p *OAuth2ClientContext) SetClient(id string, client *http.Client) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	_, found := p.Clients[id]
	if found {
		return fmt.Errorf("cannot overwrite client connection: %s", id)
	}
	p.Clients[id] = client
	return nil
}

func (p *OAuth2ClientContext) DeleteClient(id string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	delete(p.Clients, id)
}

type AuthUrlParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Endpoint struct {
	AuthURL       string           `json:"auth_url"`
	DeviceAuthURL string           `json:"device_auth_url"`
	TokenURL      string           `json:"token_url"`
	AuthStyle     oauth2.AuthStyle `json:"auth_style"`
}

type OAuth2Config struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Endpoint     Endpoint `json:"endpoint"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
}

func (v OAuth2Config) toConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     v.ClientID,
		ClientSecret: v.ClientSecret,
		RedirectURL:  v.RedirectURL,
		Scopes:       v.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:       v.Endpoint.AuthURL,
			DeviceAuthURL: v.Endpoint.DeviceAuthURL,
			TokenURL:      v.Endpoint.TokenURL,
			AuthStyle:     v.Endpoint.AuthStyle,
		},
	}
}

type GetRedirectUrlRequest struct {
	Config        OAuth2Config   `json:"config"`
	RandomState   string         `json:"random_state"`
	AuthUrlParams []AuthUrlParam `json:"auth_url_params"`
}

func (v GetRedirectUrlRequest) Validate() error {
	if v.RandomState == "" {
		return fmt.Errorf("empty random state")
	}
	return nil
}

type GetRedirectUrlResponse struct {
	Error string `json:"error"`
	Url   string `json:"url"`
}

type ExchangeCodeForTokenRequest struct {
	Config            OAuth2Config `json:"config"`
	AuthorizationCode string       `json:"authorization_code"`
}

func (v ExchangeCodeForTokenRequest) Validate() error {
	return nil
}

type ExchangeCodeForTokenResponse struct {
	Error string        `json:"error"`
	Token *oauth2.Token `json:"token"`
}

type RefreshTokenRequest struct {
	Config       OAuth2Config `json:"config"`
	RefreshToken string       `json:"refresh_token"`
}

func (v RefreshTokenRequest) Validate() error {
	if v.RefreshToken == "" {
		return fmt.Errorf("empty refresh token")
	}
	return nil
}

type RefreshTokenResponse struct {
	Error string        `json:"error"`
	Token *oauth2.Token `json:"token"`
}

type Oauth2ClientConnectResponse struct {
	Error string `json:"error"`
}

type Oauth2ClientGetRequest struct {
	Config     OAuth2Config  `json:"config"`
	Token      *oauth2.Token `json:"token"`
	RequestUri string        `json:"request_uri"`
}

func (v Oauth2ClientGetRequest) Validate() error {
	if v.Token == nil {
		return fmt.Errorf("empty token")
	}
	if v.RequestUri == "" {
		return fmt.Errorf("empty request uri")
	}
	return nil
}

type Oauth2ClientGetResponse struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

type Oauth2ClientDoRequest struct {
	Config          OAuth2Config                 `json:"config"`
	Token           *oauth2.Token                `json:"token"`
	Request         vmhttpclient.HttpRequest     `json:"request"`
	ResponseHandler vmhttpclient.ResponseHandler `json:"response_handler"`
}

func (v Oauth2ClientDoRequest) Validate() error {
	if v.Token == nil {
		return fmt.Errorf("empty token")
	}
	return nil
}

type Oauth2ClientPostRequest struct {
	Config          OAuth2Config                 `json:"config"`
	Token           *oauth2.Token                `json:"token"`
	RequestUri      string                       `json:"request_uri"`
	ContentType     string                       `json:"content_type"`
	Data            []byte                       `json:"data"`
	ResponseHandler vmhttpclient.ResponseHandler `json:"response_handler"`
}

func (v Oauth2ClientPostRequest) Validate() error {
	if v.Token == nil {
		return fmt.Errorf("empty token")
	}
	if v.RequestUri == "" {
		return fmt.Errorf("empty request_uri")
	}
	if v.ContentType == "" {
		return fmt.Errorf("empty content_type")
	}
	return nil
}
