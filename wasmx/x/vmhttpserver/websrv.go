package vmhttpserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	sdkerr "cosmossdk.io/errors"
	"cosmossdk.io/log"

	cfg "github.com/loredanacirstea/wasmx/config"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

type WebsrvServer struct {
	coreHandler    wasmxtypes.WasmxCosmosHandler
	parentCtx      context.Context
	logger         log.Logger
	cfg            *WebsrvConfig
	actionExecutor cfg.ActionExecutor
	senderAddress  string
	wasmxAuthority string
}

func NewWebsrvServer(
	coreHandler wasmxtypes.WasmxCosmosHandler,
	parentCtx context.Context,
	logger log.Logger,
	config *WebsrvConfig,
	actionExecutor cfg.ActionExecutor,
	senderAddress string,
	wasmxAuthority string,
) *WebsrvServer {
	return &WebsrvServer{
		coreHandler:    coreHandler,
		parentCtx:      parentCtx,
		logger:         logger.With(log.ModuleKey, "websrv"),
		cfg:            config,
		actionExecutor: actionExecutor,
		senderAddress:  senderAddress,
		wasmxAuthority: wasmxAuthority,
	}
}

func (k *WebsrvServer) Route(w http.ResponseWriter, r *http.Request) {
	// Check static routes first (longest prefix match)
	if k.cfg.StaticRoutes != nil && len(k.cfg.StaticRoutes) > 0 {
		if handled := k.serveStaticFile(w, r); handled {
			return
		}
	}

	response, err := k.HandleContractRoute(r)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	if response.Error != "" {
		io.WriteString(w, response.Error)
		return
	}
	for key, values := range response.Data.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	if response.Data.RedirectUrl != "" {
		http.Redirect(w, r, response.Data.RedirectUrl, response.Data.StatusCode)
		return
	}
	if response.Data.StatusCode != 0 {
		w.WriteHeader(response.Data.StatusCode)
	}
	w.Write([]byte(response.Data.Data))
}

// serveStaticFile checks if the request matches a static route and serves the file
// Returns true if the request was handled, false otherwise
// Supports SPA fallback: serves index.html for non-asset paths that don't exist
func (k *WebsrvServer) serveStaticFile(w http.ResponseWriter, r *http.Request) bool {
	urlPath := r.URL.Path

	// Sort routes by length (longest first) for longest prefix match
	routes := make([]string, 0, len(k.cfg.StaticRoutes))
	for route := range k.cfg.StaticRoutes {
		routes = append(routes, route)
	}
	sort.Slice(routes, func(i, j int) bool {
		return len(routes[i]) > len(routes[j])
	})

	// Find the best matching static route
	for _, route := range routes {
		if urlPath == strings.TrimSuffix(route, "/") && !strings.HasSuffix(urlPath, "/") {
			http.Redirect(w, r, urlPath+"/", http.StatusFound)
			return true
		}
		if strings.HasPrefix(urlPath, route) {
			folderPath := k.cfg.StaticRoutes[route]

			// Strip the route prefix to get the file path
			relativePath := strings.TrimPrefix(urlPath, route)
			if relativePath == "" || relativePath == "/" {
				relativePath = "/index.html"
			}

			// Clean the path to prevent directory traversal attacks
			cleanPath := filepath.Clean(relativePath)
			if strings.Contains(cleanPath, "..") {
				http.Error(w, "Invalid path", http.StatusBadRequest)
				return true
			}

			// Construct the full file path
			fullPath := filepath.Join(folderPath, cleanPath)

			// Check if file exists
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				// File doesn't exist - check if it's an asset or an SPA route
				ext := filepath.Ext(cleanPath)
				if ext != "" && isAssetExtension(ext) {
					// It's an asset that doesn't exist - return 404
					http.NotFound(w, r)
					return true
				}
				// It's likely an SPA route - serve index.html instead
				fullPath = filepath.Join(folderPath, "index.html")
				k.logger.Debug("SPA fallback to index.html", "route", route, "original", urlPath)
			}

			// Use http.ServeFile which handles MIME types, range requests, etc.
			k.logger.Debug("Serving static file", "route", route, "path", fullPath)
			http.ServeFile(w, r, fullPath)
			return true
		}
	}

	return false
}

// isAssetExtension returns true if the extension indicates a static asset
func isAssetExtension(ext string) bool {
	assetExts := map[string]bool{
		".js": true, ".mjs": true, ".cjs": true,
		".css": true, ".scss": true, ".sass": true, ".less": true,
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".svg": true, ".ico": true, ".webp": true,
		".woff": true, ".woff2": true, ".ttf": true, ".eot": true, ".otf": true,
		".json": true, ".xml": true, ".txt": true,
		".map": true,
		".wasm": true,
		".mp3": true, ".mp4": true, ".webm": true, ".ogg": true,
		".pdf": true,
	}
	return assetExts[strings.ToLower(ext)]
}

func (k *WebsrvServer) HandleContractRoute(r *http.Request) (*HttpResponseWrap, error) {
	// this can be a quick way to get the handling contract
	// if we need additional parsing of the route, we do it in the
	// contract starting the webserver
	contractAddress, ok := k.cfg.RouteToContractAddress[r.URL.Path]
	if !ok {
		// set default as the contract starting the webserver
		contractAddress = k.senderAddress
	}

	body := []byte{}
	var err error
	if r.ContentLength <= k.cfg.RequestBodyMaxSize {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, sdkerr.Wrapf(err, "incoming http request: cannot read body")
		}
	}

	req := HttpRequestIncoming{
		Method:        r.Method,
		Url:           r.URL.String(),
		RemoteAddr:    r.RemoteAddr,
		RequestURI:    r.RequestURI,
		Header:        r.Header,
		ContentLength: r.ContentLength,
		Data:          body,
	}

	httpReqBz, err := json.Marshal(req)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "cannot marshal HttpRequestGet")
	}

	msg := &networktypes.MsgReentry{
		Authority:  k.wasmxAuthority,
		Sender:     k.senderAddress,
		Contract:   contractAddress,
		EntryPoint: ENTRY_POINT_HTTP_SERVER,
		Msg:        httpReqBz,
	}
	// TODO use action executor directly to eliminate race conditions
	// TODO use action executor for all reentries
	_, resp, err := k.coreHandler.ExecuteCosmosMsg(msg)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "http server incoming request execution failed")
	}

	var requestResp networktypes.MsgReentryResponse
	err = requestResp.Unmarshal(resp)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "cannot unmarshal MsgReentryResponse")
	}

	respHttp := &HttpResponseWrap{}
	err = json.Unmarshal(requestResp.Data, respHttp)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "cannot unmarshal HttpResponseWrap")
	}
	return respHttp, nil
}
