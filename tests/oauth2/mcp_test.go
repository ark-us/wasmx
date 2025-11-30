package keeper_test

// import (
// 	_ "embed"
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"os/signal"
// 	"syscall"

// 	sdkmath "cosmossdk.io/math"
// 	sdk "github.com/cosmos/cosmos-sdk/types"

// 	"github.com/loredanacirstea/wasmx/x/wasmx/types"

// 	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
// )

// func (suite *KeeperTestSuite) TestMCPServer() {
// 	sender := suite.GetRandomAccount()
// 	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

// 	appA := s.AppContext()
// 	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

// 	contractAddress := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_MCP_SERVER))

// 	// Set a role to have access to protected APIs
// 	// utils.RegisterRole(suite, appA, "mcp", contractAddress, sender)

// 	// Initialize genesis with server parameters
// 	initGenesisMsg := &InitGenesisCalldata{
// 		InitGenesis: &InitGenesisRequest{
// 			Params: ServerParams{
// 				ClientID:     "test-mcp-client",
// 				ClientSecret: "test-secret-12345",
// 				RedirectURIs: []string{
// 					"https://chat.openai.com/aip/callback",
// 					"https://chatgpt.com/connector_platform_oauth_redirect",
// 					"http://localhost:3000/callback",
// 				},
// 				Scopes:       []string{"read", "write", "tools"},
// 				DbConnection: "host=localhost port=5432 user=postgres password=postgres sslmode=disable",
// 				DbName:       "mcp_test",
// 			},
// 		},
// 	}
// 	data, err := json.Marshal(initGenesisMsg)
// 	fmt.Println("InitGenesis:", string(data))
// 	suite.Require().NoError(err)
// 	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
// 	fmt.Println("Initialized MCP server genesis")

// 	// Initialize everything: connect to database, create tables, initialize OAuth2
// 	// Note: This would normally be called automatically by RoleChanged hook during bootstrap
// 	initTablesMsg := &InitTablesCalldata{
// 		InitTables: &InitTablesRequest{},
// 	}
// 	data, err = json.Marshal(initTablesMsg)
// 	fmt.Println("InitializeTables:", string(data))
// 	suite.Require().NoError(err)
// 	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
// 	fmt.Println("Initialized tables and OAuth2")

// 	// Start the MCP server
// 	startServerMsg := &StartServerCalldata{
// 		StartServer: &StartServerRequest{
// 			Address: ":8080",
// 		},
// 	}
// 	data, err = json.Marshal(startServerMsg)
// 	fmt.Println("StartServer:", string(data))
// 	suite.Require().NoError(err)
// 	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
// 	fmt.Println("Started MCP server on :8080")

// 	suite.T().Log("MCP server running on :8080... Press Ctrl+C to exit")

// 	// Create a channel to listen for interrupt/terminate signals
// 	sig := make(chan os.Signal, 1)
// 	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

// 	// Block until a signal is received
// 	<-sig

// 	suite.T().Log("Received exit signal. Test ending.")
// }

// Calldata structures for MCP server
// type InitGenesisCalldata struct {
// 	InitGenesis *InitGenesisRequest `json:"init_genesis,omitempty"`
// }

// type StartServerCalldata struct {
// 	StartServer *StartServerRequest `json:"start_server,omitempty"`
// }

// type InitGenesisRequest struct {
// 	Params ServerParams `json:"params"`
// }

// type StartServerRequest struct {
// 	Address string `json:"address"`
// }

// type InitTablesCalldata struct {
// 	InitTables *InitTablesRequest `json:"init_tables,omitempty"`
// }

// type InitTablesRequest struct{}

// type ServerParams struct {
// 	ClientID     string   `json:"client_id"`
// 	ClientSecret string   `json:"client_secret"`
// 	RedirectURIs []string `json:"redirect_uris"`
// 	Scopes       []string `json:"scopes"`
// 	DbConnection string   `json:"db_connection"`
// 	DbName       string   `json:"db_name"`
// }
