package lib

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	auth "github.com/loredanacirstea/wasmx-auth/lib"
	wasmxhttp "github.com/loredanacirstea/wasmx-env-httpserver/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// HandleHTTPRequest handles incoming HTTP requests
func HandleHTTPRequest(req *wasmxhttp.HttpRequestIncoming) []byte {
	routePrefix := LoadRoutePrefix()

	switch req.RequestURI {
	case routePrefix + "/register_init":
		return handleRegisterInit(req)
	case routePrefix + "/account_info":
		return handleAccountInfo(req)
	case routePrefix + "/contract_addresses":
		return handleContractAddresses(req)
	default:
		return respondWithError("route not found", 404)
	}
}

// handleRegisterInit handles the /register_init endpoint
// This endpoint receives a user's address and initializes it by sending native coins
func handleRegisterInit(req *wasmxhttp.HttpRequestIncoming) []byte {
	if req.Method != "POST" {
		return respondWithError("method not allowed", 405)
	}

	// Parse request body
	type RegisterInitRequest struct {
		Address string `json:"address"`
	}

	var reqData RegisterInitRequest
	if err := json.Unmarshal(req.Data, &reqData); err != nil {
		LoggerError("Failed to parse register_init request", []string{"error", err.Error()})
		return respondWithError("invalid request body", 400)
	}

	if reqData.Address == "" {
		return respondWithError("address is required", 400)
	}

	// Call InitAccount
	msg := &MsgInitAccount{
		Address: reqData.Address,
	}

	result := InitAccount(msg)

	// Parse result
	var response MsgInitAccountResponse
	if err := json.Unmarshal(result, &response); err != nil {
		LoggerError("Failed to parse InitAccount response", []string{"error", err.Error()})
		return respondWithError("internal error", 500)
	}

	if !response.Success {
		return respondWithError(response.Error, 500)
	}

	// Return success response
	return respondWithJSON(map[string]interface{}{
		"success": true,
		"tx_hash": response.TxHash,
		"address": reqData.Address,
	}, 200)
}

// handleAccountInfo handles the /account_info endpoint
// This endpoint returns account information (account number and sequence) needed for signing transactions
func handleAccountInfo(req *wasmxhttp.HttpRequestIncoming) []byte {
	if req.Method != "GET" && req.Method != "POST" {
		return respondWithError("method not allowed", 405)
	}

	// Get address from query params or body
	address := ""
	if req.Method == "GET" {
		// Parse URL for query parameters
		// The RequestURI might contain query string like "/auth/account_info?address=..."
		for idx := 0; idx < len(req.RequestURI); idx++ {
			if req.RequestURI[idx] == '?' {
				queryString := req.RequestURI[idx+1:]
				// Simple parsing of address parameter
				if len(queryString) > 8 && queryString[:8] == "address=" {
					address = queryString[8:]
				}
				break
			}
		}
	} else {
		type AccountInfoRequest struct {
			Address string `json:"address"`
		}
		var reqData AccountInfoRequest
		if err := json.Unmarshal(req.Data, &reqData); err == nil {
			address = reqData.Address
		}
	}

	if address == "" {
		return respondWithError("address is required", 400)
	}

	// Query account information from wasmx-auth contract
	authAddr := wasmx.GetAddressByRole(wasmx.ROLE_AUTH)
	if authAddr == "" {
		LoggerError("Auth contract not found", nil)
		return respondWithError("auth contract not found", 500)
	}

	// Build query message
	queryMsg := map[string]interface{}{
		"GetAccount": map[string]interface{}{
			"address": address,
		},
	}
	queryMsgBytes, err := json.Marshal(queryMsg)
	if err != nil {
		LoggerError("Failed to marshal query message", []string{"error", err.Error()})
		return respondWithError("failed to marshal query", 500)
	}

	// Query auth contract
	ok, data := wasmx.CallSimple(authAddr, queryMsgBytes, true, MODULE_NAME)
	if !ok {
		LoggerError("Failed to query account info", []string{"address", address, "response", string(data)})
		return respondWithError("failed to query account info: "+string(data), 500)
	}

	var response auth.QueryAccountResponse
	if err := json.Unmarshal(data, &response); err != nil {
		LoggerError("Failed to parse account info response", []string{"error", err.Error(), "data", string(data)})
		return respondWithError("failed to parse account info", 500)
	}

	if response.Account == nil {
		return respondWithError("acount does not exist", 500)
	}
	accbz, err := base64.StdEncoding.DecodeString(response.Account.Value)
	if err != nil {
		return respondWithError("failed to decode Any value", 500)
	}
	return respondWithBytes(accbz, 200)
}

// handleContractAddresses handles the /contract_addresses endpoint
// Returns addresses of important contracts like wasmx-identity
func handleContractAddresses(req *wasmxhttp.HttpRequestIncoming) []byte {
	if req.Method != "GET" {
		return respondWithError("method not allowed", 405)
	}

	// Get contract addresses by role
	identityAddr := wasmx.GetAddressByRole(wasmx.ROLE_ACCOUNT_IDENTITY)
	authAddr := wasmx.GetAddressByRole(wasmx.ROLE_AUTH)

	// Query chain info from auth contract
	chainId := wasmx.GetChainId()
	// TODO get denom from chain
	baseDenom := "amyt"

	return respondWithJSON(map[string]interface{}{
		"chain_id":         chainId,
		"base_denom":       baseDenom,
		"identity_address": string(identityAddr),
		"auth_address":     string(authAddr),
	}, 200)
}

// Helper functions

func respondWithJSON(data interface{}, statusCode int) []byte {
	jsonData, _ := json.Marshal(data)
	return respondWithBytes(jsonData, statusCode)
}

func respondWithBytes(jsonData []byte, statusCode int) []byte {
	resp := wasmxhttp.HttpResponseWrap{
		Error: "",
		Data: wasmxhttp.HttpResponse{
			Status:     fmt.Sprintf("%d", statusCode),
			StatusCode: statusCode,
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Data: jsonData,
		},
	}

	respBytes, _ := json.Marshal(resp)
	return respBytes
}

func respondWithError(message string, statusCode int) []byte {
	return respondWithJSON(map[string]string{
		"error": message,
	}, statusCode)
}
