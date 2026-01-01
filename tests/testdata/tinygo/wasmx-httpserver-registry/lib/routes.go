package lib

import (
	"encoding/hex"
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
			routes[i].UseOAuth2 = req.UseOAuth2
			routes[i].UseTransaction = req.UseTransaction
			found = true
			break
		}
	}
	if !found {
		routes = append(routes, RouteRecord{
			Route:           req.Route,
			ContractAddress: req.ContractAddress,
			UseOAuth2:       req.UseOAuth2,
			UseTransaction:  req.UseTransaction,
		})
	}
	storeRoutes(routes)
	return []byte(`{"success":true}`)
}

// SetRoutes registers or updates multiple route mappings in batch
func SetRoutes(req SetRoutesRequest) []byte {
	routes := loadRoutes()

	for _, routeReq := range req.Routes {
		// basic validation
		if !strings.HasPrefix(routeReq.Route, "/") {
			return marshalErr("route must start with /: " + routeReq.Route)
		}
		if routeReq.ContractAddress == "" {
			return marshalErr("contract_address is required for route: " + routeReq.Route)
		}
		fmt.Println("--wasmx.httpserver.SetRoutes.routeReq--", routeReq.Route, routeReq.UseTransaction)

		// update or insert
		found := false
		for i, r := range routes {
			if r.Route == routeReq.Route {
				routes[i].ContractAddress = routeReq.ContractAddress
				routes[i].UseOAuth2 = routeReq.UseOAuth2
				routes[i].UseTransaction = routeReq.UseTransaction
				found = true
				break
			}
		}
		if !found {
			routes = append(routes, RouteRecord{
				Route:           routeReq.Route,
				ContractAddress: routeReq.ContractAddress,
				UseOAuth2:       routeReq.UseOAuth2,
				UseTransaction:  routeReq.UseTransaction,
			})
		}
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
		UseOAuth2:       rr.UseOAuth2,
		UseTransaction:  rr.UseTransaction,
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

	LoggerDebug("handle http request", []string{"method", req.Method, "url", req.Url, "data", string(data)})
	fmt.Println("--wasmx.httpserver.HandleHttpRequestIncoming--", req.Method, req.Url, string(data))

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

	LoggerDebug("handle http request", []string{"address", best.ContractAddress, "method", req.Method})

	fmt.Println("--wasmx.httpserver.HandleHttpRequestIncoming.UseTransaction--", best.Route, best.UseTransaction, req.Method, req.Method == "POST")

	// Check if this is a mutating request (POST/PUT/DELETE) that requires transaction signing
	if best.UseTransaction && (req.Method == "POST" || req.Method == "PUT" || req.Method == "DELETE") {
		fmt.Println("--wasmx.httpserver.HandleHttpRequestIncoming.handleMutatingRequest--")
		handleMutatingRequest(req, best.ContractAddress)
		return
	}
	fmt.Println("--wasmx.httpserver.HandleHttpRequestIncoming.HttpRequestHandler--")

	// GET requests or non-transaction POST requests - call directly
	call := map[string]interface{}{
		"HttpRequestHandler": req,
	}
	calld, _ := json.Marshal(call)
	ok, respData := wasmx.CallSimple(wasmx.Bech32String(best.ContractAddress), calld, false, MODULE_NAME)
	if !ok {
		// Return HTTP 500 error instead of reverting, so clients get proper status code
		LoggerError("httpserver forward failed", []string{"error", string(respData), "address", best.ContractAddress})
		respondWithError("internal server error: "+string(respData), 500)
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

// handleMutatingRequest handles POST/PUT/DELETE requests by delegating transaction signing to oauth2-keys contract
func handleMutatingRequest(req HttpRequestIncoming, targetContract string) {
	// Extract OAuth token from Authorization header
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError("missing Authorization header", 401)
		return
	}

	// Parse Bearer token
	var oauthToken string
	if strings.HasPrefix(authHeader, "Bearer ") {
		oauthToken = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		respondWithError("invalid Authorization header format", 401)
		return
	}

	LoggerInfo("Processing mutating request", []string{"method", req.Method, "oauth_token", oauthToken[:10] + "..."})

	// Get OAuth2 keys contract address
	oauth2KeysAddr := wasmx.GetAddressByRole(wasmx.ROLE_OAUTH2_KEYS)
	if oauth2KeysAddr == "" {
		respondWithError("OAuth2 keys contract not found", 500)
		return
	}

	// Build transaction calldata for target contract
	call := map[string]interface{}{
		"HttpRequestHandler": req,
	}
	calld, _ := json.Marshal(call)

	// Load gas price from storage
	gasPrice := LoadGasPrice()
	if gasPrice == nil {
		gasPrice = &wasmx.Coin{}
	}

	// Delegate signing and broadcasting to oauth2-keys contract
	// This ensures consistent address derivation with the funded ephemeral address
	signAndBroadcastReq := map[string]interface{}{
		"sign_and_broadcast_tx": map[string]interface{}{
			"oauth_token":     oauthToken,
			"target_contract": targetContract,
			"calldata":        calld,
			"gas_limit":       uint64(30000000),
			"gas_price":       gasPrice,
		},
	}
	signAndBroadcastBz, _ := json.Marshal(signAndBroadcastReq)
	fmt.Println("--wasmx.httpserver.handleMutatingRequest--", string(signAndBroadcastBz))

	ok, signResp := wasmx.CallSimple(oauth2KeysAddr, signAndBroadcastBz, false, MODULE_NAME)
	fmt.Println("--wasmx.httpserver.handleMutatingRequest.resp--", ok, string(signResp))
	if !ok {
		LoggerError("Failed to sign and broadcast transaction", []string{"error", string(signResp)})
		respondWithError("failed to sign and broadcast transaction: "+string(signResp), 500)
		return
	}

	var signResponse struct {
		Success bool   `json:"success"`
		TxHash  []byte `json:"tx_hash"`
		Address string `json:"address"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(signResp, &signResponse); err != nil {
		LoggerError("Failed to parse sign response", []string{"error", err.Error()})
		respondWithError("failed to parse sign response", 500)
		return
	}

	if !signResponse.Success {
		LoggerError("Sign and broadcast failed", []string{"error", signResponse.Error})
		// Check if it's an auth error
		if strings.Contains(signResponse.Error, "invalid OAuth token") {
			respondWithError(signResponse.Error, 401)
		} else {
			respondWithError(signResponse.Error, 500)
		}
		return
	}

	LoggerInfo("Transaction broadcasted", []string{"tx_hash", hex.EncodeToString(signResponse.TxHash), "address", signResponse.Address})

	response := wasmxhttp.HttpResponseWrap{
		Error: "",
		Data: wasmxhttp.HttpResponse{
			Status:     "200 OK",
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Data:       []byte(fmt.Sprintf(`{"tx_hash":"%x","address":"%s"}`, signResponse.TxHash, signResponse.Address)),
		},
	}

	if bz, err := json.Marshal(response); err == nil {
		wasmx.Finish(bz)
	} else {
		wasmx.Finish([]byte(`{"error":"failed to marshal response"}`))
	}
}

// respondWithError sends an error response
func respondWithError(message string, statusCode int) {
	status := fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode))
	response := wasmxhttp.HttpResponseWrap{
		Error: "",
		Data: wasmxhttp.HttpResponse{
			Status:     status,
			StatusCode: statusCode,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Data:       []byte(fmt.Sprintf(`{"error":"%s"}`, message)),
		},
	}
	if bz, err := json.Marshal(response); err == nil {
		wasmx.Finish(bz)
	} else {
		wasmx.Finish([]byte(`{"error":"internal error"}`))
	}
}
