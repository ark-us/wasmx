package oauth2server

// #include <stdlib.h>
import "C"

import (
	"encoding/json"

	utils "github.com/loredanacirstea/wasmx-env-utils"
)

//go:wasmimport oauth2server Initialize
func Initialize_(reqPtr int64) int64

//go:wasmimport oauth2server RegisterClient
func RegisterClient_(reqPtr int64) int64

//go:wasmimport oauth2server UpdateClient
func UpdateClient_(reqPtr int64) int64

//go:wasmimport oauth2server DeleteClient
func DeleteClient_(reqPtr int64) int64

//go:wasmimport oauth2server GetClient
func GetClient_(reqPtr int64) int64

//go:wasmimport oauth2server ValidateClient
func ValidateClient_(reqPtr int64) int64

//go:wasmimport oauth2server CreateAuthorizationCode
func CreateAuthorizationCode_(reqPtr int64) int64

//go:wasmimport oauth2server ValidateAuthorizationCode
func ValidateAuthorizationCode_(reqPtr int64) int64

//go:wasmimport oauth2server IssueAccessToken
func IssueAccessToken_(reqPtr int64) int64

//go:wasmimport oauth2server IssueRefreshToken
func IssueRefreshToken_(reqPtr int64) int64

//go:wasmimport oauth2server ValidateAccessToken
func ValidateAccessToken_(reqPtr int64) int64

//go:wasmimport oauth2server RefreshAccessToken
func RefreshAccessToken_(reqPtr int64) int64

//go:wasmimport oauth2server RevokeToken
func RevokeToken_(reqPtr int64) int64

//go:wasmimport oauth2server IntrospectToken
func IntrospectToken_(reqPtr int64) int64

//go:wasmimport oauth2server CleanupExpiredTokens
func CleanupExpiredTokens_(reqPtr int64) int64

func Initialize(req *InitializeRequest) InitializeResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := Initialize_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp InitializeResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func RegisterClient(req *RegisterClientRequest) RegisterClientResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := RegisterClient_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp RegisterClientResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func UpdateClient(req *UpdateClientRequest) UpdateClientResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := UpdateClient_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp UpdateClientResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func DeleteClient(req *DeleteClientRequest) DeleteClientResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := DeleteClient_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp DeleteClientResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func GetClient(req *GetClientRequest) GetClientResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := GetClient_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp GetClientResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func ValidateClient(req *ValidateClientRequest) ValidateClientResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := ValidateClient_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp ValidateClientResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func CreateAuthorizationCode(req *CreateAuthorizationCodeRequest) CreateAuthorizationCodeResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := CreateAuthorizationCode_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp CreateAuthorizationCodeResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func ValidateAuthorizationCode(req *ValidateAuthorizationCodeRequest) ValidateAuthorizationCodeResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := ValidateAuthorizationCode_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp ValidateAuthorizationCodeResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func IssueAccessToken(req *IssueAccessTokenRequest) IssueAccessTokenResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := IssueAccessToken_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp IssueAccessTokenResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func IssueRefreshToken(req *IssueRefreshTokenRequest) IssueRefreshTokenResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := IssueRefreshToken_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp IssueRefreshTokenResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func ValidateAccessToken(req *ValidateAccessTokenRequest) ValidateAccessTokenResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := ValidateAccessToken_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp ValidateAccessTokenResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func RefreshAccessToken(req *RefreshAccessTokenRequest) RefreshAccessTokenResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := RefreshAccessToken_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp RefreshAccessTokenResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func RevokeToken(req *RevokeTokenRequest) RevokeTokenResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := RevokeToken_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp RevokeTokenResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func IntrospectToken(req *IntrospectTokenRequest) IntrospectTokenResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := IntrospectToken_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp IntrospectTokenResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func CleanupExpiredTokens(req *CleanupExpiredTokensRequest) CleanupExpiredTokensResponse {
	reqbz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	reqPtr := utils.BytesToPackedPtr(reqbz)
	ptr := CleanupExpiredTokens_(reqPtr)
	bz := utils.PackedPtrToBytes(ptr)
	var resp CleanupExpiredTokensResponse
	err = json.Unmarshal(bz, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}
