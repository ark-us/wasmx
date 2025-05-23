package httpserver_registry

import "github.com/loredanacirstea/wasmx/x/vmhttpserver"

type RolesChangedHook struct{}

type GetRoutesRequest struct{}

type GetRouteRequest struct {
	Route string `json:"route"`
}

type CalldataTestHttpRegistry struct {
	RoleChanged *RolesChangedHook `json:"RoleChanged"`

	StartWebServer *vmhttpserver.StartWebServerRequest `json:"StartWebServer"`
	Close          *vmhttpserver.CloseRequest          `json:"Close"`

	SetRoute           *vmhttpserver.SetRouteHandlerRequest    `json:"SetRoute"`
	RemoveRoute        *vmhttpserver.RemoveRouteHandlerRequest `json:"RemoveRoute"`
	HttpRequestHandler *vmhttpserver.HttpRequestIncoming       `json:"HttpRequestHandler"`
	GetRoutes          *GetRoutesRequest                       `json:"GetRoutes"`
	GetRoute           *GetRouteRequest                        `json:"GetRoute"`
}
