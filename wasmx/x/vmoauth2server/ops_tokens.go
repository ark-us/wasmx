package vmoauth2server

import (
	"encoding/json"
	"strings"
	"time"

	vmpostgresql "github.com/loredanacirstea/wasmx-vmpostgresql"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func IssueAccessToken(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req IssueAccessTokenRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	response := &IssueAccessTokenResponse{Error: ""}

	oauth2Ctx, err := GetOAuth2ServerContext(ctx.Context.GoContextParent)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	instance, found := oauth2Ctx.GetInstance(req.InstanceID)
	if !found {
		response.Error = "instance not found"
		return prepareResponse(rnh, response)
	}

	// Generate secure access token
	token, err := GenerateSecureToken(AccessTokenLength)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	expiresAt := time.Now().Add(time.Duration(req.ExpiresInSeconds) * time.Second).Unix()
	createdAt := time.Now().Unix()

	scopesJSON, _ := json.Marshal(req.Scopes)

	query := `
		INSERT INTO oauth2_access_tokens
		(token, client_id, user_id, scopes, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	queryParams := []vmpostgresql.SqlQueryParam{
		{Value: token},
		{Value: req.ClientID},
		{Value: req.UserID},
		{Value: string(scopesJSON)},
		{Value: expiresAt},
		{Value: createdAt},
	}

	err = executePgQuery(ctx, instance.ConnectionID, query, queryParams)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	response.Token = token
	return prepareResponse(rnh, response)
}

func IssueRefreshToken(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req IssueRefreshTokenRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	response := &IssueRefreshTokenResponse{Error: ""}

	oauth2Ctx, err := GetOAuth2ServerContext(ctx.Context.GoContextParent)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	instance, found := oauth2Ctx.GetInstance(req.InstanceID)
	if !found {
		response.Error = "instance not found"
		return prepareResponse(rnh, response)
	}

	// Generate secure refresh token
	token, err := GenerateSecureToken(RefreshTokenLength)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	createdAt := time.Now().Unix()
	var expiresAt interface{} = nil
	if req.ExpiresInSeconds > 0 {
		expiresAt = time.Now().Add(time.Duration(req.ExpiresInSeconds) * time.Second).Unix()
	}

	scopesJSON, _ := json.Marshal(req.Scopes)

	query := `
		INSERT INTO oauth2_refresh_tokens
		(token, client_id, user_id, scopes, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	queryParams := []vmpostgresql.SqlQueryParam{
		{Value: token},
		{Value: req.ClientID},
		{Value: req.UserID},
		{Value: string(scopesJSON)},
		{Value: expiresAt},
		{Value: createdAt},
	}

	err = executePgQuery(ctx, instance.ConnectionID, query, queryParams)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	response.Token = token
	return prepareResponse(rnh, response)
}

func ValidateAccessToken(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ValidateAccessTokenRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	response := &ValidateAccessTokenResponse{Error: "", Valid: false}

	oauth2Ctx, err := GetOAuth2ServerContext(ctx.Context.GoContextParent)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	instance, found := oauth2Ctx.GetInstance(req.InstanceID)
	if !found {
		response.Error = "instance not found"
		return prepareResponse(rnh, response)
	}

	query := `
		SELECT client_id, user_id, scopes, expires_at
		FROM oauth2_access_tokens
		WHERE token = $1
	`
	queryParams := []vmpostgresql.SqlQueryParam{{Value: req.Token}}

	data, err := queryPgDatabase(ctx, instance.ConnectionID, query, queryParams)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(data, &results); err != nil || len(results) == 0 {
		response.Error = "token not found"
		return prepareResponse(rnh, response)
	}

	result := results[0]

	// Check expiration
	expiresAt := int64(result["expires_at"].(float64))
	if time.Now().Unix() > expiresAt {
		response.Error = "token expired"
		return prepareResponse(rnh, response)
	}

	response.Valid = true
	response.ClientID = result["client_id"].(string)
	response.UserID = result["user_id"].(string)
	response.ExpiresAt = expiresAt

	// Parse scopes
	if scopesStr, ok := result["scopes"].(string); ok {
		json.Unmarshal([]byte(scopesStr), &response.Scopes)
	}

	return prepareResponse(rnh, response)
}

func RefreshAccessToken(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req RefreshAccessTokenRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	response := &RefreshAccessTokenResponse{Error: ""}

	oauth2Ctx, err := GetOAuth2ServerContext(ctx.Context.GoContextParent)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	instance, found := oauth2Ctx.GetInstance(req.InstanceID)
	if !found {
		response.Error = "instance not found"
		return prepareResponse(rnh, response)
	}

	// Query refresh token
	query := `
		SELECT client_id, user_id, scopes, expires_at
		FROM oauth2_refresh_tokens
		WHERE token = $1
	`
	queryParams := []vmpostgresql.SqlQueryParam{{Value: req.RefreshToken}}

	data, err := queryPgDatabase(ctx, instance.ConnectionID, query, queryParams)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(data, &results); err != nil || len(results) == 0 {
		response.Error = "refresh token not found"
		return prepareResponse(rnh, response)
	}

	result := results[0]

	// Check expiration (if set)
	if expiresAtVal, ok := result["expires_at"]; ok && expiresAtVal != nil {
		expiresAt := int64(expiresAtVal.(float64))
		if time.Now().Unix() > expiresAt {
			response.Error = "refresh token expired"
			return prepareResponse(rnh, response)
		}
	}

	// Extract data
	clientID := result["client_id"].(string)
	userID := result["user_id"].(string)
	var scopes []string
	if scopesStr, ok := result["scopes"].(string); ok {
		json.Unmarshal([]byte(scopesStr), &scopes)
	}

	// Get refresh token expiry for rotation
	var oldRefreshExpiresAt interface{} = nil
	if expiresAtVal, ok := result["expires_at"]; ok && expiresAtVal != nil {
		oldRefreshExpiresAt = int64(expiresAtVal.(float64))
	}

	// Generate new access token
	newAccessToken, err := GenerateSecureToken(AccessTokenLength)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	// Generate new refresh token (rotation for security)
	newRefreshToken, err := GenerateSecureToken(RefreshTokenLength)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	accessExpiresAt := time.Now().Add(time.Duration(req.ExpiresInSeconds) * time.Second).Unix()
	createdAt := time.Now().Unix()
	scopesJSON, _ := json.Marshal(scopes)

	// CRITICAL: Delete old refresh token first (prevent replay attacks)
	deleteQuery := "DELETE FROM oauth2_refresh_tokens WHERE token = $1"
	deleteParams := []vmpostgresql.SqlQueryParam{{Value: req.RefreshToken}}
	err = executePgQuery(ctx, instance.ConnectionID, deleteQuery, deleteParams)
	if err != nil {
		response.Error = "failed to invalidate old refresh token: " + err.Error()
		return prepareResponse(rnh, response)
	}

	// Insert new access token
	insertAccessQuery := `
		INSERT INTO oauth2_access_tokens
		(token, client_id, user_id, scopes, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	insertAccessParams := []vmpostgresql.SqlQueryParam{
		{Value: newAccessToken},
		{Value: clientID},
		{Value: userID},
		{Value: string(scopesJSON)},
		{Value: accessExpiresAt},
		{Value: createdAt},
	}

	err = executePgQuery(ctx, instance.ConnectionID, insertAccessQuery, insertAccessParams)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	// Insert new refresh token (rotation)
	insertRefreshQuery := `
		INSERT INTO oauth2_refresh_tokens
		(token, client_id, user_id, scopes, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	insertRefreshParams := []vmpostgresql.SqlQueryParam{
		{Value: newRefreshToken},
		{Value: clientID},
		{Value: userID},
		{Value: string(scopesJSON)},
		{Value: oldRefreshExpiresAt},
		{Value: createdAt},
	}

	err = executePgQuery(ctx, instance.ConnectionID, insertRefreshQuery, insertRefreshParams)
	if err != nil {
		response.Error = "failed to create new refresh token: " + err.Error()
		return prepareResponse(rnh, response)
	}

	// Prepare response with both new tokens
	response.AccessToken = newAccessToken
	response.RefreshToken = newRefreshToken
	response.ClientID = clientID
	response.UserID = userID
	response.Scopes = scopes
	response.ExpiresAt = accessExpiresAt
	if oldRefreshExpiresAt != nil {
		response.RefreshExpiresAt = oldRefreshExpiresAt.(int64)
	}

	return prepareResponse(rnh, response)
}

func RevokeToken(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req RevokeTokenRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	response := &RevokeTokenResponse{Error: ""}

	oauth2Ctx, err := GetOAuth2ServerContext(ctx.Context.GoContextParent)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	instance, found := oauth2Ctx.GetInstance(req.InstanceID)
	if !found {
		response.Error = "instance not found"
		return prepareResponse(rnh, response)
	}

	queryParams := []vmpostgresql.SqlQueryParam{{Value: req.Token}}
	var totalDeleted int64 = 0

	// Try to delete from access tokens
	query1 := "DELETE FROM oauth2_access_tokens WHERE token = $1"
	rowsAffected1, err1 := executePgQueryWithRowCount(ctx, instance.ConnectionID, query1, queryParams)
	if err1 != nil {
		response.Error = "failed to revoke access token: " + err1.Error()
		return prepareResponse(rnh, response)
	}
	totalDeleted += rowsAffected1

	// Try to delete from refresh tokens
	query2 := "DELETE FROM oauth2_refresh_tokens WHERE token = $1"
	rowsAffected2, err2 := executePgQueryWithRowCount(ctx, instance.ConnectionID, query2, queryParams)
	if err2 != nil {
		response.Error = "failed to revoke refresh token: " + err2.Error()
		return prepareResponse(rnh, response)
	}
	totalDeleted += rowsAffected2

	// Verify that at least one token was actually revoked
	if totalDeleted == 0 {
		response.Error = "token not found"
		return prepareResponse(rnh, response)
	}

	return prepareResponse(rnh, response)
}

func IntrospectToken(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req IntrospectTokenRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	response := &IntrospectTokenResponse{Error: "", Active: false}

	oauth2Ctx, err := GetOAuth2ServerContext(ctx.Context.GoContextParent)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	instance, found := oauth2Ctx.GetInstance(req.InstanceID)
	if !found {
		response.Error = "instance not found"
		return prepareResponse(rnh, response)
	}

	// Try access token first
	query := `
		SELECT client_id, user_id, scopes, expires_at, created_at, 'access_token' as token_type
		FROM oauth2_access_tokens
		WHERE token = $1
	`
	queryParams := []vmpostgresql.SqlQueryParam{{Value: req.Token}}

	data, err := queryPgDatabase(ctx, instance.ConnectionID, query, queryParams)
	if err == nil {
		var results []map[string]interface{}
		if err := json.Unmarshal(data, &results); err == nil && len(results) > 0 {
			result := results[0]
			expiresAt := int64(result["expires_at"].(float64))

			// Check if expired
			if time.Now().Unix() <= expiresAt {
				response.Active = true
				response.ClientID = result["client_id"].(string)
				response.Username = result["user_id"].(string)
				response.TokenType = "access_token"
				response.ExpiresAt = expiresAt
				response.IssuedAt = int64(result["created_at"].(float64))

				var scopes []string
				if scopesStr, ok := result["scopes"].(string); ok {
					json.Unmarshal([]byte(scopesStr), &scopes)
				}
				response.Scope = strings.Join(scopes, " ")

				return prepareResponse(rnh, response)
			}
		}
	}

	// Try refresh token
	query = `
		SELECT client_id, user_id, scopes, expires_at, created_at, 'refresh_token' as token_type
		FROM oauth2_refresh_tokens
		WHERE token = $1
	`

	data, err = queryPgDatabase(ctx, instance.ConnectionID, query, queryParams)
	if err == nil {
		var results []map[string]interface{}
		if err := json.Unmarshal(data, &results); err == nil && len(results) > 0 {
			result := results[0]

			// Check expiration (if set)
			active := true
			if expiresAtVal, ok := result["expires_at"]; ok && expiresAtVal != nil {
				expiresAt := int64(expiresAtVal.(float64))
				if time.Now().Unix() > expiresAt {
					active = false
				}
				response.ExpiresAt = expiresAt
			}

			if active {
				response.Active = true
				response.ClientID = result["client_id"].(string)
				response.Username = result["user_id"].(string)
				response.TokenType = "refresh_token"
				response.IssuedAt = int64(result["created_at"].(float64))

				var scopes []string
				if scopesStr, ok := result["scopes"].(string); ok {
					json.Unmarshal([]byte(scopesStr), &scopes)
				}
				response.Scope = strings.Join(scopes, " ")

				return prepareResponse(rnh, response)
			}
		}
	}

	// Token not found or expired
	return prepareResponse(rnh, response)
}

func CleanupExpiredTokens(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req CleanupExpiredTokensRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	response := &CleanupExpiredTokensResponse{Error: ""}

	oauth2Ctx, err := GetOAuth2ServerContext(ctx.Context.GoContextParent)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	instance, found := oauth2Ctx.GetInstance(req.InstanceID)
	if !found {
		response.Error = "instance not found"
		return prepareResponse(rnh, response)
	}

	now := time.Now().Unix()

	// Delete expired authorization codes
	query1 := "DELETE FROM oauth2_authorization_codes WHERE expires_at < $1"
	queryParams1 := []vmpostgresql.SqlQueryParam{{Value: now}}
	deletedAuthCodes, err := executePgQueryWithRowCount(ctx, instance.ConnectionID, query1, queryParams1)
	if err != nil {
		response.Error = "failed to cleanup authorization codes: " + err.Error()
		return prepareResponse(rnh, response)
	}

	// Delete expired access tokens
	query2 := "DELETE FROM oauth2_access_tokens WHERE expires_at < $1"
	queryParams2 := []vmpostgresql.SqlQueryParam{{Value: now}}
	deletedAccessTokens, err := executePgQueryWithRowCount(ctx, instance.ConnectionID, query2, queryParams2)
	if err != nil {
		response.Error = "failed to cleanup access tokens: " + err.Error()
		return prepareResponse(rnh, response)
	}

	// Delete expired refresh tokens (only those with expiration set)
	query3 := "DELETE FROM oauth2_refresh_tokens WHERE expires_at IS NOT NULL AND expires_at < $1"
	queryParams3 := []vmpostgresql.SqlQueryParam{{Value: now}}
	deletedRefreshTokens, err := executePgQueryWithRowCount(ctx, instance.ConnectionID, query3, queryParams3)
	if err != nil {
		response.Error = "failed to cleanup refresh tokens: " + err.Error()
		return prepareResponse(rnh, response)
	}

	// Return actual counts
	response.DeletedAuthCodes = deletedAuthCodes
	response.DeletedAccessTokens = deletedAccessTokens
	response.DeletedRefreshTokens = deletedRefreshTokens

	return prepareResponse(rnh, response)
}
