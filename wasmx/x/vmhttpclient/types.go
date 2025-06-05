package vmhttpclient

import (
	"net/http"

	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
)

const (
	// ModuleName defines the module name
	ModuleName = "vmhttpclient"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

const HOST_WASMX_ENV_HTTP_i32_VER1 = "wasmx_httpclient_i32_1"
const HOST_WASMX_ENV_HTTP_i64_VER1 = "wasmx_httpclient_i64_1"

const HOST_WASMX_ENV_HTTP_EXPORT = "wasmx_httpclient_"

const HOST_WASMX_ENV_HTTP = "httpclient"

type Context struct {
	*vmtypes.Context
}

type HttpRequest struct {
	Method string      `json:"method"`
	Url    string      `json:"url"`
	Header http.Header `json:"header"`
	Data   []byte      `json:"data"`
	// TODO
	// MultipartForm *multipart.Form
	// TLS *tls.ConnectionState
}

type ResponseHandler struct {
	MaxSize  int64  `json:"max_size"`
	FilePath string `json:"file_path"`
}

type HttpRequestWrap struct {
	Request         HttpRequest     `json:"request"`
	ResponseHandler ResponseHandler `json:"response_handler"`
}

type HttpResponse struct {
	Status        string      `json:"status"`
	StatusCode    int         `json:"status_code"`
	ContentLength int64       `json:"content_length"`
	Uncompressed  bool        `json:"uncompressed"`
	Header        http.Header `json:"header"`
	Data          []byte      `json:"data"`
	// TODO
	// TLS           *tls.ConnectionState  `json:"tls"`
}

type HttpResponseWrap struct {
	Error string       `json:"error"`
	Data  HttpResponse `json:"data"`
}
