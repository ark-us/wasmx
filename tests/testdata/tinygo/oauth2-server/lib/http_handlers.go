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
	host := req.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = req.Header.Get("Host")
	}
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
	if req.Method == "GET" {
		// Return HTML login form
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Login</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
        input { width: 100%; padding: 10px; margin: 10px 0; box-sizing: border-box; }
        button { width: 100%; padding: 10px; background: #007bff; color: white; border: none; cursor: pointer; }
        button:hover { background: #0056b3; }
        .error { color: red; }
    </style>
</head>
<body>
    <h2>Login</h2>
    <form method="POST" action="/login">
        <input type="email" name="email" placeholder="Email" required />
        <input type="password" name="password" placeholder="Password" required />
        <button type="submit">Login</button>
    </form>
</body>
</html>`
		return marshalHTTP(wasmxhttp.HttpResponseWrap{
			Error: "",
			Data: wasmxhttp.HttpResponse{
				Status:     "200 OK",
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
				Data:       []byte(html),
			},
		})
	}

	if req.Method != "POST" {
		return methodNotAllowed()
	}

	// Parse form data or JSON
	var r LoginRequest
	contentType := req.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		form, _ := url.ParseQuery(string(req.Data))
		r.Email = form.Get("email")
		r.Password = form.Get("password")
	} else {
		if err := json.Unmarshal(req.Data, &r); err != nil {
			return badRequest("invalid json")
		}
	}

	resp := Login(r)

	// Parse response to set session cookie
	var loginResp LoginResponse
	json.Unmarshal(resp, &loginResp)

	headers := http.Header{"Content-Type": []string{"application/json"}}
	if loginResp.SessionID != "" {
		// Set session cookie
		cookie := fmt.Sprintf("session_id=%s; Path=/; HttpOnly; Max-Age=%d", loginResp.SessionID, SESSION_DURATION_SECONDS)
		headers.Set("Set-Cookie", cookie)
	}

	return marshalHTTP(wasmxhttp.HttpResponseWrap{
		Error: "",
		Data: wasmxhttp.HttpResponse{
			Status:     "200 OK",
			StatusCode: 200,
			Header:     headers,
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
	if req.Method != "GET" && req.Method != "POST" {
		return methodNotAllowed()
	}

	u, err := url.Parse(req.Url)
	if err != nil {
		return badRequest("invalid url")
	}
	q := u.Query()

	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	state := q.Get("state")
	codeChallenge := q.Get("code_challenge")
	codeChallengeMethod := q.Get("code_challenge_method")

	// Extract session from cookie
	sessionID := extractSessionFromCookie(req)
	if sessionID == "" {
		// Redirect to login with return URL
		loginURL := fmt.Sprintf("/login?redirect=%s", url.QueryEscape(req.Url))
		return marshalHTTP(wasmxhttp.HttpResponseWrap{
			Error: "",
			Data: wasmxhttp.HttpResponse{
				Status:     "302 Found",
				StatusCode: 302,
				Header:     http.Header{"Location": []string{loginURL}},
				RedirectUrl: loginURL,
			},
		})
	}

	// Validate session and get user ID
	sessionResp := ValidateSession(ValidateSessionRequest{SessionID: sessionID})
	var sessionResult ValidateSessionResponse
	json.Unmarshal(sessionResp, &sessionResult)

	if !sessionResult.Valid {
		loginURL := fmt.Sprintf("/login?redirect=%s", url.QueryEscape(req.Url))
		return marshalHTTP(wasmxhttp.HttpResponseWrap{
			Error: "",
			Data: wasmxhttp.HttpResponse{
				Status:     "302 Found",
				StatusCode: 302,
				Header:     http.Header{"Location": []string{loginURL}},
				RedirectUrl: loginURL,
			},
		})
	}

	userID := sessionResult.UserID

	// If GET, show authorization page
	if req.Method == "GET" {
		// Get client info
		clientResp := GetOAuthClient(GetOAuthClientRequest{ClientID: clientID})
		var client map[string]interface{}
		json.Unmarshal(clientResp, &client)

		clientName := "Unknown Application"
		if name, ok := client["name"].(string); ok {
			clientName = name
		}

		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Authorize Application</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 500px; margin: 50px auto; padding: 20px; }
        .app-info { background: #f5f5f5; padding: 15px; margin: 20px 0; border-radius: 5px; }
        button { padding: 10px 20px; margin: 10px 5px; cursor: pointer; }
        .authorize { background: #28a745; color: white; border: none; }
        .deny { background: #dc3545; color: white; border: none; }
    </style>
</head>
<body>
    <h2>Authorize Application</h2>
    <div class="app-info">
        <p><strong>%s</strong> wants to access your account</p>
        <p>This will allow the application to:</p>
        <ul>
            <li>Access your profile information</li>
            <li>Use MCP tools on your behalf</li>
        </ul>
    </div>
    <form method="POST" action="/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s&code_challenge=%s&code_challenge_method=%s&response_type=code">
        <button type="submit" class="authorize">Authorize</button>
        <button type="button" class="deny" onclick="window.location.href='%s?error=access_denied&state=%s'">Deny</button>
    </form>
</body>
</html>`, clientName, clientID, url.QueryEscape(redirectURI), state, codeChallenge, codeChallengeMethod, redirectURI, state)

		return marshalHTTP(wasmxhttp.HttpResponseWrap{
			Error: "",
			Data: wasmxhttp.HttpResponse{
				Status:     "200 OK",
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
				Data:       []byte(html),
			},
		})
	}

	// POST - user approved, create authorization code
	if q.Get("response_type") != "code" {
		return badRequest("response_type must be code")
	}

	codeResp := CreateAuthorizationCode(CreateAuthorizationCodeRequest{
		ClientID:            clientID,
		UserID:              userID,
		RedirectURI:         redirectURI,
		Scopes:              []string{"read", "write", "tools"},
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
	})

	// Redirect with authorization code
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

func extractSessionFromCookie(req wasmxhttp.HttpRequestIncoming) string {
	cookies := req.Header.Get("Cookie")
	if cookies == "" {
		return ""
	}
	for _, cookie := range strings.Split(cookies, ";") {
		parts := strings.SplitN(strings.TrimSpace(cookie), "=", 2)
		if len(parts) == 2 && parts[0] == "session_id" {
			return parts[1]
		}
	}
	return ""
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
