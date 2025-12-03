package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	wasmxhttp "github.com/loredanacirstea/wasmx-env-httpserver/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// ConnectDatabase sets up a PostgreSQL connection and creates tables.
func ConnectDatabase(req ConnectDatabaseRequest) []byte {
	connID := req.Id
	if connID == "" {
		connID = MODULE_NAME
	}

	// Using on-chain storage; keep as no-op for compatibility
	_ = connID

	ok, _ := json.Marshal(ConnectDatabaseResponse{})
	return ok
}

// SetRoute registers or updates a route mapping.
func SetRoute(req SetRouteRequest) []byte {
	// basic validation
	if !strings.HasPrefix(req.Route, "/") {
		return marshalErr("route must start with /")
	}
	if req.ContractAddress == "" {
		return marshalErr("contract_address is required")
	}
	routes := loadRoutes()
	// update or insert
	found := false
	for i, r := range routes {
		if r.Route == req.Route {
			routes[i].ContractAddress = req.ContractAddress
			found = true
			break
		}
	}
	if !found {
		routes = append(routes, RouteRecord{Route: req.Route, ContractAddress: req.ContractAddress})
	}
	storeRoutes(routes)
	return []byte(`{"success":true}`)
}

func RemoveRoute(req RemoveRouteRequest) []byte {
	routes := loadRoutes()
	newRoutes := routes[:0]
	for _, r := range routes {
		if r.Route != req.Route {
			newRoutes = append(newRoutes, r)
		}
	}
	storeRoutes(newRoutes)
	return []byte(`{"success":true}`)
}

func GetRoute(req GetRouteRequest) []byte {
	routes := loadRoutes()
	var rr *RouteRecord
	for i := range routes {
		if routes[i].Route == req.Route {
			rr = &routes[i]
			break
		}
	}
	if rr == nil {
		return marshalErr("not found")
	}
	resp, _ := json.Marshal(GetRouteResponse{
		Route:           rr.Route,
		ContractAddress: rr.ContractAddress,
	})
	return resp
}

func GetRoutes() []byte {
	rs := loadRoutes()
	resp, _ := json.Marshal(GetRoutesResponse{Routes: rs})
	return resp
}

func StartWebServer(req StartWebServerRequest) []byte {
	// Build route map from DB
	routes := loadRoutes()
	routeMap := make(map[string]string, len(routes))

	// Get our own address - all routes should point to us
	selfAddr := string(wasmx.GetAddress())

	// Log registered routes
	LoggerInfo("HTTP Server Routes:", []string{"count", fmt.Sprintf("%d", len(routes))})
	if len(routes) == 0 {
		LoggerInfo("  No routes registered", nil)
	}
	for _, r := range routes {
		// Map all routes to ourselves, we'll forward internally
		routeMap[r.Route] = selfAddr
		LoggerInfo("  Route registered", []string{
			"route", r.Route,
			"contract", r.ContractAddress,
			"mapped_to", selfAddr,
		})
	}
	req.Config.RouteToContractAddress = routeMap

	resp := wasmxhttp.StartWebServer(&wasmxhttp.StartWebServerRequest{Config: req.Config})
	if resp.Error != "" {
		return marshalErr(resp.Error)
	}
	// Persist config for reference (optional)
	cfg := serverConfig{Config: req.Config}
	if cfgBz, err := json.Marshal(cfg); err == nil {
		wasmx.StorageStore([]byte(storageKeyConfig), cfgBz)
	}
	out, _ := json.Marshal(StartWebServerResponse{})
	return out
}

func CloseServer() []byte {
	resp := wasmxhttp.Close(&wasmxhttp.CloseRequest{})
	if resp.Error != "" {
		return marshalErr(resp.Error)
	}
	return []byte(`{"success":true}`)
}

// HandleHttpRequestIncoming is invoked via ENTRY_POINT=http_request_incoming
func HandleHttpRequestIncoming() {
	data := wasmx.GetCallData()
	var req HttpRequestIncoming
	if err := json.Unmarshal(data, &req); err != nil {
		wasmx.Revert([]byte("invalid http request: " + err.Error()))
		return
	}

	LoggerDebug("handle http request", []string{"req.Url", req.Url, "data", string(data)})

	best := pickBestRoute(loadRoutes(), req.Url)
	if best == nil {
		empty := wasmxhttp.HttpResponseWrap{
			Error: "",
			Data: wasmxhttp.HttpResponse{
				Status:     "404 Not Found",
				StatusCode: 404,
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
				Data:       []byte("route not found"),
			},
		}
		if bz, err := json.Marshal(empty); err == nil {
			wasmx.Finish(bz)
			return
		}
		wasmx.Finish([]byte(`{"error":"route not found"}`))
		return
	}

	LoggerDebug("handle http request", []string{"address", best.ContractAddress})

	// Forward to target contract
	call := map[string]interface{}{
		"HttpRequestHandler": req,
	}
	calld, _ := json.Marshal(call)
	ok, respData := wasmx.CallSimple(wasmx.Bech32String(best.ContractAddress), calld, false, MODULE_NAME)
	if !ok {
		wasmx.Revert(respData)
		return
	}
	wasmx.Finish(respData)
}

func marshalErr(err string) []byte {
	bz, _ := json.Marshal(map[string]string{"error": err})
	return bz
}

func loadRoutes() []RouteRecord {
	data := wasmx.StorageLoad([]byte(storageKeyRoutes))
	if len(data) == 0 {
		return []RouteRecord{}
	}
	var routes []RouteRecord
	if err := json.Unmarshal(data, &routes); err != nil {
		return []RouteRecord{}
	}
	return routes
}

func storeRoutes(routes []RouteRecord) {
	bz, _ := json.Marshal(routes)
	wasmx.StorageStore([]byte(storageKeyRoutes), bz)
}

// pickBestRoute returns the longest prefix match for url
func pickBestRoute(routes []RouteRecord, url string) *RouteRecord {
	if len(routes) == 0 {
		return nil
	}
	sort.Slice(routes, func(i, j int) bool {
		return len(routes[i].Route) > len(routes[j].Route)
	})
	for _, r := range routes {
		if strings.HasPrefix(url, r.Route) {
			return &r
		}
	}
	return nil
}
