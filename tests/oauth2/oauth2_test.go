package keeper_test

import (
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	testdata "github.com/loredanacirstea/mythos-tests/testdata/tinygo"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
)

func (suite *KeeperTestSuite) TestOauth2() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	// Load database configuration from environment (for Supabase testing)
	dbConnection := os.Getenv("TEST_MCP_DB_CONNECTION")
	if dbConnection == "" {
		dbConnection = "postgresql://localhost:5432/postgres"
	}

	dbName := os.Getenv("TEST_MCP_DB_NAME")
	if dbName == "" {
		dbName = "mcp_search_test"
	}

	fmt.Println("=== Using Database Configuration ===")
	fmt.Println("Connection:", dbConnection)
	fmt.Println("Database:", dbName)
	fmt.Println("====================================")

	// Prepare init data for mcp-userdata contract
	userdataInitData := &MCPContractInitGenesis{
		InitGenesis: &MCPContractInitGenesisRequest{
			RoutePrefix: "/tools/userdata",
		},
	}
	userdataInitDataJSON, _ := json.Marshal(userdataInitData)

	// Instantiate mcp-userdata with init data
	userCodeId := appA.StoreCode(sender, testdata.MCPUserdata, nil)
	userAddress := appA.InstantiateCode(sender, userCodeId, types.WasmxExecutionMessage{Data: userdataInitDataJSON}, "mcp_userdata", nil)
	fmt.Println("Instantiated mcp-userdata contract:", userAddress.String())

	// Build search configuration
	searchConfig := buildSearchConfig(dbConnection, dbName)

	// Build tool descriptions for AI agents
	toolDescriptions := buildToolDescriptions()

	// Get scripts folder from environment or use default
	scriptsFolder := os.Getenv("SCRIPTS_FOLDER")
	if scriptsFolder == "" {
		scriptsFolder = "./tests/testdata/tinygo/wasmx-mcp-search/scripts"
	}

	// Prepare init data for mcp-search contract
	searchInitData := &MCPContractInitGenesis{
		InitGenesis: &MCPSearchInitGenesisRequest{
			RoutePrefix:      "/tools/search",
			SearchConfig:     searchConfig,
			ToolDescriptions: toolDescriptions,
			EnvironmentVars: map[string]string{
				"SCRIPTS_FOLDER": scriptsFolder,
			},
		},
	}
	searchInitDataJSON, _ := json.Marshal(searchInitData)

	// Instantiate mcp-search with init data
	searchCodeId := appA.StoreCode(sender, testdata.MCPSearch, nil)
	searchAddress := appA.InstantiateCode(sender, searchCodeId, types.WasmxExecutionMessage{Data: searchInitDataJSON}, "mcp_search", nil)
	fmt.Println("Instantiated mcp-search contract:", searchAddress.String())
	fmt.Println("Search configured with", len(searchConfig.Tables), "tables:", searchConfig.Tables[0].TableName, ",", searchConfig.Tables[1].TableName)

	fmt.Println("Assigning MCP role to search contract...")
	utils.RegisterRole(suite, appA, types.ROLE_MCP, searchAddress, sender)
	fmt.Println("MCP role assigned to search contract - auto-registration triggered via RoleChanged hook")

	fmt.Println("Assigning MCP role to userdata contract...")
	utils.RegisterRole(suite, appA, types.ROLE_MCP, userAddress, sender)
	fmt.Println("MCP role assigned to userdata contract - auto-registration triggered via RoleChanged hook")

	// Configure OAuth2 Keys contract with funder private key
	fmt.Println("Configuring OAuth2 Keys contract with funder private key...")
	oauth2KeysAddr := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_OAUTH2_KEYS))

	// Export the sender's private key to use as funder
	// In a real setup, this would be a dedicated funder account
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
	fmt.Println("OAuth2 Keys contract configured")

	// Register OAuth client
	oauth2Addr := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_OAUTH2_SERVER))
	registerClientMsg := &RegisterOAuthClientCalldata{
		RegisterOAuthClient: &RegisterOAuthClientRequest{
			Name:        "MCP Test Client",
			Description: "OAuth client for MCP testing",
			RedirectURIs: []string{
				"https://chat.openai.com/aip/callback",
				"https://chat.openai.com/aip/*",
				"https://chatgpt.com/connector_platform_oauth_redirect",
				"http://localhost:3000/callback",
			},
			Scopes: []string{"read", "tools"},
		},
	}
	registerClientData, err := json.Marshal(registerClientMsg)
	suite.Require().NoError(err)
	clientResp := appA.ExecuteContract(sender, oauth2Addr, types.WasmxExecutionMessage{Data: registerClientData}, nil, nil)

	var clientResult RegisterOAuthClientResponse
	err = appA.DecodeExecuteResponse(clientResp, &clientResult)
	if err != nil {
		suite.Require().NoError(err, "Error unmarshaling client response")
	}

	fmt.Println("=== OAuth Client Credentials ===")
	fmt.Println("Client ID:    ", clientResult.ClientID)
	fmt.Println("Client Secret:", clientResult.ClientSecret)
	fmt.Println("================================")

	// Configure HTTP server registry contract
	fmt.Println("Configuring HTTP server registry contract...")
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
	fmt.Println("HTTP server registry contract configured")

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

	fmt.Println("=== Server Information ===")
	fmt.Println("Server running on: http://localhost:8080")
	fmt.Println("Login URL:         http://localhost:8080/login")
	fmt.Println("OAuth Authorize:   http://localhost:8080/oauth/authorize")
	fmt.Println("OAuth Token:       http://localhost:8080/oauth/token")
	fmt.Println("MCP SSE Endpoint:  http://localhost:8080/sse")
	fmt.Println("OpenAPI schema:  http://localhost:8080/openapi.json")
	fmt.Println("OpenAPI schema:  http://localhost:8080/.well-known/ai-plugin.json")

	fmt.Println("==========================")

	suite.T().Log("MCP registry server running on :8080 with registered tool contracts... Press Ctrl+C to exit")

	// Create a channel to listen for interrupt/terminate signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-sig

	suite.T().Log("Received exit signal. Test ending.")
}

