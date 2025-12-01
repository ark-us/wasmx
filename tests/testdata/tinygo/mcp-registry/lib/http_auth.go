package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	httpserver "github.com/loredanacirstea/wasmx-env-httpserver/lib"
)

// handleLoginPage shows the login form with optional return_to parameter
func handleLoginPage(req *httpserver.HttpRequestIncoming) httpserver.HttpResponseWrap {
	// Parse query parameters
	parsedURL, err := url.Parse(req.Url)
	if err != nil {
		return sendTextResponse("Invalid URL", 400)
	}
	query := parsedURL.Query()
	returnTo := query.Get("return_to")

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<title>Login - MCP Registry</title>
	<style>
		body { font-family: sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
		input { width: 100%%; padding: 8px; margin: 5px 0; box-sizing: border-box; }
		button { width: 100%%; padding: 10px; margin: 10px 0; background: #007bff; color: white; border: none; cursor: pointer; }
		button:hover { background: #0056b3; }
		.error { color: red; margin: 10px 0; }
		.link { text-align: center; margin-top: 20px; }
	</style>
</head>
<body>
	<h1>Login</h1>
	<div id="error" class="error"></div>
	<form id="loginForm">
		<input type="email" name="email" placeholder="Email" required />
		<input type="password" name="password" placeholder="Password" required />
		<button type="submit">Login</button>
	</form>
	<div class="link">
		<a href="/auth/register?return_to=%s">Don't have an account? Register</a>
	</div>
	<script>
		document.getElementById('loginForm').onsubmit = async (e) => {
			e.preventDefault();
			const formData = new FormData(e.target);
			const data = {
				email: formData.get('email'),
				password: formData.get('password')
			};

			try {
				const resp = await fetch('/auth/login', {
					method: 'POST',
					headers: {'Content-Type': 'application/json'},
					body: JSON.stringify(data)
				});
				const result = await resp.json();

				if (resp.ok) {
					// Store session ID in cookie
					document.cookie = 'session_id=' + result.session_id + '; path=/; max-age=86400';
					// Redirect
					const returnTo = '%s' || '/';
					window.location.href = returnTo;
				} else {
					document.getElementById('error').textContent = result.error || 'Login failed';
				}
			} catch (err) {
				document.getElementById('error').textContent = 'Network error';
			}
		};
	</script>
</body>
</html>
`, returnTo, returnTo)

	return sendHTMLResponse(html, 200)
}

// handleRegister shows registration form and handles registration
func handleRegister(req *httpserver.HttpRequestIncoming) httpserver.HttpResponseWrap {
	if req.Method == "GET" {
		// Parse query parameters
		parsedURL, _ := url.Parse(req.Url)
		query := parsedURL.Query()
		returnTo := query.Get("return_to")

		html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<title>Register - MCP Registry</title>
	<style>
		body { font-family: sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
		input { width: 100%%; padding: 8px; margin: 5px 0; box-sizing: border-box; }
		button { width: 100%%; padding: 10px; margin: 10px 0; background: #28a745; color: white; border: none; cursor: pointer; }
		button:hover { background: #218838; }
		.error { color: red; margin: 10px 0; }
		.link { text-align: center; margin-top: 20px; }
	</style>
</head>
<body>
	<h1>Register</h1>
	<div id="error" class="error"></div>
	<form id="registerForm">
		<input type="email" name="email" placeholder="Email" required />
		<input type="text" name="username" placeholder="Username (optional)" />
		<input type="password" name="password" placeholder="Password (min 8 chars)" required minlength="8" />
		<button type="submit">Register</button>
	</form>
	<div class="link">
		<a href="/login?return_to=%s">Already have an account? Login</a>
	</div>
	<script>
		document.getElementById('registerForm').onsubmit = async (e) => {
			e.preventDefault();
			const formData = new FormData(e.target);
			const data = {
				email: formData.get('email'),
				username: formData.get('username'),
				password: formData.get('password')
			};

			try {
				const resp = await fetch('/auth/register', {
					method: 'POST',
					headers: {'Content-Type': 'application/json'},
					body: JSON.stringify(data)
				});
				const result = await resp.json();

				if (resp.ok) {
					// Auto-login after registration
					const loginResp = await fetch('/auth/login', {
						method: 'POST',
						headers: {'Content-Type': 'application/json'},
						body: JSON.stringify({email: data.email, password: data.password})
					});
					const loginResult = await loginResp.json();

					if (loginResp.ok) {
						document.cookie = 'session_id=' + loginResult.session_id + '; path=/; max-age=86400';
						const returnTo = '%s' || '/';
						window.location.href = returnTo;
					}
				} else {
					document.getElementById('error').textContent = result.error || 'Registration failed';
				}
			} catch (err) {
				document.getElementById('error').textContent = 'Network error';
			}
		};
	</script>
</body>
</html>
`, returnTo, returnTo)

		return sendHTMLResponse(html, 200)
	}

	// POST - handle registration
	var regReq RegisterUserRequest
	if err := json.Unmarshal(req.Data, &regReq); err != nil {
		return sendJSONResponse(map[string]string{"error": "Invalid request"}, 400)
	}

	result := RegisterUser(regReq)
	return sendJSONResponseBytes(result, 200)
}

// handleLoginSubmit handles POST /auth/login
func handleLoginSubmit(req *httpserver.HttpRequestIncoming) httpserver.HttpResponseWrap {
	var loginReq LoginRequest
	if err := json.Unmarshal(req.Data, &loginReq); err != nil {
		return sendJSONResponse(map[string]string{"error": "Invalid request"}, 400)
	}

	result := Login(loginReq)
	return sendJSONResponseBytes(result, 200)
}

// handleLogout handles POST /auth/logout
func handleLogout(req *httpserver.HttpRequestIncoming) httpserver.HttpResponseWrap {
	// Get session ID from cookie
	sessionID := getSessionFromCookie(req)
	if sessionID == "" {
		return sendJSONResponse(map[string]string{"error": "No session"}, 400)
	}

	logoutReq := LogoutRequest{SessionID: sessionID}
	result := Logout(logoutReq)
	return sendJSONResponseBytes(result, 200)
}

// handleGetCurrentUser handles GET /auth/me
func handleGetCurrentUser(req *httpserver.HttpRequestIncoming) httpserver.HttpResponseWrap {
	// Get session ID from cookie
	sessionID := getSessionFromCookie(req)
	if sessionID == "" {
		return sendJSONResponse(map[string]string{"error": "No session"}, 401)
	}

	currentUserReq := GetCurrentUserRequest{SessionID: sessionID}
	result := GetCurrentUser(currentUserReq)
	return sendJSONResponseBytes(result, 200)
}

// Helper: extract session ID from cookie header
func getSessionFromCookie(req *httpserver.HttpRequestIncoming) string {
	cookies := req.Headers["Cookie"]
	if len(cookies) == 0 {
		return ""
	}

	// Parse cookies
	for _, cookie := range cookies {
		parts := strings.Split(cookie, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "session_id=") {
				return strings.TrimPrefix(part, "session_id=")
			}
		}
	}

	return ""
}
