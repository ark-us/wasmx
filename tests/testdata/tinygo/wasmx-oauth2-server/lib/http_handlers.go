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
	fmt.Println("--wasmx.oauth2server.HandleHttpRequestWrap--")
	path := req.RequestURI
	if path == "" {
		path = req.Url
	}
	pathOnly := strings.SplitN(path, "?", 2)[0]

	switch pathOnly {
	case "/.well-known/oauth-authorization-server", "/.well-known/oauth-authorization-server/sse":
		return handleWellKnown(req)
	case "/.well-known/openid-configuration", "/.well-known/openid-configuration/sse":
		return handleWellKnown(req)
	case "/.well-known/oauth-protected-resource", "/.well-known/oauth-protected-resource/sse":
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
	fmt.Println("--wasmx.oauth2server.handleRegister--")
	if req.Method == "GET" {
		// Return HTML registration form with blockchain key generation
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Register</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 500px; margin: 50px auto; padding: 20px; }
        input { width: 100%; padding: 10px; margin: 10px 0; box-sizing: border-box; }
        button { width: 100%; padding: 10px; background: #28a745; color: white; border: none; cursor: pointer; margin-top: 10px; }
        button:hover { background: #218838; }
        .error { color: red; margin: 10px 0; }
        .success { color: green; margin: 10px 0; }
        .checkbox-label { display: flex; align-items: center; margin: 15px 0; }
        .checkbox-label input { width: auto; margin-right: 10px; }
        #blockchainOptions { display: none; padding: 15px; background: #f8f9fa; border-radius: 5px; margin: 15px 0; }
        .info { background: #e7f3ff; padding: 10px; border-radius: 5px; margin: 10px 0; font-size: 14px; }
    </style>
</head>
<body>
    <h2>Register</h2>
    <form id="registerForm">
        <input type="email" id="email" name="email" placeholder="Email" required />
        <input type="password" id="password" name="password" placeholder="Password (min 8 characters)" required minlength="8" />
        <input type="text" id="username" name="username" placeholder="Username (optional)" />

        <label class="checkbox-label">
            <input type="checkbox" id="enableBlockchain" />
            Enable WasmX Blockchain Account
        </label>

        <div id="blockchainOptions">
            <div class="info">
                A cryptographic key pair will be generated in your browser and encrypted with a 4-digit PIN.
                Your private key never leaves your browser and is never sent to the server.
            </div>
            <input type="text" id="pin" placeholder="4-digit PIN" pattern="[0-9]{4}" maxlength="4" />
            <input type="text" id="pinConfirm" placeholder="Confirm PIN" pattern="[0-9]{4}" maxlength="4" />
        </div>

        <button type="submit">Register</button>
        <div id="error" class="error"></div>
        <div id="success" class="success"></div>
    </form>
    <p>Already have an account? <a href="/login">Login here</a></p>

    <script type="module">
        import * as secp from 'https://cdn.jsdelivr.net/npm/@noble/secp256k1@2.1.0/+esm';
        import { sha256 } from 'https://cdn.jsdelivr.net/npm/@noble/hashes@1.3.3/sha256/+esm';
        import { ripemd160 } from 'https://cdn.jsdelivr.net/npm/@noble/hashes@1.3.3/ripemd160/+esm';
        import { bech32 } from 'https://cdn.jsdelivr.net/npm/bech32@2.0.0/+esm';

        document.getElementById('enableBlockchain').addEventListener('change', function(e) {
            document.getElementById('blockchainOptions').style.display = e.target.checked ? 'block' : 'none';
            document.getElementById('pin').required = e.target.checked;
            document.getElementById('pinConfirm').required = e.target.checked;
        });

        // Derive encryption key from PIN using PBKDF2
        async function deriveKeyFromPIN(pin) {
            const enc = new TextEncoder();
            const pinData = enc.encode(pin);
            const salt = enc.encode('wasmx-identity-v1'); // Fixed salt for deterministic derivation

            const keyMaterial = await crypto.subtle.importKey(
                'raw',
                pinData,
                'PBKDF2',
                false,
                ['deriveBits', 'deriveKey']
            );

            return await crypto.subtle.deriveKey(
                {
                    name: 'PBKDF2',
                    salt: salt,
                    iterations: 100000,
                    hash: 'SHA-256'
                },
                keyMaterial,
                { name: 'AES-GCM', length: 256 },
                true,
                ['encrypt', 'decrypt']
            );
        }

        // Encrypt private key with PIN-derived key
        async function encryptPrivateKey(privateKeyHex, pin) {
            const key = await deriveKeyFromPIN(pin);
            const iv = crypto.getRandomValues(new Uint8Array(12));
            const enc = new TextEncoder();
            const data = enc.encode(privateKeyHex);

            const encrypted = await crypto.subtle.encrypt(
                { name: 'AES-GCM', iv: iv },
                key,
                data
            );

            // Combine IV and encrypted data
            const combined = new Uint8Array(iv.length + encrypted.byteLength);
            combined.set(iv);
            combined.set(new Uint8Array(encrypted), iv.length);

            // Return as hex
            return Array.from(combined).map(b => b.toString(16).padStart(2, '0')).join('');
        }

        // Generate secp256k1 key pair
        function generateKeyPair() {
            const privateKey = secp.utils.randomPrivateKey();
            const publicKey = secp.getPublicKey(privateKey, true); // compressed
            return {
                privateKey: Array.from(privateKey).map(b => b.toString(16).padStart(2, '0')).join(''),
                publicKey: Array.from(publicKey).map(b => b.toString(16).padStart(2, '0')).join('')
            };
        }

        // Derive Cosmos SDK-compatible bech32 address from public key
        // This follows the standard Cosmos SDK address derivation (coin type 118)
        function deriveAddress(publicKeyHex) {
            // Convert hex public key to bytes
            const pubKeyBytes = new Uint8Array(publicKeyHex.match(/.{1,2}/g).map(byte => parseInt(byte, 16)));

            // Cosmos SDK uses sha256 then ripemd160
            const sha256Hash = sha256(pubKeyBytes);
            const ripemd160Hash = ripemd160(sha256Hash);

            // Convert to bech32 with 'wasmx' prefix
            const words = bech32.toWords(ripemd160Hash);
            const address = bech32.encode('wasmx', words);

            return address;
        }

        document.getElementById('registerForm').addEventListener('submit', async function(e) {
            e.preventDefault();

            const errorDiv = document.getElementById('error');
            const successDiv = document.getElementById('success');
            errorDiv.textContent = '';
            successDiv.textContent = '';

            const email = document.getElementById('email').value;
            const password = document.getElementById('password').value;
            const username = document.getElementById('username').value;
            const enableBlockchain = document.getElementById('enableBlockchain').checked;

            let requestData = { email, password, username };

            if (enableBlockchain) {
                const pin = document.getElementById('pin').value;
                const pinConfirm = document.getElementById('pinConfirm').value;

                if (pin !== pinConfirm) {
                    errorDiv.textContent = 'PINs do not match';
                    return;
                }

                if (!/^[0-9]{4}$/.test(pin)) {
                    errorDiv.textContent = 'PIN must be exactly 4 digits';
                    return;
                }

                try {
                    successDiv.textContent = 'Generating cryptographic keys...';

                    // Generate key pair
                    const keyPair = generateKeyPair();
                    const address = deriveAddress(keyPair.publicKey);

                    // LOG: Show generated keys in console
                    console.log('=== BROWSER-SIDE KEY GENERATION ===');
                    console.log('Private Key (hex):', keyPair.privateKey);
                    console.log('Public Key (compressed, hex):', keyPair.publicKey);
                    console.log('Derived Address (Cosmos SDK bech32):', address);
                    console.log('✓ Address format IS Cosmos SDK coin 118 compatible!');
                    console.log('===================================');

                    // Encrypt private key with PIN
                    const encryptedPrivateKey = await encryptPrivateKey(keyPair.privateKey, pin);

                    // Store encrypted private key in localStorage
                    localStorage.setItem('wasmx_encrypted_key_' + email, encryptedPrivateKey);

                    // Add blockchain info to registration
                    requestData.public_key = keyPair.publicKey;
                    requestData.address = address;

                    successDiv.textContent = 'Keys generated and encrypted. Registering...';
                } catch (err) {
                    errorDiv.textContent = 'Failed to generate keys: ' + err.message;
                    return;
                }
            }

            try {
                const response = await fetch('/register', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(requestData)
                });

                const result = await response.json();

                if (response.ok && result.user_id) {
                    successDiv.textContent = 'Registration successful! Redirecting to login...';
                    setTimeout(() => window.location.href = '/login', 1500);
                } else {
                    errorDiv.textContent = result.error || 'Registration failed';
                    successDiv.textContent = '';
                }
            } catch (err) {
                errorDiv.textContent = 'Network error: ' + err.message;
                successDiv.textContent = '';
            }
        });
    </script>
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
	fmt.Println("--wasmx.oauth2server.handleRegister.POST--")

	// Parse form data or JSON
	var r RegisterUserRequest
	contentType := req.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		form, _ := url.ParseQuery(string(req.Data))
		r.Email = form.Get("email")
		r.Password = form.Get("password")
		r.Username = form.Get("username")
		r.PublicKey = form.Get("public_key")
		r.Address = form.Get("address")
	} else {
		if err := json.Unmarshal(req.Data, &r); err != nil {
			return badRequest("invalid json")
		}
	}

	resp := RegisterUser(r)

	// Parse response to check for errors
	var registerResp map[string]interface{}
	json.Unmarshal(resp, &registerResp)

	// If registration failed, return error
	if _, hasError := registerResp["error"]; hasError {
		return marshalHTTP(wasmxhttp.HttpResponseWrap{
			Error: "",
			Data: wasmxhttp.HttpResponse{
				Status:     "400 Bad Request",
				StatusCode: 400,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Data:       resp,
			},
		})
	}

	// If form submission, redirect to login
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return marshalHTTP(wasmxhttp.HttpResponseWrap{
			Error: "",
			Data: wasmxhttp.HttpResponse{
				Status:      "302 Found",
				StatusCode:  302,
				Header:      http.Header{"Location": []string{"/login"}},
				RedirectUrl: "/login",
			},
		})
	}

	// JSON request - return JSON response
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
	// Parse URL to get redirect parameter
	u, _ := url.Parse(req.Url)
	q := u.Query()
	redirectURL := q.Get("redirect")

	if req.Method == "GET" {
		// Return HTML login form with hidden redirect field and WasmX blockchain option
		// Escape HTML to prevent XSS and format string issues
		escapedRedirect := strings.ReplaceAll(redirectURL, `"`, `&quot;`)
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Login</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
        input { width: 100%; padding: 10px; margin: 10px 0; box-sizing: border-box; }
        button { width: 100%; padding: 10px; background: #007bff; color: white; border: none; cursor: pointer; margin: 5px 0; }
        button:hover { background: #0056b3; }
        button.wasmx-btn { background: #28a745; }
        button.wasmx-btn:hover { background: #218838; }
        .error { color: red; margin: 10px 0; }
        .success { color: green; margin: 10px 0; }
        .blockchain-section { margin-top: 20px; padding-top: 20px; border-top: 2px solid #ddd; display: none; }
        .hidden { display: none; }
        label { display: block; margin-top: 10px; }
    </style>
</head>
<body>
    <h2>Login</h2>
    <div id="error" class="error"></div>
    <div id="success" class="success"></div>

    <form id="loginForm" method="POST" action="/login">
        <input type="email" id="email" name="email" placeholder="Email" required />
        <input type="password" id="password" name="password" placeholder="Password" required />
        <input type="hidden" id="redirect" name="redirect" value="` + escapedRedirect + `" />
        <button type="submit">Login with Password</button>
    </form>

    <div id="blockchainSection" class="blockchain-section">
        <h3>WasmX Blockchain Login</h3>
        <p>You have blockchain keys registered for this account.</p>
        <label for="pin">Enter your 4-digit PIN:</label>
        <input type="password" id="pin" pattern="[0-9]{4}" maxlength="4" placeholder="4-digit PIN" />
        <button id="wasmxBtn" type="button" class="wasmx-btn">
            Login with WasmX
        </button>
    </div>

    <p style="margin-top: 20px; text-align: center;">
        Don't have an account? <a href="/register">Register here</a>
    </p>

    <script type="module">
        import { sha256 } from 'https://cdn.jsdelivr.net/npm/@noble/hashes@1.3.3/sha256/+esm';

        // Check if user has blockchain keys when email is entered
        function checkBlockchainKeys() {
            const email = document.getElementById('email').value;
            const blockchainSection = document.getElementById('blockchainSection');

            if (email && localStorage.getItem('wasmx_encrypted_key_' + email)) {
                blockchainSection.style.display = 'block';
            } else {
                blockchainSection.style.display = 'none';
            }
        }

        // Derive decryption key from PIN using PBKDF2
        async function deriveKeyFromPIN(pin) {
            const enc = new TextEncoder();
            const pinData = enc.encode(pin);
            const salt = enc.encode('wasmx-identity-v1');

            const keyMaterial = await crypto.subtle.importKey(
                'raw',
                pinData,
                'PBKDF2',
                false,
                ['deriveBits', 'deriveKey']
            );

            return await crypto.subtle.deriveKey(
                {
                    name: 'PBKDF2',
                    salt: salt,
                    iterations: 100000,
                    hash: 'SHA-256'
                },
                keyMaterial,
                { name: 'AES-GCM', length: 256 },
                true,
                ['encrypt', 'decrypt']
            );
        }

        // Decrypt private key with PIN-derived key
        async function decryptPrivateKey(encryptedHex, pin) {
            const key = await deriveKeyFromPIN(pin);
            const encryptedBytes = new Uint8Array(encryptedHex.match(/.{1,2}/g).map(byte => parseInt(byte, 16)));
            const iv = encryptedBytes.slice(0, 12);
            const ciphertext = encryptedBytes.slice(12);

            try {
                const decrypted = await crypto.subtle.decrypt(
                    { name: 'AES-GCM', iv: iv },
                    key,
                    ciphertext
                );

                const dec = new TextDecoder();
                return dec.decode(decrypted);
            } catch (err) {
                throw new Error('Invalid PIN or corrupted key data');
            }
        }

        // Handle traditional login
        async function handleTraditionalLogin(event) {
            event.preventDefault();
            const form = document.getElementById('loginForm');
            const errorDiv = document.getElementById('error');
            const successDiv = document.getElementById('success');

            errorDiv.textContent = '';
            successDiv.textContent = '';

            const formData = new FormData(form);

            try {
                const response = await fetch('/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                    body: new URLSearchParams(formData)
                });

                if (response.redirected) {
                    window.location.href = response.url;
                } else if (response.ok) {
                    const redirect = formData.get('redirect');
                    if (redirect) {
                        window.location.href = redirect;
                    } else {
                        successDiv.textContent = 'Login successful!';
                        setTimeout(() => window.location.href = '/', 1000);
                    }
                } else {
                    const data = await response.json();
                    errorDiv.textContent = data.error || 'Login failed';
                }
            } catch (err) {
                errorDiv.textContent = 'Network error: ' + err.message;
            }
        }

        // Handle WasmX blockchain login
        async function handleWasmXLogin(event) {
            event.preventDefault();
            const errorDiv = document.getElementById('error');
            const successDiv = document.getElementById('success');

            errorDiv.textContent = '';
            successDiv.textContent = '';

            const email = document.getElementById('email').value;
            const pin = document.getElementById('pin').value;

            if (!email || !pin) {
                errorDiv.textContent = 'Email and PIN are required for WasmX login';
                return;
            }

            if (!/^[0-9]{4}$/.test(pin)) {
                errorDiv.textContent = 'PIN must be exactly 4 digits';
                return;
            }

            const encryptedKey = localStorage.getItem('wasmx_encrypted_key_' + email);
            if (!encryptedKey) {
                errorDiv.textContent = 'No blockchain keys found for this account';
                return;
            }

            try {
                successDiv.textContent = 'Decrypting private key...';
                const privateKey = await decryptPrivateKey(encryptedKey, pin);

                successDiv.textContent = 'Private key decrypted! Signing in...';
                sessionStorage.setItem('wasmx_private_key', privateKey);

                const formData = new URLSearchParams();
                formData.append('email', email);
                formData.append('password', document.getElementById('password').value);
                formData.append('redirect', document.getElementById('redirect').value);
                formData.append('use_blockchain', 'true');

                const response = await fetch('/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                    body: formData
                });

                if (response.redirected) {
                    window.location.href = response.url;
                } else if (response.ok) {
                    const redirect = document.getElementById('redirect').value;
                    if (redirect) {
                        window.location.href = redirect;
                    } else {
                        successDiv.textContent = 'WasmX login successful!';
                        setTimeout(() => window.location.href = '/', 1000);
                    }
                } else {
                    const data = await response.json();
                    errorDiv.textContent = data.error || 'Login failed';
                }
            } catch (err) {
                errorDiv.textContent = err.message;
            }
        }

        // Attach event listeners
        document.getElementById('email').addEventListener('input', checkBlockchainKeys);
        document.getElementById('loginForm').addEventListener('submit', handleTraditionalLogin);
        document.getElementById('wasmxBtn').addEventListener('click', handleWasmXLogin);

        // Check on page load if email field has value
        checkBlockchainKeys();
    </script>
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
	var formRedirect string
	var useBlockchain bool
	contentType := req.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		form, _ := url.ParseQuery(string(req.Data))
		r.Email = form.Get("email")
		r.Password = form.Get("password")
		formRedirect = form.Get("redirect")
		useBlockchain = form.Get("use_blockchain") == "true"
	} else {
		if err := json.Unmarshal(req.Data, &r); err != nil {
			return badRequest("invalid json")
		}
	}

	// If using blockchain login, the private key was decrypted in the browser
	// and stored in sessionStorage. The actual authentication still uses the
	// password for now, but we can enhance this later to support pure blockchain auth.
	if useBlockchain {
		LoggerInfo("WasmX blockchain login requested", []string{"email", r.Email})
	}

	resp := Login(r)

	// Parse response to check for errors and get session
	var loginResp LoginResponse
	json.Unmarshal(resp, &loginResp)

	// If login failed, return error
	if loginResp.SessionID == "" {
		return marshalHTTP(wasmxhttp.HttpResponseWrap{
			Error: "",
			Data: wasmxhttp.HttpResponse{
				Status:     "401 Unauthorized",
				StatusCode: 401,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Data:       resp,
			},
		})
	}

	// Set session cookie
	cookie := fmt.Sprintf("session_id=%s; Path=/; HttpOnly; Max-Age=%d", loginResp.SessionID, SESSION_DURATION_SECONDS)

	// If there's a redirect URL, redirect there (for OAuth flow)
	if formRedirect != "" {
		headers := http.Header{
			"Set-Cookie": []string{cookie},
			"Location":   []string{formRedirect},
		}
		return marshalHTTP(wasmxhttp.HttpResponseWrap{
			Error: "",
			Data: wasmxhttp.HttpResponse{
				Status:      "302 Found",
				StatusCode:  302,
				Header:      headers,
				RedirectUrl: formRedirect,
			},
		})
	}

	// No redirect - return JSON response (for API calls)
	headers := http.Header{
		"Content-Type": []string{"application/json"},
		"Set-Cookie":   []string{cookie},
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
				Status:      "302 Found",
				StatusCode:  302,
				Header:      http.Header{"Location": []string{loginURL}},
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
				Status:      "302 Found",
				StatusCode:  302,
				Header:      http.Header{"Location": []string{loginURL}},
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
				Status:      "302 Found",
				StatusCode:  302,
				Header:      http.Header{"Location": []string{rurl.String()}},
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