func (suite *KeeperTestSuite) TestIdentityRegistrationSignature() {
	// Test variables - customize these as needed
	const (
		// Use the private key from browser logs for comparison
		testPrivateKeyHex = "a88dd3f8ea9c59a54a2733646af24c7ac715fb0970c864d6f8e4e7d394534424"
		chainID           = "mythos_7000-14"
		baseDenom         = "amyt"
		gasLimit          = uint64(30000000)
		gasPrice          = "100" // 100amyt per gas
	)

	appA := s.AppContext()

	// Get identity contract address
	identityAddr := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_ACCOUNT_IDENTITY))
	fmt.Println("Identity contract address:", identityAddr.String())

	// Parse private key from hex
	privKeyBytes, err := hex.DecodeString(testPrivateKeyHex)
	suite.Require().NoError(err, "Failed to decode private key hex")
	suite.Require().Equal(32, len(privKeyBytes), "Private key must be 32 bytes")

	// Create secp256k1 private key
	privKey := &secp256k1.PrivKey{Key: privKeyBytes}
	pubKey := privKey.PubKey()
	pubKeyBytes := pubKey.Bytes()

	// Derive address
	testAddr := sdk.AccAddress(pubKey.Address())
	testAddrPrefixed := appA.BytesToAccAddressPrefixed(testAddr)

	fmt.Println("=== TEST KEY INFO ===")
	fmt.Println("Private Key (hex):", testPrivateKeyHex)
	fmt.Println("Public Key (hex):", hex.EncodeToString(pubKeyBytes))
	fmt.Println("Public Key (base64):", base64.StdEncoding.EncodeToString(pubKeyBytes))
	fmt.Println("Address:", testAddrPrefixed.String())
	fmt.Println("====================")

	// Fund the test account
	fundAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(100000000000000000).MulRaw(10000))
	appA.Faucet.Fund(appA.Context(), testAddrPrefixed, fundAmount)
	fmt.Println("Funded test account with:", fundAmount.String())

	// Get account info
	acc, err := appA.App.AccountKeeper.GetAccountPrefixed(appA.Context(), testAddrPrefixed)
	suite.Require().NoError(err)
	suite.Require().NotNil(acc)
	accno := acc.GetAccountNumber()
	accseq := acc.GetSequence()
	fmt.Println("Account Number:", accno)
	fmt.Println("Sequence:", accseq)

	// accno = uint64(71)
	// accseq = uint64(0)

	// Create register_user message - exact same as in /register page
	registerUserMsg := map[string]interface{}{
		"register_user": map[string]interface{}{
			"address":        testAddrPrefixed.String(),
			"public_key":     base64.StdEncoding.EncodeToString(pubKeyBytes),
			"service_domain": "",
			"permissions":    []string{},
			"expires_at":     0,
		},
	}
	innerMsgBytes, err := json.Marshal(registerUserMsg)
	suite.Require().NoError(err)

	// Wrap in WasmxExecutionMessage format: {"data": "<base64>"}
	wrappedMsg := map[string]string{
		"data": base64.StdEncoding.EncodeToString(innerMsgBytes),
	}
	msgBytes, err := json.Marshal(wrappedMsg)
	suite.Require().NoError(err)

	fmt.Println("=== MESSAGE ===")
	fmt.Println("Inner message:", string(innerMsgBytes))
	fmt.Println("Wrapped message:", string(msgBytes))
	fmt.Println("===============")

	// Create MsgExecuteContract
	executeMsg := &types.MsgExecuteContract{
		Sender:       testAddrPrefixed.String(),
		Contract:     identityAddr.String(),
		Msg:          msgBytes,
		Funds:        sdk.NewCoins(),
		Dependencies: []string{},
	}

	// Get TxConfig and create tx builder
	txConfig := appA.App.TxConfig()
	txBuilder := txConfig.NewTxBuilder()

	// Set gas limit
	txBuilder.SetGasLimit(gasLimit)

	// Calculate and set fee
	gasPriceInt, ok := sdkmath.NewIntFromString(gasPrice)
	suite.Require().True(ok, "Failed to parse gas price")
	feeAmount := gasPriceInt.MulRaw(int64(gasLimit))
	fees := sdk.NewCoins(sdk.NewCoin(baseDenom, feeAmount))
	txBuilder.SetFeeAmount(fees)

	// Set message
	err = txBuilder.SetMsgs(executeMsg)
	suite.Require().NoError(err)

	// First round: set empty signature
	sigV2 := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode(txConfig.SignModeHandler().DefaultMode()),
			Signature: nil,
		},
		Sequence: accseq,
	}

	err = txBuilder.SetSignatures(sigV2)
	suite.Require().NoError(err)

	// Second round: sign transaction
	signerData := authsigning.SignerData{
		ChainID:       appA.Context().ChainID(),
		AccountNumber: accno,
		Sequence:      accseq,
		PubKey:        pubKey,
		Address:       acc.String(),
	}

	fmt.Println("=== SIGNER DATA ===")
	fmt.Println("Chain ID:", signerData.ChainID)
	fmt.Println("Account Number:", signerData.AccountNumber)
	fmt.Println("Sequence:", signerData.Sequence)
	fmt.Println("Address:", signerData.Address)
	fmt.Println("==================")

	sigV2, err = tx.SignWithPrivKey(
		appA.Context().Context(),
		signing.SignMode(txConfig.SignModeHandler().DefaultMode()),
		signerData,
		txBuilder,
		privKey,
		txConfig,
		accseq,
	)
	suite.Require().NoError(err)

	err = txBuilder.SetSignatures(sigV2)
	suite.Require().NoError(err)

	// Encode transaction
	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	suite.Require().NoError(err)

	fmt.Println("=== TRANSACTION BYTES ===")
	fmt.Println("Hex:", hex.EncodeToString(txBytes))
	fmt.Println("Base64:", base64.StdEncoding.EncodeToString(txBytes))
	fmt.Println("Length:", len(txBytes), "bytes")
	fmt.Println("========================")

	req := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	checkResp, err := appA.App.CheckTx(req)
	suite.Require().NoError(err)
	suite.Require().Equal(checkResp.Code, abci.CodeTypeOK)

	// Broadcast transaction
	// res, err := appA.DeliverTxRaw(txBytes)
	// suite.Require().NoError(err)
	// suite.Require().True(res.IsOK(), "Transaction failed: "+res.GetLog())

	// fmt.Println("=== TRANSACTION RESULT ===")
	// fmt.Println("Success:", res.IsOK())
	// fmt.Println("Log:", res.GetLog())
	// fmt.Println("Gas Used:", res.GasUsed)
	// fmt.Println("=========================")

	// Broadcast transaction2
	txBytes2, err := hex.DecodeString("0a93030a90030a232f6d7974686f732e7761736d782e76312e4d736745786563757465436f6e747261637412e8020a2d6d7974686f7331327077713279736c75776d306c3438663763716a363533306173787a7039657661346a383037122d6d7974686f7331717171717171717171717171717171717171717171717171717171717171726637716a7a79671a87027b2264617461223a2265794a795a5764706333526c636c393163325679496a7037496d466b5a484a6c63334d694f694a746558526f62334d784d6e423363544a356332783164323077624451345a6a646a63576f324e544d7759584e34656e41355a585a684e476f344d4463694c434a7764574a7361574e6661325635496a6f695153383254484534636a41316431646b644456425933525461544172655768555a30347654555645557a683064334e4c57573169516b5a73534545694c434a7a5a584a3261574e6c58325276625746706269493649694973496e426c636d317063334e706232357a496a7062585377695a58687761584a6c63313968644349364d483139227d12740a500a460a1f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657912230a2103fe8babcaf4e7059db7901cb528b4fb285380dfcc1034bcb70b0a6266c11651c012040a020801180112200a190a04616d797412113130303030303030303030303030303030108087a70e1a403ac984da1bd14bf9b447da5c9fcaf37d6929366df72af109d91b3be6130d51dc4a5cb1ee205d3f3c0a2cc56451e9a98e0ed3f1e3650c0e491db8cef6b959cf98")

	req = &abci.RequestCheckTx{Tx: txBytes2, Type: abci.CheckTxType_New}
	checkResp, err = appA.App.CheckTx(req)
	suite.Require().NoError(err)
	suite.Require().Equal(abci.CodeTypeOK, checkResp.Code, checkResp.Info, checkResp.Log, string(checkResp.Data))

	// res, err = appA.DeliverTxRaw(txBytes2)
	// suite.Require().NoError(err)
	// suite.Require().True(res.IsOK(), "Transaction failed: "+res.GetLog())

	// fmt.Println("=== TRANSACTION RESULT ===")
	// fmt.Println("Success:", res.IsOK())
	// fmt.Println("Log:", res.GetLog())
	// fmt.Println("Gas Used:", res.GasUsed)
	// fmt.Println("=========================")
}
