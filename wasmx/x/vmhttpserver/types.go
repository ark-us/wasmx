package vmhttpserver

import (
	"net/http"

	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
)

const (
	// ModuleName defines the module name
	ModuleName = "vmhttpserver"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

const HOST_WASMX_ENV_HTTP_VER1 = "wasmx_httpserver_1"

const HOST_WASMX_ENV_HTTP_EXPORT = "wasmx_httpserver_"

const HOST_WASMX_ENV_HTTP = "httpserver"

const ENTRY_POINT_HTTP_SERVER = "http_request_incoming"

const ROLE = "http_server_registry"

type ContextKey string

const HttpServerContextKey ContextKey = "httpserver-context"

var DefaultCORSAllowedOrigins = []string{"*"}
var DefaultCORSAllowedMethods = []string{"GET"}
var DefaultCORSAllowedHeaders = []string{"Origin", "Accept", "Content-Type", "X-Requested-With", "X-Server-Time"}

type Context struct {
	*vmtypes.Context
}

type HttpServerContext struct {
	Server            *http.Server
	WebsrvServer      *WebsrvServer
	ServerChannelDone chan struct{}
}

type StartWebServerRequest struct {
	Config WebsrvConfig `json:"config"`
}

type StartWebServerResponse struct {
	Error string `json:"error"`
}

type SetRouteHandlerRequest struct {
	Route           string `json:"route"`
	ContractAddress string `json:"contract_address"`
}

type SetRouteHandlerResponse struct {
	Error string `json:"error"`
}

type RemoveRouteHandlerRequest struct {
	Route string `json:"route"`
}

type RemoveRouteHandlerResponse struct {
	Error string `json:"error"`
}

type CloseRequest struct{}

type CloseResponse struct {
	Error string `json:"error"`
}

type HttpRequestIncoming struct {
	Method        string      `json:"method"`
	Url           string      `json:"url"`
	Header        http.Header `json:"header"`
	ContentLength int64       `json:"content_length"`
	Data          []byte      `json:"data"`
	RemoteAddr    string      `json:"remote_address"`
	RequestURI    string      `json:"request_uri"`
	// TODO
	// MultipartForm *multipart.Form
	// TLS *tls.ConnectionState
}

// WebsrvConfig defines the application configuration values for websrv module.
type WebsrvConfig struct {
	EnableOAuth        bool     `json:"enable_oauth"`
	Address            string   `json:"address"`
	CORSAllowedOrigins []string `json:"cors_allowed_origins"`
	CORSAllowedMethods []string `json:"cors_allowed_methods"`
	CORSAllowedHeaders []string `json:"cors_allowed_headers"`
	// MaxOpenConnections sets the maximum number of simultaneous connections
	// for the server listener.
	MaxOpenConnections     int               `json:"max_open_connections"`
	RouteToContractAddress map[string]string `json:"route_to_contract_address"`
	RequestBodyMaxSize     int64             `json:"request_body_max_size"`
}

func (c WebsrvConfig) Validate() error {
	// TODO
	return nil
}

// IsCorsEnabled returns true if cross-origin resource sharing is enabled.
func (c *WebsrvConfig) IsCorsEnabled() bool {
	return len(c.CORSAllowedOrigins) != 0
}

type HttpResponse struct {
	Status      string      `json:"status"`
	StatusCode  int         `json:"status_code"`
	Header      http.Header `json:"header"`
	Data        []byte      `json:"data"`
	RedirectUrl string      `json:"redirect_url"`
	// TODO: what other http response aside from redirect?
}

type HttpResponseWrap struct {
	Error string       `json:"error"`
	Data  HttpResponse `json:"data"`
}
