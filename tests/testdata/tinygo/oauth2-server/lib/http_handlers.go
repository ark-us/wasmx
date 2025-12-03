package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	wasmxhttp "github.com/loredanacirstea/wasmx-env-httpserver/lib"
)

// HandleHttpRequestWrap handles HTTP requests forwarded from HTTP server registry via HttpRequestHandler
func HandleHttpRequestWrap(req wasmxhttp.HttpRequestIncoming) []byte {
	path := req.RequestURI
	if path == "" {
		path = req.Url
	}
	pathOnly := strings.SplitN(path, "?", 2)[0]

	switch pathOnly {
	case "/.well-known/oauth-authorization-server":
		return handleWellKnown(req)
	case "/oauth/authorize":
		return handleAuthorize(req)
	case "/oauth/token":
		return handleToken(req)
	case "/register":
		return handleRegister(req)
	case "/login":
		return handleLogin(req)
	case "/logout":
		return handleLogout(req)
	default:
		return marshalHTTP(wasmxhttp.HttpResponseWrap{
			Error: "",
			Data: wasmxhttp.HttpResponse{
				Status:     "404 Not Found",
				StatusCode: 404,
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
				Data:       []byte("not found"),
			},
		})
	}
}

func handleWellKnown(req wasmxhttp.HttpRequestIncoming) []byte {
	scheme := "http"
	if strings.EqualFold(req.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	host := req.Header.Get("Host")
	base := fmt.Sprintf("%s://%s", scheme, host)
	body, _ := json.Marshal(map[string]interface{}{
		"issuer":                                base,
		"authorization_endpoint":                base + "/oauth/authorize",
		"token_endpoint":                        base + "/oauth/token",
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"response_types_supported":              []string{"code"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post"},
		"code_challenge_methods_supported":      []string{"plain", "S256"},
	})
	return marshalHTTP(wasmxhttp.HttpResponseWrap{
		Error: "",
		Data: wasmxhttp.HttpResponse{
			Status:     "200 OK",
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Data:       body,
		},
	})
}

func handleRegister(req wasmxhttp.HttpRequestIncoming) []byte {
	if req.Method != "POST" {
		return methodNotAllowed()
	}
	var r RegisterUserRequest
	if err := json.Unmarshal(req.Data, &r); err != nil {
		return badRequest("invalid json")
	}
	resp := RegisterUser(r)
	return marshalHTTP(wasmxhttp.HttpResponseWrap{
		Error: "",
		Data: wasmxhttp.HttpResponse{
			Status:     "200 OK",
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Data:       resp,
		},
	})
}

func handleLogin(req wasmxhttp.HttpRequestIncoming) []byte {
	if req.Method != "POST" {
		return methodNotAllowed()
	}
	var r LoginRequest
	if err := json.Unmarshal(req.Data, &r); err != nil {
		return badRequest("invalid json")
	}
	resp := Login(r)
	return marshalHTTP(wasmxhttp.HttpResponseWrap{
		Error: "",
		Data: wasmxhttp.HttpResponse{
			Status:     "200 OK",
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Data:       resp,
		},
	})
}

func handleLogout(req wasmxhttp.HttpRequestIncoming) []byte {
	if req.Method != "POST" {
		return methodNotAllowed()
	}
	var r LogoutRequest
	if err := json.Unmarshal(req.Data, &r); err != nil {
		return badRequest("invalid json")
	}
	resp := Logout(r)
	return marshalHTTP(wasmxhttp.HttpResponseWrap{
		Error: "",
		Data: wasmxhttp.HttpResponse{
			Status:     "200 OK",
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Data:       resp,
		},
	})
}

func handleAuthorize(req wasmxhttp.HttpRequestIncoming) []byte {
	if req.Method != "GET" {
		return methodNotAllowed()
	}
	u, err := url.Parse(req.Url)
	if err != nil {
		return badRequest("invalid url")
	}
	q := u.Query()
	if q.Get("response_type") != "code" {
		return badRequest("response_type must be code")
	}
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	state := q.Get("state")
	codeChallenge := q.Get("code_challenge")
	codeChallengeMethod := q.Get("code_challenge_method")
	userID := q.Get("user_id") // simplistic; real app should use session

	codeResp := CreateAuthorizationCode(CreateAuthorizationCodeRequest{
		ClientID:            clientID,
		UserID:              userID,
		RedirectURI:         redirectURI,
		Scopes:              []string{}, // optional; could accept from query
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
	})

	// Redirect if redirect_uri provided
	if redirectURI != "" {
		rurl, _ := url.Parse(redirectURI)
		qq := rurl.Query()
		var respObj CreateAuthorizationCodeResponse
		_ = json.Unmarshal(codeResp, &respObj)
		if respObj.Code != "" {
			qq.Set("code", respObj.Code)
		}
		if state != "" {
			qq.Set("state", state)
		}
		rurl.RawQuery = qq.Encode()
		return marshalHTTP(wasmxhttp.HttpResponseWrap{
			Error: "",
			Data: wasmxhttp.HttpResponse{
				Status:     "302 Found",
				StatusCode: 302,
				Header:     http.Header{"Location": []string{rurl.String()}},
				RedirectUrl: rurl.String(),
			},
		})
	}

	return marshalHTTP(wasmxhttp.HttpResponseWrap{
		Error: "",
		Data: wasmxhttp.HttpResponse{
			Status:     "200 OK",
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Data:       codeResp,
		},
	})
}

func handleToken(req wasmxhttp.HttpRequestIncoming) []byte {
	if req.Method != "POST" {
		return methodNotAllowed()
	}
	// Try form-encoded first
	form, _ := url.ParseQuery(string(req.Data))
	grantType := form.Get("grant_type")
	switch grantType {
	case "authorization_code":
		r := ExchangeCodeForTokenRequest{
			Code:         form.Get("code"),
			ClientID:     form.Get("client_id"),
			ClientSecret: form.Get("client_secret"),
			RedirectURI:  form.Get("redirect_uri"),
			CodeVerifier: form.Get("code_verifier"),
		}
		resp := ExchangeCodeForToken(r)
		return marshalHTTP(wasmxhttp.HttpResponseWrap{
			Error: "",
			Data: wasmxhttp.HttpResponse{
				Status:     "200 OK",
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Data:       resp,
			},
		})
	case "refresh_token":
		r := RefreshAccessTokenRequest{
			RefreshToken: form.Get("refresh_token"),
			ClientID:     form.Get("client_id"),
			ClientSecret: form.Get("client_secret"),
		}
		resp := RefreshAccessToken(r)
		return marshalHTTP(wasmxhttp.HttpResponseWrap{
			Error: "",
			Data: wasmxhttp.HttpResponse{
				Status:     "200 OK",
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Data:       resp,
			},
		})
	default:
		return badRequest("unsupported_grant_type")
	}
}

func methodNotAllowed() []byte {
	return marshalHTTP(wasmxhttp.HttpResponseWrap{
		Error: "",
		Data: wasmxhttp.HttpResponse{
			Status:     "405 Method Not Allowed",
			StatusCode: 405,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Data:       []byte("method not allowed"),
		},
	})
}

func badRequest(msg string) []byte {
	return marshalHTTP(wasmxhttp.HttpResponseWrap{
		Error: "",
		Data: wasmxhttp.HttpResponse{
			Status:     "400 Bad Request",
			StatusCode: 400,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Data:       []byte(msg),
		},
	})
}

func marshalHTTP(resp wasmxhttp.HttpResponseWrap) []byte {
	bz, _ := json.Marshal(resp)
	return bz
}
