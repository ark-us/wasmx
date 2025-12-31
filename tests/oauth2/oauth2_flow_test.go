package keeper_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
)

// OAuth2 response types
type LoginResponseOAuth2 struct {
	SessionID         string `json:"session_id"`
	UserID            string `json:"user_id"`
	Email             string `json:"email"`
	ExpiresAt         int64  `json:"expires_at"`
	IdentityUserID    string `json:"identity_user_id,omitempty"`
	BlockchainAddress string `json:"blockchain_address,omitempty"`
}

type LogoutResponseOAuth2 struct {
	Success bool `json:"success"`
}

type ExchangeCodeForTokenResponseOAuth2 struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// Test configuration
const (
	testServerAddr = "http://localhost:8080"
	testEmail      = "test@example.com"
	testPassword   = "testpassword123"
	testUsername   = "testuser"
)

// setupOAuth2Server sets up the OAuth2 server and returns the client credentials
func (suite *KeeperTestSuite) setupOAuth2Server() (clientID, clientSecret string) {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	// Configure OAuth2 Keys contract with funder private key
	oauth2KeysAddr := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_OAUTH2_KEYS))
	funderPrivKeyHex := fmt.Sprintf("%x", sender.PrivKey.Bytes())

	oauth2KeysInitMsg := map[string]interface{}{
		"init_genesis": map[string]interface{}{
			"funder_priv_key": funderPrivKeyHex,
			"init_account_amt": map[string]string{
				"amount": "1000000",
				"denom":  "amyt",
			},
			"gas_price": map[string]string{
				"amount": "10",
				"denom":  "amyt",
			},
			"route_prefix": "/auth",
		},
	}
	oauth2KeysInitData, err := json.Marshal(oauth2KeysInitMsg)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, oauth2KeysAddr, types.WasmxExecutionMessage{Data: oauth2KeysInitData}, nil, nil)

	// Register OAuth client
	oauth2Addr := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_OAUTH2_SERVER))
	registerClientMsg := &RegisterOAuthClientCalldata{
		RegisterOAuthClient: &RegisterOAuthClientRequest{
			Name:        "Test Client",
			Description: "OAuth client for testing",
			RedirectURIs: []string{
				"http://localhost:3000/callback",
				"http://localhost:8080/callback",
			},
			Scopes: []string{"read", "write", "tools"},
		},
	}
	registerClientData, err := json.Marshal(registerClientMsg)
	suite.Require().NoError(err)
	clientResp := appA.ExecuteContract(sender, oauth2Addr, types.WasmxExecutionMessage{Data: registerClientData}, nil, nil)

	var clientResult RegisterOAuthClientResponse
	err = appA.DecodeExecuteResponse(clientResp, &clientResult)
	suite.Require().NoError(err)

	clientID = clientResult.ClientID
	clientSecret = clientResult.ClientSecret

	// Configure HTTP server registry contract
	httpRegistryAddr := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_HTTPSERVER_REGISTRY))
	httpRegistryInitMsg := map[string]interface{}{
		"init_genesis": map[string]interface{}{
			"gas_price": map[string]string{
				"amount": "10",
				"denom":  "amyt",
			},
		},
	}
	httpRegistryInitData, err := json.Marshal(httpRegistryInitMsg)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, httpRegistryAddr, types.WasmxExecutionMessage{Data: httpRegistryInitData}, nil, nil)

	// Start the HTTP server
	startServerMsg := &StartWebServerCalldata{
		StartWebServer: &StartWebServerRequest{
			Config: WebsrvConfig{
				EnableOAuth:        true,
				Address:            "0.0.0.0:8080",
				CORSAllowedOrigins: []string{"*"},
				CORSAllowedMethods: []string{},
				CORSAllowedHeaders: []string{},
				MaxOpenConnections: 1000,
				RequestBodyMaxSize: 1000000000,
			},
		},
	}
	startServerData, err := json.Marshal(startServerMsg)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, httpRegistryAddr, types.WasmxExecutionMessage{Data: startServerData}, nil, nil)

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	return clientID, clientSecret
}

