package vmoauth2server

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/loredanacirstea/wasmx/wasmx-vmpostgresql"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func Initialize(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req InitializeRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	response := &InitializeResponse{Error: ""}

	oauth2Ctx, err := GetOAuth2ServerContext(ctx.Context.GoContextParent)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	// Store the instance
	instance := &OAuth2ServerInstance{
		ConnectionID: req.ConnectionID,
		ChainID:      ctx.Ctx.ChainID(),
	}

	err = oauth2Ctx.SetInstance(req.InstanceID, instance)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	// Create tables if they don't exist
	err = createTables(ctx, req.ConnectionID)
	if err != nil {
		response.Error = fmt.Sprintf("failed to create tables: %s", err.Error())
		return prepareResponse(rnh, response)
	}

	return prepareResponse(rnh, response)
}

func RegisterClient(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req RegisterClientRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	response := &RegisterClientResponse{Error: ""}

	// Get instance
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

	// Hash the client secret
	hashedSecret, err := HashSecret(req.ClientSecret)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	// Insert into database
	query := `
		INSERT INTO oauth2_clients
		(client_id, client_secret, redirect_uris, scopes, grant_types, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	redirectURIsJSON, _ := json.Marshal(req.RedirectURIs)
	scopesJSON, _ := json.Marshal(req.Scopes)
	grantTypesJSON, _ := json.Marshal(req.GrantTypes)
	now := time.Now().Unix()

	queryParams := []vmpostgresql.SqlQueryParam{
		{Value: req.ClientID},
		{Value: hashedSecret},
		{Value: string(redirectURIsJSON)},
		{Value: string(scopesJSON)},
		{Value: string(grantTypesJSON)},
		{Value: now},
		{Value: now},
	}

	err = executePgQuery(ctx, instance.ConnectionID, query, queryParams)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	return prepareResponse(rnh, response)
}

func UpdateClient(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req UpdateClientRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	response := &UpdateClientResponse{Error: ""}

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

	// Build dynamic update query
	updates := []string{}
	queryParams := []vmpostgresql.SqlQueryParam{}
	paramIdx := 1

	if len(req.RedirectURIs) > 0 {
		redirectURIsJSON, _ := json.Marshal(req.RedirectURIs)
		updates = append(updates, fmt.Sprintf("redirect_uris = $%d", paramIdx))
		queryParams = append(queryParams, vmpostgresql.SqlQueryParam{Value: string(redirectURIsJSON)})
		paramIdx++
	}

	if len(req.Scopes) > 0 {
		scopesJSON, _ := json.Marshal(req.Scopes)
		updates = append(updates, fmt.Sprintf("scopes = $%d", paramIdx))
		queryParams = append(queryParams, vmpostgresql.SqlQueryParam{Value: string(scopesJSON)})
		paramIdx++
	}

	if len(req.GrantTypes) > 0 {
		grantTypesJSON, _ := json.Marshal(req.GrantTypes)
		updates = append(updates, fmt.Sprintf("grant_types = $%d", paramIdx))
		queryParams = append(queryParams, vmpostgresql.SqlQueryParam{Value: string(grantTypesJSON)})
		paramIdx++
	}

	if len(updates) == 0 {
		response.Error = "no fields to update"
		return prepareResponse(rnh, response)
	}

	now := time.Now().Unix()
	updates = append(updates, fmt.Sprintf("updated_at = $%d", paramIdx))
	queryParams = append(queryParams, vmpostgresql.SqlQueryParam{Value: now})
	paramIdx++

	query := fmt.Sprintf("UPDATE oauth2_clients SET %s WHERE client_id = $%d", strings.Join(updates, ", "), paramIdx)
	queryParams = append(queryParams, vmpostgresql.SqlQueryParam{Value: req.ClientID})

	err = executePgQuery(ctx, instance.ConnectionID, query, queryParams)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	return prepareResponse(rnh, response)
}

func DeleteClient(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req DeleteClientRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	response := &DeleteClientResponse{Error: ""}

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

	query := "DELETE FROM oauth2_clients WHERE client_id = $1"
	queryParams := []vmpostgresql.SqlQueryParam{{Value: req.ClientID}}

	err = executePgQuery(ctx, instance.ConnectionID, query, queryParams)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	return prepareResponse(rnh, response)
}

func GetClient(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req GetClientRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	response := &GetClientResponse{Error: ""}

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
		SELECT client_id, redirect_uris, scopes, grant_types, created_at, updated_at
		FROM oauth2_clients
		WHERE client_id = $1
	`
	queryParams := []vmpostgresql.SqlQueryParam{{Value: req.ClientID}}

	data, err := queryPgDatabase(ctx, instance.ConnectionID, query, queryParams)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(data, &results); err != nil {
		response.Error = "failed to parse query results"
		return prepareResponse(rnh, response)
	}

	if len(results) == 0 {
		response.Error = "client not found"
		return prepareResponse(rnh, response)
	}

	result := results[0]
	response.ClientID = result["client_id"].(string)

	// Parse JSON arrays
	if redirectURIs, ok := result["redirect_uris"].(string); ok {
		json.Unmarshal([]byte(redirectURIs), &response.RedirectURIs)
	}
	if scopes, ok := result["scopes"].(string); ok {
		json.Unmarshal([]byte(scopes), &response.Scopes)
	}
	if grantTypes, ok := result["grant_types"].(string); ok {
		json.Unmarshal([]byte(grantTypes), &response.GrantTypes)
	}

	if createdAt, ok := result["created_at"].(float64); ok {
		response.CreatedAt = int64(createdAt)
	}
	if updatedAt, ok := result["updated_at"].(float64); ok {
		response.UpdatedAt = int64(updatedAt)
	}

	return prepareResponse(rnh, response)
}

