package lib

import (
	wasmxhttp "github.com/loredanacirstea/wasmx-env-httpserver/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "httpserver-registry"

const (
	storageKeyConfig = "config"
	storageKeyRoutes = "routes"
)

// CallData defines all entrypoints
type CallData struct {
	ConnectDatabase *ConnectDatabaseRequest `json:"connect_database,omitempty"`
	SetRoute        *SetRouteRequest        `json:"set_route,omitempty"`
	SetRoutes       *SetRoutesRequest       `json:"set_routes,omitempty"`
	RemoveRoute     *RemoveRouteRequest     `json:"remove_route,omitempty"`
	GetRoute        *GetRouteRequest        `json:"get_route,omitempty"`
	GetRoutes       *GetRoutesRequest       `json:"get_routes,omitempty"`
	StartWebServer  *StartWebServerRequest  `json:"start_web_server,omitempty"`
	CloseServer     *CloseServerRequest     `json:"close_server,omitempty"`
	RoleChanged     *wasmx.RolesChangedHook `json:"RoleChanged,omitempty"`
}

// Requests / responses
type ConnectDatabaseRequest struct {
	// retained for compatibility; no-op when using on-chain storage
	Connection string `json:"connection"`
	DbName     string `json:"db_name"`
	Id         string `json:"id"`
}

type ConnectDatabaseResponse struct {
	Error string `json:"error,omitempty"`
}

type SetRouteRequest struct {
	Route           string `json:"route"`
	ContractAddress string `json:"contract_address"`
	UseOAuth2       bool   `json:"use_oauth2"`
	UseTransaction  bool   `json:"use_transaction"`
}

type SetRoutesRequest struct {
	Routes []SetRouteRequest `json:"routes"`
}

type RemoveRouteRequest struct {
	Route string `json:"route"`
}

type GetRouteRequest struct {
	Route string `json:"route"`
}

type GetRouteResponse struct {
	Error           string `json:"error,omitempty"`
	Route           string `json:"route,omitempty"`
	ContractAddress string `json:"contract_address,omitempty"`
	UseOAuth2       bool   `json:"use_oauth2"`
	UseTransaction  bool   `json:"use_transaction"`
}

type GetRoutesRequest struct{}

type RouteRecord struct {
	Route           string `json:"route"`
	ContractAddress string `json:"contract_address"`
	UseOAuth2       bool   `json:"use_oauth2"`
	UseTransaction  bool   `json:"use_transaction"`
}

type GetRoutesResponse struct {
	Routes []RouteRecord `json:"routes"`
}

type StartWebServerRequest struct {
	Config wasmxhttp.WebsrvConfig `json:"config"`
}

type StartWebServerResponse struct {
	Error string `json:"error,omitempty"`
}

type CloseServerRequest struct{}

type CloseServerResponse struct {
	Error string `json:"error,omitempty"`
}

// Internal
type serverConfig struct {
	Config wasmxhttp.WebsrvConfig `json:"config"`
}

// HTTP dispatch
type HttpRequestIncoming = wasmxhttp.HttpRequestIncoming
type HttpResponseWrap = wasmxhttp.HttpResponseWrap
type HttpResponse = wasmxhttp.HttpResponse