// TestOAuth2Flow tests the complete OAuth2 registration, login, and logout flow
func (suite *KeeperTestSuite) TestOAuth2Flow() {
	clientID, clientSecret := suite.setupOAuth2Server()

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	suite.T().Log("=== Testing OAuth2 Flow ===")

	// ========================================
	// PART 1: User Registration
	// ========================================
	suite.T().Log("--- Part 1: User Registration ---")

	// Test 1.1: Register a new user via POST /register
	registerData := map[string]string{
		"email":    testEmail,
		"password": testPassword,
		"username": testUsername,
	}
	registerBody, _ := json.Marshal(registerData)

	resp, err := client.Post(
		testServerAddr+"/register",
		"application/json",
		bytes.NewBuffer(registerBody),
	)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	suite.T().Logf("Register response: %s", string(body))

	suite.Require().Equal(http.StatusOK, resp.StatusCode, "Registration should succeed")

	var registerResp RegisterUserResponse
	err = json.Unmarshal(body, &registerResp)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(registerResp.UserID, "User ID should be returned")
	suite.Require().Equal(testEmail, registerResp.Email, "Email should match")

	suite.T().Logf("User registered successfully: UserID=%s", registerResp.UserID)

	// Test 1.2: Attempt to register with the same email (should fail)
	resp2, err := client.Post(
		testServerAddr+"/register",
		"application/json",
		bytes.NewBuffer(registerBody),
	)
	suite.Require().NoError(err)
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()

	suite.T().Logf("Duplicate registration response: status=%d body=%s", resp2.StatusCode, string(body2))
	suite.Require().Equal(http.StatusBadRequest, resp2.StatusCode, "Duplicate registration should return 400")
	suite.Require().Contains(string(body2), "email already registered", "Duplicate registration should return error message")

	// Test 1.3: Register with invalid data (missing email)
	invalidData := map[string]string{
		"password": testPassword,
	}
	invalidBody, _ := json.Marshal(invalidData)

	resp3, err := client.Post(
		testServerAddr+"/register",
		"application/json",
		bytes.NewBuffer(invalidBody),
	)
	suite.Require().NoError(err)
	body3, _ := io.ReadAll(resp3.Body)
	resp3.Body.Close()
	suite.T().Logf("Registration without email response: status=%d body=%s", resp3.StatusCode, string(body3))
	suite.Require().Equal(http.StatusBadRequest, resp3.StatusCode, "Registration without email should return 400")

	// Test 1.4: Register with short password
	shortPwdData := map[string]string{
		"email":    "another@example.com",
		"password": "short",
	}
	shortPwdBody, _ := json.Marshal(shortPwdData)

	resp4, err := client.Post(
		testServerAddr+"/register",
		"application/json",
		bytes.NewBuffer(shortPwdBody),
	)
	suite.Require().NoError(err)
	body4, _ := io.ReadAll(resp4.Body)
	resp4.Body.Close()
	suite.T().Logf("Registration with short password response: status=%d body=%s", resp4.StatusCode, string(body4))
	suite.Require().Equal(http.StatusBadRequest, resp4.StatusCode, "Registration with short password should return 400")

	suite.T().Log("Registration tests passed!")

	// ========================================
	// PART 2: User Login
	// ========================================
	suite.T().Log("--- Part 2: User Login ---")

	// Test 2.1: Login with valid credentials
	loginData := url.Values{
		"email":    {testEmail},
		"password": {testPassword},
	}

	resp, err = client.PostForm(testServerAddr+"/login", loginData)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	suite.T().Logf("Login response: %s", string(body))

	suite.Require().Equal(http.StatusOK, resp.StatusCode, "Login should succeed")

	var loginResp LoginResponseOAuth2
	err = json.Unmarshal(body, &loginResp)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(loginResp.SessionID, "Session ID should be returned")
	suite.Require().Equal(testEmail, loginResp.Email, "Email should match")

	sessionID := loginResp.SessionID
	suite.T().Logf("User logged in successfully: SessionID=%s", sessionID)

	// Check for session cookie
	var sessionCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session_id" {
			sessionCookie = cookie
			break
		}
	}
	suite.Require().NotNil(sessionCookie, "Session cookie should be set")
	suite.Require().Equal(sessionID, sessionCookie.Value, "Cookie session ID should match response")

	// Test 2.2: Login with invalid password
	invalidLoginData := url.Values{
		"email":    {testEmail},
		"password": {"wrongpassword"},
	}

	resp, err = client.PostForm(testServerAddr+"/login", invalidLoginData)
	suite.Require().NoError(err)
	resp.Body.Close()
	suite.Require().Equal(http.StatusUnauthorized, resp.StatusCode, "Login with wrong password should fail")

	// Test 2.3: Login with non-existent email
	nonExistentLoginData := url.Values{
		"email":    {"nonexistent@example.com"},
		"password": {testPassword},
	}

	resp, err = client.PostForm(testServerAddr+"/login", nonExistentLoginData)
	suite.Require().NoError(err)
	resp.Body.Close()
	suite.Require().Equal(http.StatusUnauthorized, resp.StatusCode, "Login with non-existent email should fail")

	suite.T().Log("Login tests passed!")

	// ========================================
	// PART 3: OAuth2 Authorization Code Flow
	// ========================================
	suite.T().Log("--- Part 3: OAuth2 Authorization Code Flow ---")

	// Create a cookie jar to maintain session across requests
	jar := &testCookieJar{cookies: make(map[string][]*http.Cookie)}
	oauthClient := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Store session cookie for OAuth flow
	jar.SetCookies(
		&url.URL{Scheme: "http", Host: "localhost:8080"},
		[]*http.Cookie{{Name: "session_id", Value: sessionID}},
	)

	// Generate PKCE code verifier and challenge
	codeVerifier := "test_code_verifier_string_that_is_at_least_43_chars"
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])

	// Initiate authorization request
	authParams := url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {"http://localhost:3000/callback"},
		"response_type":         {"code"},
		"state":                 {"test_state_123"},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
	}

	authURL := testServerAddr + "/oauth/authorize?" + authParams.Encode()
	resp, err = oauthClient.Get(authURL)
	suite.Require().NoError(err)

	suite.T().Logf("Authorization response status: %d", resp.StatusCode)

	// If we get a 200, we need to POST to approve
	if resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		resp, err = oauthClient.PostForm(authURL, url.Values{})
		suite.Require().NoError(err)
	}
	defer resp.Body.Close()

	// Should redirect to callback with authorization code
	suite.Require().Equal(http.StatusFound, resp.StatusCode, "Should redirect after authorization")

	location := resp.Header.Get("Location")
	suite.Require().NotEmpty(location, "Redirect location should be set")
	suite.T().Logf("Redirect location: %s", location)

	// Parse the redirect URL to get the authorization code
	redirectURL, err := url.Parse(location)
	suite.Require().NoError(err)

	authCode := redirectURL.Query().Get("code")
	state := redirectURL.Query().Get("state")

	suite.Require().NotEmpty(authCode, "Authorization code should be returned")
	suite.Require().Equal("test_state_123", state, "State should be preserved")

	suite.T().Logf("Authorization code received: %s", authCode)

	// Exchange authorization code for tokens
	tokenData := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {authCode},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {"http://localhost:3000/callback"},
		"code_verifier": {codeVerifier},
	}

	resp, err = oauthClient.PostForm(testServerAddr+"/oauth/token", tokenData)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	suite.T().Logf("Token exchange response: %s", string(body))

	suite.Require().Equal(http.StatusOK, resp.StatusCode, "Token exchange should succeed")

	var tokenResp ExchangeCodeForTokenResponseOAuth2
	err = json.Unmarshal(body, &tokenResp)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(tokenResp.AccessToken, "Access token should be returned")
	suite.Require().NotEmpty(tokenResp.RefreshToken, "Refresh token should be returned")
	suite.Require().Equal("Bearer", tokenResp.TokenType, "Token type should be Bearer")

	suite.T().Logf("Tokens received - Access: %s..., Refresh: %s...",
		tokenResp.AccessToken[:min(20, len(tokenResp.AccessToken))],
		tokenResp.RefreshToken[:min(20, len(tokenResp.RefreshToken))])

	// Try to reuse the authorization code (should fail)
	resp, err = oauthClient.PostForm(testServerAddr+"/oauth/token", tokenData)
	suite.Require().NoError(err)
	resp.Body.Close()
	suite.Require().NotEqual(http.StatusOK, resp.StatusCode, "Reusing authorization code should fail")

	// Refresh the access token
	refreshData := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {tokenResp.RefreshToken},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}

	resp, err = oauthClient.PostForm(testServerAddr+"/oauth/token", refreshData)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	suite.T().Logf("Token refresh response: %s", string(body))

	suite.Require().Equal(http.StatusOK, resp.StatusCode, "Token refresh should succeed")

	var refreshResp ExchangeCodeForTokenResponseOAuth2
	err = json.Unmarshal(body, &refreshResp)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(refreshResp.AccessToken, "New access token should be returned")
	suite.Require().NotEqual(tokenResp.AccessToken, refreshResp.AccessToken, "Access token should be different after refresh")

	suite.T().Log("OAuth2 authorization code flow tests passed!")

	// ========================================
	// PART 4: User Logout
	// ========================================
	suite.T().Log("--- Part 4: User Logout ---")

	logoutData := map[string]string{
		"session_id": sessionID,
	}
	logoutBody, _ := json.Marshal(logoutData)

	resp, err = client.Post(
		testServerAddr+"/logout",
		"application/json",
		bytes.NewBuffer(logoutBody),
	)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	suite.T().Logf("Logout response: %s", string(body))

	suite.Require().Equal(http.StatusOK, resp.StatusCode, "Logout should succeed")

	var logoutResp LogoutResponseOAuth2
	err = json.Unmarshal(body, &logoutResp)
	suite.Require().NoError(err)
	suite.Require().True(logoutResp.Success, "Logout should return success")

	suite.T().Log("Logout tests passed!")

	// ========================================
	// PART 5: Well-Known Endpoints
	// ========================================
	suite.T().Log("--- Part 5: Well-Known Endpoints ---")

	endpoints := []string{
		"/.well-known/oauth-authorization-server",
		"/.well-known/openid-configuration",
		"/.well-known/oauth-protected-resource",
	}

	for _, endpoint := range endpoints {
		resp, err := client.Get(testServerAddr + endpoint)
		suite.Require().NoError(err)

		suite.Require().Equal(http.StatusOK, resp.StatusCode, "Well-known endpoint %s should return 200", endpoint)

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		suite.T().Logf("Well-known %s response: %s", endpoint, string(body))

		var wellKnown map[string]interface{}
		err = json.Unmarshal(body, &wellKnown)
		suite.Require().NoError(err)

		suite.Require().NotEmpty(wellKnown["issuer"], "issuer should be present")
		suite.Require().NotEmpty(wellKnown["authorization_endpoint"], "authorization_endpoint should be present")
		suite.Require().NotEmpty(wellKnown["token_endpoint"], "token_endpoint should be present")
	}

	suite.T().Log("Well-known endpoint tests passed!")

	// ========================================
	// PART 6: Contract Addresses Endpoint
	// ========================================
	suite.T().Log("--- Part 6: Contract Addresses Endpoint ---")

	resp, err = client.Get(testServerAddr + "/auth/contract_addresses")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Require().Equal(http.StatusOK, resp.StatusCode, "Contract addresses endpoint should return 200")

	body, _ = io.ReadAll(resp.Body)
	suite.T().Logf("Contract addresses response: %s", string(body))

	var addresses map[string]interface{}
	err = json.Unmarshal(body, &addresses)
	suite.Require().NoError(err)

	suite.Require().NotEmpty(addresses["identity_address"], "identity_address should be present")
	suite.Require().NotEmpty(addresses["chain_id"], "chain_id should be present")
	suite.Require().NotEmpty(addresses["base_denom"], "base_denom should be present")
	suite.Require().NotEmpty(addresses["bech32_prefix"], "bech32_prefix should be present")

	suite.T().Log("Contract addresses endpoint tests passed!")

	suite.T().Log("=== All OAuth2 Flow Tests Passed! ===")
}

// testCookieJar is a simple cookie jar implementation for testing
type testCookieJar struct {
	cookies map[string][]*http.Cookie
}

func (j *testCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.cookies[u.Host] = append(j.cookies[u.Host], cookies...)
}

func (j *testCookieJar) Cookies(u *url.URL) []*http.Cookie {
	return j.cookies[u.Host]
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
