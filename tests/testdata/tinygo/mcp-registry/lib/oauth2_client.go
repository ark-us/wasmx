package lib

import (
	"encoding/json"
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// OAuth2 server contract call helpers

// getOAuth2ServerAddress gets the OAuth2 server contract address by role
func getOAuth2ServerAddress() wasmx.Bech32String {
	return wasmx.GetAddressByRole(wasmx.ROLE_OAUTH2_SERVER)
}

// ValidateSession calls OAuth2 server to validate a session and get user_id
func callOAuth2ValidateSession(sessionID string) (string, error) {
	req := map[string]interface{}{
		"validate_session": map[string]string{
			"session_id": sessionID,
		},
	}

	calldata, _ := json.Marshal(req)
	value := sdkmath.NewInt(0)
	gas := big.NewInt(10000000)

	success, responseData := wasmx.Call(
		getOAuth2ServerAddress(),
		&value,
		calldata,
		gas,
		MODULE_NAME,
	)

	if !success {
		return "", fmt.Errorf("OAuth2 server call failed")
	}

	var response struct {
		Valid  bool   `json:"valid"`
		UserID string `json:"user_id,omitempty"`
	}

	if err := json.Unmarshal(responseData, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %s", err.Error())
	}

	if !response.Valid {
		return "", fmt.Errorf("invalid session")
	}

	return response.UserID, nil
}

// ValidateAccessToken calls OAuth2 server to validate an access token and get user_id
func callOAuth2ValidateAccessToken(token string) (string, error) {
	fmt.Println("=== callOAuth2ValidateAccessToken ===")
	fmt.Println("Token:", token)

	oauth2Addr := getOAuth2ServerAddress()
	fmt.Println("OAuth2 server address:", string(oauth2Addr))

	req := map[string]interface{}{
		"validate_access_token": map[string]string{
			"token": token,
		},
	}

	calldata, _ := json.Marshal(req)
	fmt.Println("Request calldata:", string(calldata))

	value := sdkmath.NewInt(0)
	gas := big.NewInt(10000000)

	success, responseData := wasmx.Call(
		oauth2Addr,
		&value,
		calldata,
		gas,
		MODULE_NAME,
	)

	fmt.Println("Call success:", success)
	fmt.Println("Response data:", string(responseData))

	if !success {
		fmt.Println("OAuth2 server call failed")
		return "", fmt.Errorf("OAuth2 server call failed")
	}

	var response struct {
		Valid  bool   `json:"valid"`
		UserID string `json:"user_id,omitempty"`
	}

	if err := json.Unmarshal(responseData, &response); err != nil {
		fmt.Println("Failed to parse response:", err.Error())
		return "", fmt.Errorf("failed to parse response: %s", err.Error())
	}

	fmt.Printf("Response parsed: Valid=%v, UserID=%s\n", response.Valid, response.UserID)

	if !response.Valid {
		fmt.Println("Token marked as invalid by OAuth2 server")
		return "", fmt.Errorf("invalid token")
	}

	return response.UserID, nil
}

// CreateAuthorizationCode calls OAuth2 server to create an authorization code
func callOAuth2CreateAuthorizationCode(clientID, userID, redirectURI string, scopes []string, codeChallenge, codeChallengeMethod string) (string, error) {
	req := map[string]interface{}{
		"create_authorization_code": map[string]interface{}{
			"client_id":             clientID,
			"user_id":               userID,
			"redirect_uri":          redirectURI,
			"scopes":                scopes,
			"code_challenge":        codeChallenge,
			"code_challenge_method": codeChallengeMethod,
		},
	}

	calldata, _ := json.Marshal(req)
	value := sdkmath.NewInt(0)
	gas := big.NewInt(10000000)

	success, responseData := wasmx.Call(
		getOAuth2ServerAddress(),
		&value,
		calldata,
		gas,
		MODULE_NAME,
	)

	if !success {
		return "", fmt.Errorf("OAuth2 server call failed: %s", string(responseData))
	}

	var response struct {
		Code      string `json:"code"`
		ExpiresAt int64  `json:"expires_at"`
	}

	if err := json.Unmarshal(responseData, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %s", err.Error())
	}

	return response.Code, nil
}

// ExchangeCodeForToken calls OAuth2 server to exchange authorization code for tokens
func callOAuth2ExchangeCodeForToken(code, clientID, clientSecret, redirectURI, codeVerifier string) (*TokenResponse, error) {
	req := map[string]interface{}{
		"exchange_code_for_token": map[string]string{
			"code":          code,
			"client_id":     clientID,
			"client_secret": clientSecret,
			"redirect_uri":  redirectURI,
			"code_verifier": codeVerifier,
		},
	}

	calldata, _ := json.Marshal(req)
	value := sdkmath.NewInt(0)
	gas := big.NewInt(10000000)

	success, responseData := wasmx.Call(
		getOAuth2ServerAddress(),
		&value,
		calldata,
		gas,
		MODULE_NAME,
	)

	if !success {
		return nil, fmt.Errorf("OAuth2 server call failed: %s", string(responseData))
	}

	var response TokenResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %s", err.Error())
	}

	return &response, nil
}

