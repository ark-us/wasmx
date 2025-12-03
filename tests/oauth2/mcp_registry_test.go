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

	testdata "github.com/loredanacirstea/mythos-tests/testdata/tinygo"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
)

func (suite *KeeperTestSuite) TestMCPRegistry() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	// registryAddress := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_MCP_REGISTRY))

	// Prepare init data for mcp-execute contract
	executeInitData := &MCPContractInitGenesis{
		InitGenesis: &MCPContractInitGenesisRequest{
			RoutePrefix: "/tools/execute",
		},
	}
	executeInitDataJSON, _ := json.Marshal(executeInitData)

	// Instantiate mcp-execute with init data
	executeCodeId := appA.StoreCode(sender, testdata.MCPExecute, nil)
	executeAddress := appA.InstantiateCode(sender, executeCodeId, types.WasmxExecutionMessage{Data: executeInitDataJSON}, "mcp_execute", nil)
	fmt.Println("Instantiated mcp-execute contract:", executeAddress.String())

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

	// Assign MCP role - this will trigger RoleChanged hook which automatically registers with MCP registry
	fmt.Println("Assigning MCP role to execute contract...")
	utils.RegisterRole(suite, appA, types.ROLE_MCP, executeAddress, sender)
	fmt.Println("MCP role assigned to execute contract - auto-registration triggered via RoleChanged hook")

	fmt.Println("Assigning MCP role to userdata contract...")
	utils.RegisterRole(suite, appA, types.ROLE_MCP, userAddress, sender)
	fmt.Println("MCP role assigned to userdata contract - auto-registration triggered via RoleChanged hook")

	// Start the HTTP server
	httpRegistryAddr := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_HTTPSERVER_REGISTRY))
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

	suite.T().Log("MCP registry server running on :8080 with registered tool contracts... Press Ctrl+C to exit")

	// Create a channel to listen for interrupt/terminate signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-sig

	suite.T().Log("Received exit signal. Test ending.")
}

// Calldata structures for MCP tool contract initialization
type MCPContractInitGenesis struct {
	InitGenesis *MCPContractInitGenesisRequest `json:"init_genesis,omitempty"`
}

type MCPContractInitGenesisRequest struct {
	RoutePrefix string `json:"route_prefix"`
}

// HTTP server structures
type StartWebServerCalldata struct {
	StartWebServer *StartWebServerRequest `json:"start_web_server,omitempty"`
}

type StartWebServerRequest struct {
	Config WebsrvConfig `json:"config"`
}

type WebsrvConfig struct {
	EnableOAuth        bool     `json:"enable_oauth"`
	Address            string   `json:"address"`
	CORSAllowedOrigins []string `json:"cors_allowed_origins"`
	CORSAllowedMethods []string `json:"cors_allowed_methods"`
	CORSAllowedHeaders []string `json:"cors_allowed_headers"`
	MaxOpenConnections int64    `json:"max_open_connections"`
	RequestBodyMaxSize int64    `json:"request_body_max_size"`
}

type StartWebServerResponse struct {
	Error string `json:"error"`
}
