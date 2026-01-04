package httpserver

import "net/http"

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
}

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
	// StaticRoutes maps route prefixes to filesystem folder paths
	// e.g. "/static" -> "/var/www/static"
	StaticRoutes map[string]string `json:"static_routes"`
}

type HttpResponse struct {
	Status      string      `json:"status"`
	StatusCode  int         `json:"status_code"`
	Header      http.Header `json:"header"`
	Data        []byte      `json:"data"`
	RedirectUrl string      `json:"redirect_url"`
}

type HttpResponseWrap struct {
	Error string       `json:"error"`
	Data  HttpResponse `json:"data"`
}

type GenerateJWTRequest struct {
	Secret          []byte           `json:"secret"`
	Claims          RegisteredClaims `json:"claims"`
	AdditionalClaim string           `json:"additional_claim"`
	SigningMethod   string           `json:"signing_method"`
}

type GenerateJWTResponse struct {
	Token string `json:"token"`
	Error string `json:"error"`
}

type VerifyJWTRequest struct {
	Secret          []byte           `json:"secret"`
	Token           string           `json:"token"`
	Claims          RegisteredClaims `json:"claims"`
	AdditionalClaim string           `json:"additional_claim"`
}

type VerifyJWTResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error"`
}

type RegisteredClaims struct {
	Issuer    string   `json:"issuer"`
	Subject   string   `json:"subject"`
	Audience  []string `json:"audience"`
	ExpiresAt int64    `json:"expires_at"`
	NotBefore int64    `json:"not_before"`
	IssuedAt  int64    `json:"issued_at"`
	ID        string   `json:"id"`
}
