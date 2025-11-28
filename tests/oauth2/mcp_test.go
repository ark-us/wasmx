package keeper_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
)

func (suite *KeeperTestSuite) TestMCPServer() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	contractAddress := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_MCP_SERVER))

	// Set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "mcp", contractAddress, sender)

	// Initialize genesis with server parameters
	initGenesisMsg := &InitGenesisCalldata{
		InitGenesis: &InitGenesisRequest{
			Params: ServerParams{
				ClientID:     "test-mcp-client",
				ClientSecret: "test-secret-12345",
				RedirectURIs: []string{"http://localhost:3000/callback"},
				Scopes:       []string{"read", "write", "tools"},
			},
		},
	}
	data, err := json.Marshal(initGenesisMsg)
	fmt.Println("InitGenesis:", string(data))
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
	fmt.Println("Initialized MCP server genesis")

	// Connect to database
	connectDbMsg := &ConnectDatabaseCalldata{
		ConnectDatabase: &ConnectDatabaseRequest{
			Connection: "host=localhost port=5432 user=postgres password=postgres sslmode=disable",
			DbName:     "mcp_test",
		},
	}
	data, err = json.Marshal(connectDbMsg)
	fmt.Println("ConnectDatabase:", string(data))
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
	fmt.Println("Connected to database")

	// Start the MCP server
	startServerMsg := &StartServerCalldata{
		StartServer: &StartServerRequest{
			Address: ":8080",
		},
	}
	data, err = json.Marshal(startServerMsg)
	fmt.Println("StartServer:", string(data))
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
	fmt.Println("Started MCP server on :8080")

	suite.T().Log("MCP server running on :8080... Press Ctrl+C to exit")

	// Create a channel to listen for interrupt/terminate signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-sig

	suite.T().Log("Received exit signal. Test ending.")
}

// Calldata structures for MCP server
type InitGenesisCalldata struct {
	InitGenesis *InitGenesisRequest `json:"init_genesis,omitempty"`
}

type ConnectDatabaseCalldata struct {
	ConnectDatabase *ConnectDatabaseRequest `json:"connect_database,omitempty"`
}

type StartServerCalldata struct {
	StartServer *StartServerRequest `json:"start_server,omitempty"`
}

type InitGenesisRequest struct {
	Params ServerParams `json:"params"`
}

type ConnectDatabaseRequest struct {
	Connection string `json:"connection"`
	DbName     string `json:"db_name"`
}

type StartServerRequest struct {
	Address string `json:"address"`
}

type ServerParams struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
}