func ValidateClient(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ValidateClientRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	response := &ValidateClientResponse{Error: "", Valid: false}

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

	query := "SELECT client_secret FROM oauth2_clients WHERE client_id = $1"
	queryParams := []vmpostgresql.SqlQueryParam{{Value: req.ClientID}}

	data, err := queryPgDatabase(ctx, instance.ConnectionID, query, queryParams)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(data, &results); err != nil || len(results) == 0 {
		response.Error = "client not found"
		return prepareResponse(rnh, response)
	}

	hashedSecret := results[0]["client_secret"].(string)
	response.Valid = ValidateSecret(req.ClientSecret, hashedSecret)

	return prepareResponse(rnh, response)
}

func CreateAuthorizationCode(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req CreateAuthorizationCodeRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	response := &CreateAuthorizationCodeResponse{Error: ""}

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

	// Generate secure authorization code
	code, err := GenerateSecureToken(AuthorizationCodeLength)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	expiresAt := time.Now().Add(time.Duration(req.ExpiresInSeconds) * time.Second).Unix()
	createdAt := time.Now().Unix()

	scopesJSON, _ := json.Marshal(req.Scopes)

	query := `
		INSERT INTO oauth2_authorization_codes
		(code, client_id, user_id, redirect_uri, scopes, code_challenge, code_challenge_method, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	queryParams := []vmpostgresql.SqlQueryParam{
		{Value: code},
		{Value: req.ClientID},
		{Value: req.UserID},
		{Value: req.RedirectURI},
		{Value: string(scopesJSON)},
		{Value: req.CodeChallenge},
		{Value: req.CodeChallengeMethod},
		{Value: expiresAt},
		{Value: createdAt},
	}

	err = executePgQuery(ctx, instance.ConnectionID, query, queryParams)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	response.Code = code
	return prepareResponse(rnh, response)
}

func ValidateAuthorizationCode(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ValidateAuthorizationCodeRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	response := &ValidateAuthorizationCodeResponse{Error: "", Valid: false}

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

	// Query and delete the authorization code (single use)
	query := `
		SELECT user_id, client_id, redirect_uri, scopes, code_challenge, code_challenge_method, expires_at
		FROM oauth2_authorization_codes
		WHERE code = $1
	`
	queryParams := []vmpostgresql.SqlQueryParam{{Value: req.Code}}

	data, err := queryPgDatabase(ctx, instance.ConnectionID, query, queryParams)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(data, &results); err != nil || len(results) == 0 {
		response.Error = "authorization code not found"
		return prepareResponse(rnh, response)
	}

	result := results[0]

	// Validate client ID
	if result["client_id"].(string) != req.ClientID {
		response.Error = "client ID mismatch"
		return prepareResponse(rnh, response)
	}

	// Validate redirect URI
	if result["redirect_uri"].(string) != req.RedirectURI {
		response.Error = "redirect URI mismatch"
		return prepareResponse(rnh, response)
	}

	// Check expiration
	expiresAt := int64(result["expires_at"].(float64))
	if time.Now().Unix() > expiresAt {
		// Delete expired code
		deleteQuery := "DELETE FROM oauth2_authorization_codes WHERE code = $1"
		executePgQuery(ctx, instance.ConnectionID, deleteQuery, queryParams)
		response.Error = "authorization code expired"
		return prepareResponse(rnh, response)
	}

	// Validate PKCE if provided
	codeChallenge := ""
	codeChallengeMethod := ""
	if cc, ok := result["code_challenge"].(string); ok {
		codeChallenge = cc
	}
	if ccm, ok := result["code_challenge_method"].(string); ok {
		codeChallengeMethod = ccm
	}

	if !ValidatePKCE(req.CodeVerifier, codeChallenge, codeChallengeMethod) {
		response.Error = "PKCE validation failed"
		return prepareResponse(rnh, response)
	}

	// Delete the code (single use)
	deleteQuery := "DELETE FROM oauth2_authorization_codes WHERE code = $1"
	err = executePgQuery(ctx, instance.ConnectionID, deleteQuery, queryParams)
	if err != nil {
		response.Error = fmt.Sprintf("failed to delete authorization code: %s", err.Error())
		return prepareResponse(rnh, response)
	}

	response.Valid = true
	response.UserID = result["user_id"].(string)

	// Parse scopes
	if scopesStr, ok := result["scopes"].(string); ok {
		json.Unmarshal([]byte(scopesStr), &response.Scopes)
	}

	return prepareResponse(rnh, response)
}

// IssueAccessToken, IssueRefreshToken, ValidateAccessToken, RefreshAccessToken, RevokeToken, IntrospectToken, CleanupExpiredTokens
// will be in the next file continuation...

func prepareResponse(rnh memc.RuntimeHandler, response interface{}) ([]interface{}, error) {
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(responsebz)
}

// Helper functions for PostgreSQL operations
func executePgQuery(ctx *Context, connectionID string, query string, queryParams []vmpostgresql.SqlQueryParam) error {
	sqlCtx, err := vmpostgresql.GetSqlContext(ctx.Context.GoContextParent)
	if err != nil {
		return err
	}

	connId := buildConnectionId(ctx.Ctx.ChainID(), connectionID, ctx.Env.Contract.Address.String())
	conn, found := sqlCtx.GetConnection(connId)
	if !found {
		return fmt.Errorf("postgresql connection not found: %s", connId)
	}

	// Convert query params to the format expected by PostgreSQL
	params := make([]interface{}, len(queryParams))
	for i, p := range queryParams {
		params[i] = p.Value
	}

	// Use the existing transaction if available
	if conn.OpenSavepointTx == nil {
		batch, err := conn.Db.NewBatchWithError()
		if err != nil {
			return fmt.Errorf("cannot begin db transaction: %s", err.Error())
		}
		conn.OpenSavepointTx = batch
		conn.setSavePoint("sp0")
		_, err = batch.Tx().Exec(ctx.Ctx, "SAVEPOINT sp0")
		if err != nil {
			return fmt.Errorf("cannot add savepoint: %v", err)
		}
	}

	_, err = conn.OpenSavepointTx.Tx().Exec(ctx.Ctx, query, params...)
	return err
}

func queryPgDatabase(ctx *Context, connectionID string, query string, queryParams []vmpostgresql.SqlQueryParam) ([]byte, error) {
	sqlCtx, err := vmpostgresql.GetSqlContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	connId := buildConnectionId(ctx.Ctx.ChainID(), connectionID, ctx.Env.Contract.Address.String())
	conn, found := sqlCtx.GetConnection(connId)
	if !found {
		return nil, fmt.Errorf("postgresql connection not found: %s", connId)
	}

	// Convert query params
	params := make([]interface{}, len(queryParams))
	for i, p := range queryParams {
		params[i] = p.Value
	}

	// Use the existing transaction if available
	if conn.OpenSavepointTx == nil {
		batch, err := conn.Db.NewBatchWithError()
		if err != nil {
			return nil, fmt.Errorf("cannot begin db transaction: %s", err.Error())
		}
		conn.OpenSavepointTx = batch
		conn.setSavePoint("sp0")
		_, err = batch.Tx().Exec(ctx.Ctx, "SAVEPOINT sp0")
		if err != nil {
			return nil, fmt.Errorf("cannot add savepoint: %v", err)
		}
	}

	rows, err := conn.OpenSavepointTx.Tx().Query(ctx.Ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return vmpostgresql.RowsToJSON(rows)
}

func buildConnectionId(chainId, contractAddress, connectionID string) string {
	return fmt.Sprintf("%s_%s_%s", chainId, contractAddress, connectionID)
}

func createTables(ctx *Context, connectionID string) error {
	schema := `
	CREATE TABLE IF NOT EXISTS oauth2_clients (
		client_id VARCHAR(255) PRIMARY KEY,
		client_secret VARCHAR(255) NOT NULL,
		redirect_uris JSONB NOT NULL,
		scopes JSONB NOT NULL,
		grant_types JSONB NOT NULL,
		created_at BIGINT NOT NULL,
		updated_at BIGINT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS oauth2_authorization_codes (
		code VARCHAR(255) PRIMARY KEY,
		client_id VARCHAR(255) NOT NULL,
		user_id VARCHAR(255) NOT NULL,
		redirect_uri TEXT,
		scopes JSONB NOT NULL,
		code_challenge VARCHAR(255),
		code_challenge_method VARCHAR(10),
		expires_at BIGINT NOT NULL,
		created_at BIGINT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_oauth2_auth_codes_expires_at ON oauth2_authorization_codes(expires_at);

	CREATE TABLE IF NOT EXISTS oauth2_access_tokens (
		token VARCHAR(255) PRIMARY KEY,
		client_id VARCHAR(255) NOT NULL,
		user_id VARCHAR(255),
		scopes JSONB NOT NULL,
		expires_at BIGINT NOT NULL,
		created_at BIGINT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_oauth2_access_tokens_expires_at ON oauth2_access_tokens(expires_at);
	CREATE INDEX IF NOT EXISTS idx_oauth2_access_tokens_user_id ON oauth2_access_tokens(user_id);

	CREATE TABLE IF NOT EXISTS oauth2_refresh_tokens (
		token VARCHAR(255) PRIMARY KEY,
		client_id VARCHAR(255) NOT NULL,
		user_id VARCHAR(255) NOT NULL,
		scopes JSONB NOT NULL,
		access_token VARCHAR(255),
		expires_at BIGINT,
		created_at BIGINT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_oauth2_refresh_tokens_user_id ON oauth2_refresh_tokens(user_id);
	`

	statements := strings.Split(schema, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		err := executePgQuery(ctx, connectionID, stmt, nil)
		if err != nil {
			return fmt.Errorf("failed to execute schema statement: %s, error: %s", stmt, err.Error())
		}
	}

	return nil
}