// RefreshAccessToken calls OAuth2 server to refresh an access token
func callOAuth2RefreshAccessToken(refreshToken, clientID, clientSecret string) (*TokenResponse, error) {
	req := map[string]interface{}{
		"refresh_access_token": map[string]string{
			"refresh_token": refreshToken,
			"client_id":     clientID,
			"client_secret": clientSecret,
		},
	}

	calldata, _ := json.Marshal(req)
	value := sdkmath.NewInt(0)
	gas := big.NewInt(10000000)

	success, responseData := wasmx.Call(
		getOAuth2ServerAddress(),
		&value,
		calldata,
		gas,
		MODULE_NAME,
	)

	if !success {
		return nil, fmt.Errorf("OAuth2 server call failed: %s", string(responseData))
	}

	var response TokenResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %s", err.Error())
	}

	return &response, nil
}

// RegisterUser forwards user registration to OAuth2 server
func callOAuth2RegisterUser(email, password, username string) ([]byte, error) {
	req := map[string]interface{}{
		"register_user": map[string]string{
			"email":    email,
			"password": password,
			"username": username,
		},
	}

	calldata, _ := json.Marshal(req)
	value := sdkmath.NewInt(0)
	gas := big.NewInt(10000000)

	success, responseData := wasmx.Call(
		getOAuth2ServerAddress(),
		&value,
		calldata,
		gas,
		MODULE_NAME,
	)

	if !success {
		return responseData, fmt.Errorf("OAuth2 server call failed")
	}

	return responseData, nil
}

// Login forwards login to OAuth2 server
func callOAuth2Login(email, password string) ([]byte, error) {
	req := map[string]interface{}{
		"login": map[string]string{
			"email":    email,
			"password": password,
		},
	}

	calldata, _ := json.Marshal(req)
	value := sdkmath.NewInt(0)
	gas := big.NewInt(10000000)

	success, responseData := wasmx.Call(
		getOAuth2ServerAddress(),
		&value,
		calldata,
		gas,
		MODULE_NAME,
	)

	if !success {
		return responseData, fmt.Errorf("OAuth2 server call failed")
	}

	return responseData, nil
}

// Logout forwards logout to OAuth2 server
func callOAuth2Logout(sessionID string) ([]byte, error) {
	req := map[string]interface{}{
		"logout": map[string]string{
			"session_id": sessionID,
		},
	}

	calldata, _ := json.Marshal(req)
	value := sdkmath.NewInt(0)
	gas := big.NewInt(10000000)

	success, responseData := wasmx.Call(
		getOAuth2ServerAddress(),
		&value,
		calldata,
		gas,
		MODULE_NAME,
	)

	if !success {
		return responseData, fmt.Errorf("OAuth2 server call failed")
	}

	return responseData, nil
}

// GetCurrentUser forwards get current user to OAuth2 server
func callOAuth2GetCurrentUser(sessionID string) ([]byte, error) {
	req := map[string]interface{}{
		"get_current_user": map[string]string{
			"session_id": sessionID,
		},
	}

	calldata, _ := json.Marshal(req)
	value := sdkmath.NewInt(0)
	gas := big.NewInt(10000000)

	success, responseData := wasmx.Call(
		getOAuth2ServerAddress(),
		&value,
		calldata,
		gas,
		MODULE_NAME,
	)

	if !success {
		return responseData, fmt.Errorf("OAuth2 server call failed")
	}

	return responseData, nil
}
