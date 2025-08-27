package httpclient

import "net/http"

type HttpRequest struct {
	Method string      `json:"method"`
	Url    string      `json:"url"`
	Header http.Header `json:"header"`
	Data   []byte      `json:"data"`
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
}

type HttpResponseWrap struct {
	Error string       `json:"error"`
	Data  HttpResponse `json:"data"`
}
