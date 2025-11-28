package vmoauth2server

import (
	"time"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InitializeMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &InitializeResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func RegisterClientMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &RegisterClientResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func UpdateClientMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &UpdateClientResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func DeleteClientMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &DeleteClientResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func GetClientMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &GetClientResponse{
		Error:        "",
		ClientID:     "mock_client_id",
		RedirectURIs: []string{"http://localhost/callback"},
		Scopes:       []string{"read", "write"},
		GrantTypes:   []string{"authorization_code", "refresh_token"},
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
	}
	return prepareResponse(rnh, response)
}

func ValidateClientMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &ValidateClientResponse{
		Error: "",
		Valid: true,
	}
	return prepareResponse(rnh, response)
}

func CreateAuthorizationCodeMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &CreateAuthorizationCodeResponse{
		Error: "",
		Code:  "mock_auth_code_1234567890",
	}
	return prepareResponse(rnh, response)
}

func ValidateAuthorizationCodeMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &ValidateAuthorizationCodeResponse{
		Error:  "",
		Valid:  true,
		UserID: "mock_user_id",
		Scopes: []string{"read", "write"},
	}
	return prepareResponse(rnh, response)
}

func IssueAccessTokenMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &IssueAccessTokenResponse{
		Error: "",
		Token: "mock_access_token_1234567890abcdef",
	}
	return prepareResponse(rnh, response)
}

func IssueRefreshTokenMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &IssueRefreshTokenResponse{
		Error: "",
		Token: "mock_refresh_token_1234567890abcdef",
	}
	return prepareResponse(rnh, response)
}

func ValidateAccessTokenMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &ValidateAccessTokenResponse{
		Error:     "",
		Valid:     true,
		ClientID:  "mock_client_id",
		UserID:    "mock_user_id",
		Scopes:    []string{"read", "write"},
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
	}
	return prepareResponse(rnh, response)
}

func RefreshAccessTokenMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &RefreshAccessTokenResponse{
		Error:       "",
		AccessToken: "mock_new_access_token_1234567890",
		ClientID:    "mock_client_id",
		UserID:      "mock_user_id",
		Scopes:      []string{"read", "write"},
		ExpiresAt:   time.Now().Add(1 * time.Hour).Unix(),
	}
	return prepareResponse(rnh, response)
}

func RevokeTokenMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &RevokeTokenResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func IntrospectTokenMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &IntrospectTokenResponse{
		Error:     "",
		Active:    true,
		Scope:     "read write",
		ClientID:  "mock_client_id",
		Username:  "mock_user_id",
		TokenType: "access_token",
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
		IssuedAt:  time.Now().Unix(),
	}
	return prepareResponse(rnh, response)
}

func CleanupExpiredTokensMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &CleanupExpiredTokensResponse{
		Error:                "",
		DeletedAuthCodes:     0,
		DeletedAccessTokens:  0,
		DeletedRefreshTokens: 0,
	}
	return prepareResponse(rnh, response)
}
