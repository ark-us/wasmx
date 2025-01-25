package cmdutils

// DONTCOVER

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	tmconfig "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/rand"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cometbft/cometbft/types"
	tmtime "github.com/cometbft/cometbft/types/time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	pruningtypes "cosmossdk.io/store/pruning/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdkconfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	address "github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	dbm "github.com/cosmos/cosmos-db"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p/core/crypto"

	pvm "github.com/cometbft/cometbft/privval"

	app "github.com/loredanacirstea/wasmx/app"
	mcodec "github.com/loredanacirstea/wasmx/codec"
	mcfg "github.com/loredanacirstea/wasmx/config"
	mctx "github.com/loredanacirstea/wasmx/context"
	menc "github.com/loredanacirstea/wasmx/encoding"
	"github.com/loredanacirstea/wasmx/multichain"
	"github.com/loredanacirstea/wasmx/server"
	config "github.com/loredanacirstea/wasmx/server/config"
	websrvconfig "github.com/loredanacirstea/wasmx/x/websrv/server/config"
	websrvflags "github.com/loredanacirstea/wasmx/x/websrv/server/flags"

	jsonrpcconfig "github.com/loredanacirstea/wasmx/x/wasmx/server/config"
	jsonrpcflags "github.com/loredanacirstea/wasmx/x/wasmx/server/flags"

	cosmosmodtypes "github.com/loredanacirstea/wasmx/x/cosmosmod/types"
	networksrvflags "github.com/loredanacirstea/wasmx/x/network/server/flags"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

var (
	flagNodeDirPrefix      = "node-dir-prefix"
	flagNumValidators      = "v"
	flagOutputDir          = "output-dir"
	flagNodeDaemonHome     = "node-daemon-home"
	flagStartingIPAddress  = "starting-ip-address"
	flagSameMachine        = "same-machine"
	flagNoCors             = "nocors"
	flagEnableLogging      = "enable-logging"
	flagRPCAddress         = "rpc.address"
	flagAPIAddress         = "api.address"
	flagPrintMnemonic      = "print-mnemonic"
	flagP2P                = "libp2p"
	flagMinLevelValidators = "min-level-validators"
	flagEnableEIDCheck     = "enable-eid"
)

type initArgs struct {
	algo               string
	chainID            string
	keyringBackend     string
	minGasPrices       string
	nodeDaemonHome     string
	nodeDirPrefix      string
	numValidators      int
	outputDir          string
	startingIPAddress  string
	sameMachine        bool
	noCors             bool
	p2p                bool
	minLevelValidators int
	enableEIDCheck     bool
	initialChains      []string
}

type startArgs struct {
	algo              string
	apiAddress        string
	chainID           string
	grpcAddress       string
	minGasPrices      string
	outputDir         string
	rpcAddress        string
	numValidators     int
	enableLogging     bool
	printMnemonic     bool
	websrvEnable      bool
	WebsrvEnableOAuth bool
	websrvAddress     string

	jsonRpcEnable    bool
	jsonRpcAddress   string
	jsonRpcWsAddress string

	p2p bool
}

func addTestnetFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().Int(flagNumValidators, 4, "Number of validators to initialize the testnet with")
	cmd.Flags().StringP(flagOutputDir, "o", "./.testnets", "Directory to store initialization data for the testnet")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(sdkserver.FlagMinGasPrices, fmt.Sprintf("0.000006%s", mcfg.BaseDenom), "Minimum gas prices to accept for transactions; All fees in a tx must meet this minimum (e.g. 10000amyt)")
	cmd.Flags().String(flags.FlagKeyAlgorithm, string(hd.Secp256k1Type), "Key signing algorithm to generate keys for")
	cmd.Flags().Bool(flagP2P, false, "wether the consensus algorithm uses libp2p or not")
	cmd.Flags().Int(flagMinLevelValidators, 2, "minimum number of validators for chain levels")
	cmd.Flags().Bool(flagEnableEIDCheck, false, "enable eID checks")
}

// NewTestnetCmd creates a root testnet command with subcommands to run an in-process testnet or initialize
// validator configuration files for running a multi-validator testnet in a separate process
func NewTestnetCmd(wasmVmMeta memc.IWasmVmMeta, mbm module.BasicManager, genBalIterator cosmosmodtypes.GenesisBalancesIterator) *cobra.Command {
	testnetCmd := &cobra.Command{
		Use:                        "testnet",
		Short:                      "subcommands for starting or configuring local testnets",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	testnetCmd.AddCommand(testnetStartCmd(wasmVmMeta))
	testnetCmd.AddCommand(testnetInitFilesCmd(wasmVmMeta, mbm, genBalIterator))
	testnetCmd.AddCommand(testnetAddNodeCmd(wasmVmMeta, mbm, genBalIterator))
	testnetCmd.AddCommand(testnetCreateHierarchyCmd(mbm, genBalIterator))

	return testnetCmd
}

// get cmd to initialize all files for tendermint testnet and application
func testnetInitFilesCmd(wasmVmMeta memc.IWasmVmMeta, mbm module.BasicManager, genBalIterator cosmosmodtypes.GenesisBalancesIterator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init-files",
		Short: "Initialize config directories & files for a multi-validator testnet running locally via separate processes (e.g. Docker Compose or similar)",
		Long: `init-files will setup "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.) for running "v" validator nodes.

Booting up a network with these validator folders is intended to be used with Docker Compose,
or a similar setup where each node has a manually configurable IP address.

Note, strict routability for addresses is turned off in the config file.

Example:
	mythosd testnet init-files --network.initial-chains=mythos,level0 --v 4 --output-dir ./.testnets --starting-ip-address 192.168.10.2
	`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			serverCtx := server.GetServerContextFromCmd(cmd)

			args := initArgs{}
			args.outputDir, _ = cmd.Flags().GetString(flagOutputDir)
			args.keyringBackend, _ = cmd.Flags().GetString(flags.FlagKeyringBackend)
			args.chainID, _ = cmd.Flags().GetString(flags.FlagChainID)
			args.minGasPrices, _ = cmd.Flags().GetString(sdkserver.FlagMinGasPrices)
			args.nodeDirPrefix, _ = cmd.Flags().GetString(flagNodeDirPrefix)
			args.nodeDaemonHome, _ = cmd.Flags().GetString(flagNodeDaemonHome)
			args.startingIPAddress, _ = cmd.Flags().GetString(flagStartingIPAddress)
			args.sameMachine, _ = cmd.Flags().GetBool(flagSameMachine)
			args.noCors, _ = cmd.Flags().GetBool(flagNoCors)
			args.numValidators, _ = cmd.Flags().GetInt(flagNumValidators)
			args.algo, _ = cmd.Flags().GetString(flags.FlagKeyAlgorithm)
			args.p2p, _ = cmd.Flags().GetBool(flagP2P)
			args.minLevelValidators, _ = cmd.Flags().GetInt(flagMinLevelValidators)
			args.enableEIDCheck, _ = cmd.Flags().GetBool(flagEnableEIDCheck)
			args.initialChains, _ = cmd.Flags().GetStringSlice(networksrvflags.NetworkInitialChains)
			return initTestnetFiles(wasmVmMeta, clientCtx, cmd, serverCtx.Config, mbm, genBalIterator, args)
		},
	}

	addTestnetFlagsToCmd(cmd)
	cmd.Flags().String(flagNodeDirPrefix, "node", "Prefix the directory name for each node with (node results in node0, node1, ...)")
	cmd.Flags().String(flagNodeDaemonHome, "mythosd", "Home directory of the node's daemon configuration")
	cmd.Flags().String(flagStartingIPAddress, "192.168.0.1", "Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	cmd.Flags().Bool(flagSameMachine, false, "Starting nodes on the same machine, on different ports")
	cmd.Flags().Bool(flagNoCors, false, "If present, sets cors to *")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")
	cmd.Flags().StringSlice(networksrvflags.NetworkInitialChains, []string{"mythos"}, "Initialized chains, separated by comma. E.g. 'mythos', 'mythos,level0'")
	return cmd
}

func testnetAddNodeCmd(wasmVmMeta memc.IWasmVmMeta, mbm module.BasicManager, genBalIterator cosmosmodtypes.GenesisBalancesIterator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-node [node_index] [leader_uri]",
		Args:  cobra.ExactArgs(2),
		Short: "Initialize config directories & files for a new testnet node in an already initialized testnet",
		Long: `add-node will setup one directory and populate it with
necessary files (private validator, genesis, config, etc.)

Example:
	mythosd testnet add-node 2 "mythos12vh8m68santvwvvez7s8668mnmjtm93jy6e8m7@/ip4/127.0.0.1/tcp/5001/p2p/12D3KooWDE3DBBmziE7twUVeAshN2gmNFtEpw2xtwZE3LRo9kxsf"  --output-dir ./.testnets
	`,
		RunE: func(cmd *cobra.Command, args_ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			serverCtx := server.GetServerContextFromCmd(cmd)

			args := initArgs{}
			args.outputDir, _ = cmd.Flags().GetString(flagOutputDir)
			args.keyringBackend, _ = cmd.Flags().GetString(flags.FlagKeyringBackend)
			args.chainID, _ = cmd.Flags().GetString(flags.FlagChainID)
			args.minGasPrices, _ = cmd.Flags().GetString(sdkserver.FlagMinGasPrices)
			args.nodeDirPrefix, _ = cmd.Flags().GetString(flagNodeDirPrefix)
			args.nodeDaemonHome, _ = cmd.Flags().GetString(flagNodeDaemonHome)
			args.startingIPAddress, _ = cmd.Flags().GetString(flagStartingIPAddress)
			args.sameMachine, _ = cmd.Flags().GetBool(flagSameMachine)
			args.noCors, _ = cmd.Flags().GetBool(flagNoCors)
			args.numValidators, _ = cmd.Flags().GetInt(flagNumValidators)
			args.algo, _ = cmd.Flags().GetString(flags.FlagKeyAlgorithm)
			args.p2p, _ = cmd.Flags().GetBool(flagP2P)
			args.initialChains, _ = cmd.Flags().GetStringSlice(networksrvflags.NetworkInitialChains)

			nodeIndex, err := strconv.Atoi(args_[0])
			if err != nil {
				panic(err)
			}
			leaderURI := args_[1]

			return testnetAddNode(wasmVmMeta, clientCtx, cmd, serverCtx.Config, mbm, genBalIterator, args, nodeIndex, leaderURI)
		},
	}

	addTestnetFlagsToCmd(cmd)
	cmd.Flags().String(flagNodeDirPrefix, "node", "Prefix the directory name for each node with (node results in node0, node1, ...)")
	cmd.Flags().String(flagNodeDaemonHome, "mythosd", "Home directory of the node's daemon configuration")
	cmd.Flags().String(flagStartingIPAddress, "192.168.0.1", "Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	cmd.Flags().Bool(flagSameMachine, false, "Starting nodes on the same machine, on different ports")
	cmd.Flags().Bool(flagNoCors, false, "If present, sets cors to *")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")
	cmd.Flags().StringSlice(networksrvflags.NetworkInitialChains, []string{"mythos"}, "Initialized chains, separated by comma. E.g. 'mythos', 'mythos,level0'")

	return cmd
}

// get cmd to start multi validator in-process testnet
func testnetStartCmd(wasmVmMeta memc.IWasmVmMeta) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Launch an in-process multi-validator testnet",
		Long: `testnet will launch an in-process multi-validator testnet,
and generate "v" directories, populated with necessary validator configuration files
(private validator, genesis, config, etc.).

Example:
	mythosd testnet --v 4 --output-dir ./.testnets
	`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			args := startArgs{}
			args.outputDir, _ = cmd.Flags().GetString(flagOutputDir)
			args.chainID, _ = cmd.Flags().GetString(flags.FlagChainID)
			args.minGasPrices, _ = cmd.Flags().GetString(sdkserver.FlagMinGasPrices)
			args.numValidators, _ = cmd.Flags().GetInt(flagNumValidators)
			args.algo, _ = cmd.Flags().GetString(flags.FlagKeyAlgorithm)
			args.enableLogging, _ = cmd.Flags().GetBool(flagEnableLogging)
			args.rpcAddress, _ = cmd.Flags().GetString(flagRPCAddress)
			args.apiAddress, _ = cmd.Flags().GetString(flagAPIAddress)
			args.printMnemonic, _ = cmd.Flags().GetBool(flagPrintMnemonic)
			args.websrvEnable, _ = cmd.Flags().GetBool(websrvflags.WebsrvEnable)
			args.WebsrvEnableOAuth, _ = cmd.Flags().GetBool(websrvflags.WebsrvEnableOAuth)
			args.websrvAddress, _ = cmd.Flags().GetString(websrvflags.WebsrvAddress)
			args.jsonRpcEnable, _ = cmd.Flags().GetBool(jsonrpcflags.JsonRpcEnable)
			args.jsonRpcAddress, _ = cmd.Flags().GetString(jsonrpcflags.JsonRpcAddress)
			args.jsonRpcWsAddress, _ = cmd.Flags().GetString(jsonrpcflags.JsonRpcWsAddress)
			args.p2p, _ = cmd.Flags().GetBool(flagP2P)

			return startTestnet(cmd, args, wasmVmMeta)
		},
	}

	addTestnetFlagsToCmd(cmd)
	cmd.Flags().Bool(flagEnableLogging, false, "Enable INFO logging of tendermint validator nodes")
	cmd.Flags().String(flagRPCAddress, "tcp://0.0.0.0:26657", "the RPC address to listen on")
	cmd.Flags().String(flagAPIAddress, "tcp://0.0.0.0:1317", "the address to listen on for REST API")
	cmd.Flags().Bool(flagPrintMnemonic, true, "print mnemonic of first validator to stdout for manual testing")
	cmd.Flags().Bool(websrvflags.WebsrvEnable, true, "Define if the websrv web server should be enabled")
	cmd.Flags().Bool(websrvflags.WebsrvEnableOAuth, true, "Define if the websrv oauth server should be enabled")
	cmd.Flags().String(websrvflags.WebsrvAddress, websrvconfig.DefaultWebsrvAddress, "the websrv web server address to listen on")

	cmd.Flags().Bool(jsonrpcflags.JsonRpcEnable, true, "Define if the json-rpc server should be enabled")
	cmd.Flags().String(jsonrpcflags.JsonRpcAddress, jsonrpcconfig.DefaultJsonRpcAddress, "the json-rpc server address to listen on")
	cmd.Flags().String(jsonrpcflags.JsonRpcWsAddress, jsonrpcconfig.DefaultJsonRpcWsAddress, "the json-rpc websocket server address to listen on")
	cmd.Flags().Duration(jsonrpcflags.JsonRpcEVMTimeout, jsonrpcconfig.DefaultEVMTimeout, "Sets a timeout used for eth_call (0=infinite)")
	cmd.Flags().Duration(jsonrpcflags.JsonRpcHTTPTimeout, jsonrpcconfig.DefaultHTTPTimeout, "Sets a read/write timeout for json-rpc http server (0=infinite)")
	cmd.Flags().Duration(jsonrpcflags.JsonRpcHTTPIdleTimeout, jsonrpcconfig.DefaultHTTPIdleTimeout, "Sets a idle timeout for json-rpc http server (0=infinite)")
	cmd.Flags().Bool(jsonrpcflags.JsonRpcAllowUnprotectedTxs, jsonrpcconfig.DefaultAllowUnprotectedTxs, "Allow for unprotected (non EIP155 signed) transactions to be submitted via the node's RPC when the global parameter is disabled")
	cmd.Flags().Int(jsonrpcflags.JsonRpcMaxOpenConnections, jsonrpcconfig.DefaultMaxOpenConnections, "Sets the maximum number of simultaneous connections for the server listener")
	return cmd
}

func testnetCreateHierarchyCmd(mbm module.BasicManager, genBalIterator cosmosmodtypes.GenesisBalancesIterator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-levels [level_count] [validator_count_per_level]",
		Args:  cobra.ExactArgs(2),
		Short: "Initialize hierarchical chains up to given number of levels",
		Long: `create-levels will initialize hierarchical chains up to given number of levels

Example:
	mythosd testnet create-levels 2 2 --output-dir ./.testnets
	`,
		RunE: func(cmd *cobra.Command, args_ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			serverCtx := server.GetServerContextFromCmd(cmd)

			args := initArgs{}
			args.outputDir, _ = cmd.Flags().GetString(flagOutputDir)
			args.keyringBackend, _ = cmd.Flags().GetString(flags.FlagKeyringBackend)
			args.chainID, _ = cmd.Flags().GetString(flags.FlagChainID)
			args.minGasPrices, _ = cmd.Flags().GetString(sdkserver.FlagMinGasPrices)
			args.nodeDirPrefix, _ = cmd.Flags().GetString(flagNodeDirPrefix)
			args.nodeDaemonHome, _ = cmd.Flags().GetString(flagNodeDaemonHome)
			args.startingIPAddress, _ = cmd.Flags().GetString(flagStartingIPAddress)
			args.sameMachine, _ = cmd.Flags().GetBool(flagSameMachine)
			args.noCors, _ = cmd.Flags().GetBool(flagNoCors)
			args.numValidators, _ = cmd.Flags().GetInt(flagNumValidators)
			args.algo, _ = cmd.Flags().GetString(flags.FlagKeyAlgorithm)
			args.p2p, _ = cmd.Flags().GetBool(flagP2P)

			levels, err := strconv.Atoi(args_[0])
			if err != nil {
				panic(err)
			}

			validatorCountPerLevel, err := strconv.Atoi(args_[1])
			if err != nil {
				panic(err)
			}

			return testnetCreateHierarchy(clientCtx, cmd, serverCtx.Config, mbm, genBalIterator, args, levels, validatorCountPerLevel)
		},
	}

	addTestnetFlagsToCmd(cmd)
	cmd.Flags().String(flagNodeDirPrefix, "node", "Prefix the directory name for each node with (node results in node0, node1, ...)")
	cmd.Flags().String(flagNodeDaemonHome, "mythosd", "Home directory of the node's daemon configuration")
	cmd.Flags().String(flagStartingIPAddress, "192.168.0.1", "Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	cmd.Flags().Bool(flagSameMachine, false, "Starting nodes on the same machine, on different ports")
	cmd.Flags().Bool(flagNoCors, false, "If present, sets cors to *")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")

	return cmd
}

const nodeDirPerm = 0o755

// const p2pListenAddress = "26656"
// const p2pListenAddressMulti = 26756
// replace with network module // TODO maybe have network module on 26656
const p2pListenAddress = "8090"
const p2pListenAddressMulti = 8090

// initTestnetFiles initializes testnet files for a testnet to be run in a separate process
func initTestnetFiles(
	wasmVmMeta memc.IWasmVmMeta,
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *tmconfig.Config,
	mbm module.BasicManager,
	genBalIterator cosmosmodtypes.GenesisBalancesIterator,
	args initArgs,
) error {
	os.RemoveAll(args.outputDir)
	os.MkdirAll(args.outputDir, nodeDirPerm)
	return initTestnetFilesInternal(wasmVmMeta, clientCtx, cmd, nodeConfig, mbm, genBalIterator, args, 0, "")
}

func testnetAddNode(
	wasmVmMeta memc.IWasmVmMeta,
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *tmconfig.Config,
	mbm module.BasicManager,
	genBalIterator cosmosmodtypes.GenesisBalancesIterator,
	args initArgs,
	nodeIndex int,
	leaderURI string,
) error {
	generateMythos := slices.Contains(args.initialChains, "mythos")

	args.numValidators = nodeIndex + 1
	args.sameMachine = true
	err := initTestnetFilesInternal(wasmVmMeta, clientCtx, cmd, nodeConfig, mbm, genBalIterator, args, nodeIndex, leaderURI)
	if err != nil {
		return err
	}

	// first get the node0 genesis file path
	nodeDirName := fmt.Sprintf("%s%d", args.nodeDirPrefix, 0)
	nodeDir := filepath.Join(args.outputDir, nodeDirName, args.nodeDaemonHome)
	nodeConfig.SetRoot(nodeDir)
	genesisFile0 := nodeConfig.GenesisFile()

	// our current node genesis file path
	nodeDirName = fmt.Sprintf("%s%d", args.nodeDirPrefix, nodeIndex)
	nodeDir = filepath.Join(args.outputDir, nodeDirName, args.nodeDaemonHome)
	nodeConfig.SetRoot(nodeDir)
	genesisFileNew := nodeConfig.GenesisFile()

	// copy genesis from node0 into our node
	bz, err := os.ReadFile(genesisFile0)
	if err != nil {
		return err
	}
	err = os.WriteFile(genesisFileNew, bz, 0o644)
	if err != nil {
		return err
	}

	if generateMythos {
		genesisFileMythos := strings.Replace(genesisFile0, ".json", "_"+args.chainID+".json", 1)
		genesisFileNewMythos := strings.Replace(genesisFileNew, ".json", "_"+args.chainID+".json", 1)

		bz, err = os.ReadFile(genesisFileMythos)
		if err != nil {
			return err
		}
		err = os.WriteFile(genesisFileNewMythos, bz, 0o644)
		if err != nil {
			return err
		}
	}
	return nil
}

// initTestnetFiles initializes testnet files for a testnet to be run in a separate process
func initTestnetFilesInternal(
	wasmVmMeta memc.IWasmVmMeta,
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *tmconfig.Config,
	mbm module.BasicManager,
	genBalIterator cosmosmodtypes.GenesisBalancesIterator,
	args initArgs,
	nodeIndexStart int,
	leaderURI string,
) error {
	var err error
	mockNodeHome := path.Join(args.outputDir, "tmp")
	generateMythos := slices.Contains(args.initialChains, "mythos")
	generateLevel0 := slices.Contains(args.initialChains, "level0")
	if args.chainID == "" {
		args.chainID = fmt.Sprintf("mythos_%d-1", tmrand.Int63n(9999999999999)+1)
	}
	chainId := args.chainID
	chaincfg, err := mcfg.GetChainConfig(args.chainID)
	if err != nil {
		panic(err)
	}
	mcfg.SetGlobalChainConfigById(args.chainID)

	addrCodec := mcodec.NewAccBech32Codec(chaincfg.Bech32PrefixAccAddr, mcodec.NewAddressPrefixedFromAcc)
	valAddrCodec := mcodec.NewValBech32Codec(chaincfg.Bech32PrefixValAddr, mcodec.NewAddressPrefixedFromVal)
	_, appCreator := createMockAppCreator(wasmVmMeta, mockNodeHome)

	chaincfgmyt, err := mcfg.GetChainConfig(chainId)
	if err != nil {
		panic(err)
	}
	imythosapp := appCreator(chainId, chaincfgmyt)
	mythosapp := imythosapp.(*app.App)

	chainId0 := mcfg.LEVEL0_CHAIN_ID
	chaincfg0, err := mcfg.GetChainConfig(chainId0)
	if err != nil {
		panic(err)
	}
	ilevel0app := appCreator(chainId0, chaincfg0)
	level0app := ilevel0app.(*app.App)
	addrCodec0 := level0app.TxConfig().SigningContext().AddressCodec()

	initialChainIds := []string{}
	if generateMythos {
		initialChainIds = append(initialChainIds, chainId)
	}
	if generateLevel0 {
		initialChainIds = append(initialChainIds, chainId0)
	}
	nodeIDs := make([]string, args.numValidators)
	valPubKeys := make([]cryptotypes.PubKey, args.numValidators)
	nodeIPs := make([]string, args.numValidators)

	appConfig := config.DefaultConfig()
	appConfig.MinGasPrices = args.minGasPrices
	appConfig.API.Enable = true
	appConfig.API.RPCReadTimeout = 30
	appConfig.API.RPCWriteTimeout = 30
	appConfig.API.Address = "tcp://0.0.0.0:1317"
	appConfig.Telemetry.Enabled = true
	appConfig.Telemetry.PrometheusRetentionTime = 60
	appConfig.Telemetry.EnableHostnameLabel = false
	appConfig.Telemetry.GlobalLabels = [][]string{{"chain_id", args.chainID}}
	if args.noCors {
		appConfig.API.EnableUnsafeCORS = true
	}

	genAccounts := make([]cosmosmodtypes.GenesisAccount, args.numValidators)
	genBalances := make([]banktypes.Balance, args.numValidators)
	genFiles := make([]string, args.numValidators)

	genAccounts0 := make([][]byte, args.numValidators)

	networkIps := make([]string, args.numValidators)
	networkIps0 := make([]string, args.numValidators)
	valaddr := make([]sdk.AccAddress, args.numValidators)
	nodeDirNames := make([]string, args.numValidators)
	nodeDirs := make([]string, args.numValidators)
	kbs := make([]keyring.Keyring, args.numValidators)
	inBuf := bufio.NewReader(cmd.InOrStdin())

	nodeOffset := int32(nodeIndexStart)
	portOffset := nodeOffset * int32(len(initialChainIds))

	// generate private keys, node IDs, and initial transactions
	for i := nodeIndexStart; i < args.numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", args.nodeDirPrefix, i)
		nodeDir := filepath.Join(args.outputDir, nodeDirName, args.nodeDaemonHome)

		nodeDirNames[i] = nodeDirName
		nodeDirs[i] = nodeDir

		if args.sameMachine {
			nodeOffset = int32(i)
			portOffset = nodeOffset * int32(len(initialChainIds))
		}
		ports := mctx.GetChainNodePorts(0, portOffset)
		ports0 := mctx.GetChainNodePorts(1, portOffset)

		nodeConfig.SetRoot(nodeDir)
		nodeConfig.RPC.ListenAddress = "tcp://0.0.0.0:" + strconv.Itoa(int(ports.TendermintRpc))
		if args.noCors {
			nodeConfig.RPC.CORSAllowedOrigins = []string{"*"}
		}
		nodeConfig.ProxyApp = nodeConfig.RPC.ListenAddress

		if err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm); err != nil {
			_ = os.RemoveAll(args.outputDir)
			return err
		}

		nodeConfig.Moniker = nodeDirName

		nodeIPs[i], err = getIP(i, args.startingIPAddress)
		if err != nil {
			return err
		}

		nodeIDs[i], valPubKeys[i], err = genutil.InitializeNodeValidatorFiles(nodeConfig)
		if err != nil {
			_ = os.RemoveAll(args.outputDir)
			return err
		}

		genFiles[i] = nodeConfig.GenesisFile()

		kb, err := keyring.New(sdk.KeyringServiceName(), args.keyringBackend, nodeDir, inBuf, clientCtx.Codec)
		if err != nil {
			return err
		}
		kbs[i] = kb

		keyringAlgos, _ := kb.SupportedAlgorithms()
		algo, err := keyring.NewSigningAlgoFromString(args.algo, keyringAlgos)
		if err != nil {
			return err
		}

		addr, secret, err := testutil.GenerateSaveCoinKey(kb, nodeDirName, "", true, algo)
		if err != nil {
			_ = os.RemoveAll(args.outputDir)
			return err
		}

		info := map[string]string{"secret": secret}

		cliPrint, err := json.Marshal(info)
		if err != nil {
			return err
		}

		// save private key seed words
		if err := writeFile(fmt.Sprintf("%v.json", "key_seed"), nodeDir, cliPrint); err != nil {
			return err
		}

		addrstr, err := addrCodec.BytesToString(addr)
		if err != nil {
			return err
		}
		addrstr0, err := addrCodec0.BytesToString(addr)
		if err != nil {
			return err
		}

		host := fmt.Sprintf("%s@%s:%d", addrstr, nodeIPs[i], ports.WasmxNetworkGrpc)
		host0 := fmt.Sprintf("%s@%s:%d", addrstr0, nodeIPs[i], ports0.WasmxNetworkGrpc)

		if args.p2p {
			valStr := addrstr
			valStr0 := addrstr0

			privValid := pvm.LoadOrGenFilePV(nodeConfig.PrivValidatorKeyFile(), nodeConfig.PrivValidatorStateFile())
			pk := privValid.Key.PrivKey.Bytes()
			pkcrypto, err := crypto.UnmarshalEd25519PrivateKey(pk)
			if err != nil {
				return err
			}
			identity := libp2p.Identity(pkcrypto)

			ipaddr := "127.0.0.1"
			// ipaddr := nodeIPs[i]
			// ipaddr = strings.Replace(ipaddr, "0.0.0.0", "127.0.0.1", 1)
			node, err := libp2p.New(
				libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", ipaddr, ports.WasmxNetworkP2P)),
				libp2p.Ping(false),
				identity,
			)
			if err != nil {
				return err
			}
			p2pid := node.ID()
			host = fmt.Sprintf("%s@/ip4/%s/tcp/%d/p2p/%s", valStr, ipaddr, ports.WasmxNetworkP2P, p2pid)
			host0 = fmt.Sprintf("%s@/ip4/%s/tcp/%d/p2p/%s", valStr0, ipaddr, ports0.WasmxNetworkP2P, p2pid)
		}

		valaddr[i] = addr
		networkIps[i] = host
		networkIps0[i] = host0
		genAccounts0[i] = addr.Bytes()
	}
	networkIpsStr := strings.Join(networkIps, ",")
	networkIpsStr = strings.Trim(networkIpsStr, ",")
	// leaderURI is last
	if leaderURI != "" {
		networkIpsStr = networkIpsStr + "," + leaderURI
	}

	basePow := int64(500)

	for i := nodeIndexStart; i < args.numValidators; i++ {
		gentxsDir := filepath.Join(args.outputDir, "gentxs")
		nodeDirName := nodeDirNames[i]
		nodeDir := nodeDirs[i]
		addr := valaddr[i]
		kb := kbs[i]
		memo := networkIps[i]

		powdelta := int64((args.numValidators - i) * 10)
		accStakingTokens := sdk.TokensFromConsensusPower(basePow+powdelta, mcfg.PowerReduction)
		coins := sdk.Coins{
			sdk.NewCoin(chaincfg.BaseDenom, accStakingTokens),
		}
		addrprefixed := mcodec.MustUnwrapAccBech32Codec(addrCodec).BytesToAccAddressPrefixed(addr)

		genBalances[i] = banktypes.Balance{Address: addrprefixed.String(), Coins: coins.Sort()}
		genAccounts[i] = cosmosmodtypes.NewBaseAccount(addrprefixed, nil, 0, 0)

		valStr, err := valAddrCodec.BytesToString(sdk.ValAddress(addr))
		if err != nil {
			return err
		}
		valTokens := sdk.TokensFromConsensusPower(basePow/5+powdelta, mcfg.PowerReduction)
		err = createGentx(clientCtx, mythosapp.TxConfig(), args.chainID, valStr, valPubKeys[i], valTokens, nodeDirName, gentxsDir, memo, chaincfg.BaseDenom, kb, cmd)
		if err != nil {
			return err
		}

		customAppTemplate, customAppConfig := config.AppConfig()
		srvconfig.SetConfigTemplate(customAppTemplate)

		serverCtx, err := InterceptConfigsAndCreateContext(cmd, customAppTemplate, customAppConfig, tmconfig.DefaultConfig())
		if err != nil {
			return err
		}
		logger := server.NewDefaultLogger(cmd.Flags())
		serverCtx.Logger = logger.With(log.ModuleKey, "server")
		err = sdkserver.SetCmdServerContext(cmd, serverCtx)
		if err != nil {
			return err
		}

		// value copy, not reference!
		appConfigCopy := *appConfig
		// we do not care about other ports, they are set by the application

		// current index for mythos, then level0
		appConfigCopy.Network.Id = fmt.Sprintf("%s:%d;%s:%d", chainId, int32(i-nodeIndexStart), chainId0, 0)

		networkIpsJointStr := fmt.Sprintf(`%s:%s;%s:%s`, chainId, networkIpsStr, chainId0, networkIps0[i])
		appConfigCopy.Network.Ips = networkIpsJointStr
		appConfigCopy.Network.InitialChains = initialChainIds

		srvconfig.WriteConfigFile(filepath.Join(nodeDir, "config/app.toml"), appConfigCopy)
		initClientConfig(clientCtx, nodeDir, i)
	}

	if nodeIndexStart == 0 {
		if generateMythos {
			if err := initGenFiles(clientCtx, mbm, args.chainID, genAccounts, genBalances, genFiles, args.numValidators, args.minLevelValidators, args.enableEIDCheck); err != nil {
				return err
			}

			for i := nodeIndexStart; i < args.numValidators; i++ {
				err = collectGenFiles(
					clientCtx,
					mythosapp.TxConfig(),
					nodeConfig,
					args.chainID, nodeIDs[i], valPubKeys[i], i,
					args.outputDir, args.nodeDirPrefix, args.nodeDaemonHome, genBalIterator, valAddrCodec, args.sameMachine, nodeConfig.GenesisFile(), "gentxs",
				)
				if err != nil {
					return err
				}
			}
		}
	}

	if generateLevel0 {
		// initialize level0
		genBalanceCoins := make([]sdk.Coins, len(genBalances))
		for i := 0; i < len(genBalances); i++ {
			genBalanceCoins[i] = genBalances[i].Coins
		}

		// set this only after we get address bytes
		mcfg.SetGlobalChainConfigById(chainId0)

		valAddrCodec = level0app.TxConfig().SigningContext().ValidatorAddressCodec()

		for i := nodeIndexStart; i < args.numValidators; i++ {
			// we add the nodeid to this folder, so we only have 1 gentx for each level0
			gentxsDir := filepath.Join(args.outputDir, "gentxs_"+chainId0+nodeIDs[i])
			nodeDirName := nodeDirNames[i]
			kb := kbs[i]
			// memo := fmt.Sprintf("%s@%s:%s", nodeIDs[i], nodeIPs[i], p2pListenAddress)
			memo := networkIps0[i]

			addrprefixed := mcodec.MustUnwrapAccBech32Codec(addrCodec0).BytesToAccAddressPrefixed(genAccounts0[i])
			coins := make([]sdk.Coin, len(genBalanceCoins[i]))
			for i, coin := range genBalanceCoins[i] {
				coins[i] = sdk.NewCoin(chaincfg0.BaseDenom, coin.Amount)
			}
			valStr, err := valAddrCodec.BytesToString(sdk.ValAddress(addrprefixed.Bytes()))
			if err != nil {
				return err
			}
			valTokens := sdk.TokensFromConsensusPower(100, mcfg.PowerReduction)
			err = createGentx(clientCtx, level0app.TxConfig(), chainId0, valStr, valPubKeys[i], valTokens, nodeDirName, gentxsDir, memo, chaincfg0.BaseDenom, kb, cmd)
			if err != nil {
				return err
			}
		}
		genBalances := []banktypes.Balance{}
		genAccounts := []cosmosmodtypes.GenesisAccount{}

		for i := nodeIndexStart; i < args.numValidators; i++ {
			addrprefixed := mcodec.MustUnwrapAccBech32Codec(addrCodec0).BytesToAccAddressPrefixed(genAccounts0[i])
			genAccount := cosmosmodtypes.NewBaseAccount(addrprefixed, nil, 0, 0)
			genAccounts = append(genAccounts, genAccount)
			addrstr := addrprefixed.String()

			coins := make([]sdk.Coin, len(genBalanceCoins[i]))
			for i, coin := range genBalanceCoins[i] {
				coins[i] = sdk.NewCoin(chaincfg0.BaseDenom, coin.Amount)
			}
			genBalance := banktypes.Balance{Address: addrstr, Coins: coins}
			genBalances = append(genBalances, genBalance)

			genFile := strings.Replace(genFiles[i], ".json", "_"+chainId0+".json", 1)

			if err := initGenFilesLevel0(clientCtx, mbm, chainId0, genAccount, genBalance, genFile, 1, args.minLevelValidators, args.enableEIDCheck); err != nil {
				return err
			}
			err = collectGenFiles(
				clientCtx,
				level0app.TxConfig(),
				nodeConfig, chainId0, nodeIDs[i], valPubKeys[i], i,
				args.outputDir, args.nodeDirPrefix, args.nodeDaemonHome, genBalIterator, valAddrCodec, args.sameMachine, genFile, "gentxs_"+chainId0+nodeIDs[i],
			)
			if err != nil {
				return err
			}
		}
	}

	// cleanup mocks
	err = os.RemoveAll(mockNodeHome)
	if err != nil {
		cmd.PrintErrf("could not remove temporary folders: %s", err.Error())
	}

	cmd.PrintErrf("Successfully initialized %d node directories\n", args.numValidators)
	return nil
}

func initClientConfig(ctx client.Context, nodeDir string, index int) {
	configFilePath := filepath.Join(nodeDir, "config/client.toml")
	conf := sdkconfig.DefaultConfig()
	conf.ChainID = ctx.ChainID
	conf.Node = fmt.Sprintf("tcp://localhost:%d", 26657+index)
	// conf.KeyringBackend = "os" // os|test
	// conf.Output = "text"
	// conf.BroadcastMode = "sync"
	sdkconfig.WriteConfigToFile(configFilePath, conf)
}

func createGentx(
	clientCtx client.Context,
	txConfig client.TxConfig,
	chainID string,
	valStr string, valPubKey cryptotypes.PubKey, valTokens math.Int,
	nodeDirName string, gentxsDir string, memo string,
	baseDenom string,
	kb keyring.Keyring,
	cmd *cobra.Command,
) error {
	createValMsg, err := stakingtypes.NewMsgCreateValidator(
		valStr,
		valPubKey,
		sdk.NewCoin(baseDenom, valTokens),
		stakingtypes.NewDescription(nodeDirName, "", "", "", ""),
		stakingtypes.NewCommissionRates(math.LegacyOneDec(), math.LegacyOneDec(), math.LegacyOneDec()),
		math.OneInt(),
	)
	if err != nil {
		return err
	}

	txBuilder := txConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(createValMsg); err != nil {
		return err
	}

	txBuilder.SetMemo(memo)

	txFactory := tx.Factory{}
	txFactory = txFactory.
		WithChainID(chainID).
		WithMemo(memo).
		WithKeybase(kb).
		WithTxConfig(txConfig)

	if err := tx.Sign(cmd.Context(), txFactory, nodeDirName, txBuilder, true); err != nil {
		return err
	}

	txBz, err := txConfig.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		return err
	}

	if err := writeFile(fmt.Sprintf("%v.json", nodeDirName), gentxsDir, txBz); err != nil {
		return err
	}
	return nil
}

func initGenFiles(
	clientCtx client.Context,
	mbm module.BasicManager,
	chainID string,
	genAccounts []cosmosmodtypes.GenesisAccount,
	genBalances []banktypes.Balance,
	genFiles []string,
	numValidators int,
	minLevelValidators int,
	enableEIDCheck bool,
) error {
	appGenState := mbm.DefaultGenesis(clientCtx.Codec)
	chaincfg, err := mcfg.GetChainConfig(chainID)
	if err != nil {
		panic(err)
	}

	addrCodec := mcodec.NewAccBech32Codec(chaincfg.Bech32PrefixAccAddr, mcodec.NewAddressPrefixedFromAcc)

	var cosmosmodGenState cosmosmodtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[cosmosmodtypes.ModuleName], &cosmosmodGenState)

	cosmosmodGenState.Bank.DenomInfo = cosmosmodtypes.DefaultBankDenoms(addrCodec.(mcodec.AccBech32Codec), mcfg.DenomUnit, uint32(mcfg.BaseDenomUnit))
	cosmosmodGenState.Bank.Balances = genBalances
	cosmosmodGenState.Staking.Params.BondDenom = mcfg.BondBaseDenom
	cosmosmodGenState.Staking.BaseDenom = mcfg.BaseDenom
	p, _ := math.LegacyNewDecFromStr("0.6")
	cosmosmodGenState.Slashing.Params.MinSignedPerWindow = p
	cosmosmodGenState.Slashing.Params.DowntimeJailDuration = time.Hour * 24
	cosmosmodGenState.Slashing.Params.SignedBlocksWindow = 40000
	// cosmosmodGenState.Slashing.Params.DowntimeJailDuration = time.Minute
	// cosmosmodGenState.Slashing.Params.SignedBlocksWindow = 4
	cosmosmodGenState.Distribution.BaseDenom = mcfg.BaseDenom
	cosmosmodGenState.Distribution.RewardsDenom = cosmosmodGenState.Bank.DenomInfo[2].Metadata.Base
	cosmosmodGenState.Gov.Params.MinDeposit[0].Denom = mcfg.BaseDenom
	cosmosmodGenState.Gov.Params.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewCoin(mcfg.BaseDenom, math.NewInt(50000000)))
	// TODO make this bigger once we have our own governance contract
	votingP := time.Minute * 2
	cosmosmodGenState.Gov.Params.VotingPeriod = votingP.Milliseconds()

	// set the accounts in the genesis state
	authGenesis, err := cosmosmodtypes.NewAuthGenesisStateFromCosmos(clientCtx.Codec, cosmosmodGenState.Auth.Params, genAccounts)
	if err != nil {
		return err
	}
	cosmosmodGenState.Auth = *authGenesis

	// set cosmosmod genesis
	appGenState[cosmosmodtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&cosmosmodGenState)

	var crisisGenState crisistypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[crisistypes.ModuleName], &crisisGenState)
	crisisGenState.ConstantFee.Denom = mcfg.BaseDenom
	appGenState[crisistypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&crisisGenState)

	var mintGenState minttypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[minttypes.ModuleName], &mintGenState)
	mintGenState.Params.MintDenom = mcfg.BaseDenom
	appGenState[minttypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&mintGenState)

	feeCollectorBech32, err := addrCodec.BytesToString(cosmosmodtypes.NewModuleAddress(wasmxtypes.FEE_COLLECTOR))
	if err != nil {
		panic(err)
	}
	mintAddressBech32, err := addrCodec.BytesToString(cosmosmodtypes.NewModuleAddress("mint"))
	if err != nil {
		panic(err)
	}

	bootstrapAccount, err := addrCodec.BytesToString(sdk.AccAddress(rand.Bytes(address.Len)))
	if err != nil {
		panic(err)
	}

	var wasmxGenState wasmxtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[wasmxtypes.ModuleName], &wasmxGenState)
	wasmxGenState.SystemContracts = wasmxtypes.DefaultSystemContracts(addrCodec.(mcodec.AccBech32Codec), feeCollectorBech32, mintAddressBech32, int32(minLevelValidators), enableEIDCheck, "{}")
	wasmxGenState.BootstrapAccountAddress = bootstrapAccount
	appGenState[wasmxtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&wasmxGenState)

	appGenStateJSON, err := json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return err
	}

	genDoc := types.GenesisDoc{
		ChainID:    chainID,
		AppState:   appGenStateJSON,
		Validators: nil,
	}

	// generate empty genesis files for each validator and save
	for i := 0; i < numValidators; i++ {
		if err := genDoc.SaveAs(genFiles[i]); err != nil {
			return err
		}
	}
	return nil
}

func initGenFilesLevel0(
	clientCtx client.Context,
	mbm module.BasicManager,
	chainID string,
	genAccount cosmosmodtypes.GenesisAccount,
	genBalance banktypes.Balance,
	genFile string,
	numValidators int,
	minLevelValidators int,
	enableEIDCheck bool,
) error {
	chaincfg, err := mcfg.GetChainConfig(chainID)
	if err != nil {
		panic(err)
	}
	appGenState := mbm.DefaultGenesis(clientCtx.Codec)

	addrCodec := mcodec.NewAccBech32Codec(chaincfg.Bech32PrefixAccAddr, mcodec.NewAddressPrefixedFromAcc)

	feeCollectorBech32, err := addrCodec.BytesToString(cosmosmodtypes.NewModuleAddress(wasmxtypes.FEE_COLLECTOR))
	if err != nil {
		panic(err)
	}
	mintAddressBech32, err := addrCodec.BytesToString(cosmosmodtypes.NewModuleAddress("mint"))
	if err != nil {
		panic(err)
	}

	bootstrapAccount, err := addrCodec.BytesToString(sdk.AccAddress(rand.Bytes(address.Len)))
	if err != nil {
		panic(err)
	}

	var wasmxGenState wasmxtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[wasmxtypes.ModuleName], &wasmxGenState)
	wasmxGenState.SystemContracts = wasmxtypes.DefaultTimeChainContracts(addrCodec.(mcodec.AccBech32Codec), feeCollectorBech32, mintAddressBech32, int32(minLevelValidators), enableEIDCheck, "{}")
	wasmxGenState.BootstrapAccountAddress = bootstrapAccount
	appGenState[wasmxtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&wasmxGenState)

	var cosmosmodGenState cosmosmodtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[cosmosmodtypes.ModuleName], &cosmosmodGenState)

	cosmosmodGenState.Bank.DenomInfo = cosmosmodtypes.DefaultBankDenoms(addrCodec.(mcodec.AccBech32Codec), chaincfg.DenomUnit, uint32(chaincfg.BaseDenomUnit))
	cosmosmodGenState.Bank.Balances = []banktypes.Balance{genBalance}
	cosmosmodGenState.Staking.Params.BondDenom = chaincfg.BondBaseDenom
	cosmosmodGenState.Staking.BaseDenom = chaincfg.BaseDenom
	cosmosmodGenState.Distribution.BaseDenom = chaincfg.BaseDenom
	cosmosmodGenState.Distribution.RewardsDenom = cosmosmodGenState.Bank.DenomInfo[2].Metadata.Base
	cosmosmodGenState.Gov.Params.MinDeposit[0].Denom = chaincfg.BaseDenom
	cosmosmodGenState.Gov.Params.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewCoin(chaincfg.BaseDenom, math.NewInt(50000000)))
	// TODO make this bigger once we have our own governance contract
	votingP := time.Minute * 2
	cosmosmodGenState.Gov.Params.VotingPeriod = votingP.Milliseconds()

	// set the accounts in the genesis state
	authGenesis, err := cosmosmodtypes.NewAuthGenesisStateFromCosmos(clientCtx.Codec, cosmosmodGenState.Auth.Params, []cosmosmodtypes.GenesisAccount{genAccount})
	if err != nil {
		return err
	}
	cosmosmodGenState.Auth = *authGenesis

	// set cosmosmod genesis
	appGenState[cosmosmodtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&cosmosmodGenState)

	var crisisGenState crisistypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[crisistypes.ModuleName], &crisisGenState)
	crisisGenState.ConstantFee.Denom = chaincfg.BaseDenom
	appGenState[crisistypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&crisisGenState)

	var mintGenState minttypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[minttypes.ModuleName], &mintGenState)
	mintGenState.Params.MintDenom = chaincfg.BaseDenom
	appGenState[minttypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&mintGenState)

	appGenStateJSON, err := json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return err
	}

	genDoc := types.GenesisDoc{
		ChainID:    chainID,
		AppState:   appGenStateJSON,
		Validators: nil,
	}

	// generate empty genesis file
	if err := genDoc.SaveAs(genFile); err != nil {
		return err
	}
	return nil
}

func collectGenFiles(
	clientCtx client.Context,
	txConfig client.TxConfig,
	nodeConfig *tmconfig.Config,
	chainID string,
	nodeID string, valPubKey cryptotypes.PubKey,
	i int,
	outputDir, nodeDirPrefix, nodeDaemonHome string, genBalIterator cosmosmodtypes.GenesisBalancesIterator, valAddrCodec runtime.ValidatorAddressCodec,
	sameMachine bool,
	genFile string, genTxDirName string,
) error {
	genTime := tmtime.Now()

	nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
	nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)
	// we add the nodeid to this folder, so we only have 1 gentx for each level0
	gentxsDir := filepath.Join(outputDir, genTxDirName)
	nodeConfig.Moniker = nodeDirName
	nodeConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
	// if sameMachine {
	// 	nodeConfig.RPC.ListenAddress = "tcp://0.0.0.0:" + strconv.Itoa(26657+i)
	// }
	nodeConfig.ProxyApp = nodeConfig.RPC.ListenAddress

	nodeConfig.SetRoot(nodeDir)
	initCfg := genutiltypes.NewInitConfig(chainID, gentxsDir, nodeID, valPubKey)

	appGenesis, err := genutiltypes.AppGenesisFromFile(genFile)
	if err != nil {
		return err
	}

	nodeAppState, err := GenAppStateFromConfig(clientCtx.Codec, txConfig, nodeConfig, initCfg, appGenesis, genBalIterator, genutiltypes.DefaultMessageValidator, valAddrCodec, genFile)
	if err != nil {
		return err
	}

	// overwrite each validator's genesis file to have a canonical genesis time
	if err := genutil.ExportGenesisFileWithTime(genFile, chainID, nil, nodeAppState, genTime); err != nil {
		return err
	}

	if !strings.Contains(genFile, chainID) {
		newGenFile := strings.Replace(genFile, ".json", "_"+chainID+".json", 1)
		copyFile(genFile, newGenFile)
	}

	return nil
}

func getIP(i int, startingIPAddr string) (ip string, err error) {
	if len(startingIPAddr) == 0 {
		ip, err = sdkserver.ExternalIP()
		if err != nil {
			return "", err
		}
		return ip, nil
	}
	return calculateIP(startingIPAddr, i)
}

func calculateIP(ip string, i int) (string, error) {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return "", fmt.Errorf("%v: non ipv4 address", ip)
	}

	for j := 0; j < i; j++ {
		ipv4[3]++
	}

	return ipv4.String(), nil
}

func writeFile(name string, dir string, contents []byte) error {
	writePath := filepath.Join(dir)
	file := filepath.Join(writePath, name)

	err := EnsureDir(writePath, 0o755)
	if err != nil {
		return err
	}

	err = os.WriteFile(file, contents, 0o644)
	if err != nil {
		return err
	}

	return nil
}

// startTestnet starts an in-process testnet
func startTestnet(cmd *cobra.Command, args startArgs, wasmVmMeta memc.IWasmVmMeta) error {
	networkConfig := network.DefaultConfig(app.NewTestNetworkFixture(wasmVmMeta))

	// Default networkConfig.ChainID is random, and we should only override it if chainID provided
	// is non-empty
	if args.chainID != "" {
		networkConfig.ChainID = args.chainID
	}
	networkConfig.SigningAlgo = args.algo
	networkConfig.MinGasPrices = args.minGasPrices
	networkConfig.NumValidators = args.numValidators
	networkConfig.EnableLogging = args.enableLogging
	networkConfig.RPCAddress = args.rpcAddress
	networkConfig.APIAddress = args.apiAddress
	networkConfig.GRPCAddress = args.grpcAddress
	networkConfig.PrintMnemonic = args.printMnemonic
	networkLogger := network.NewCLILogger(cmd)

	baseDir := fmt.Sprintf("%s/%s", args.outputDir, networkConfig.ChainID)
	if _, err := os.Stat(baseDir); !os.IsNotExist(err) {
		return fmt.Errorf(
			"testnests directory already exists for chain-id '%s': %s, please remove or select a new --chain-id",
			networkConfig.ChainID, baseDir)
	}

	testnet, err := network.New(networkLogger, baseDir, networkConfig)
	if err != nil {
		return err
	}

	_, err = testnet.WaitForHeight(1)
	if err != nil {
		return err
	}

	cmd.Println("press the Enter Key to terminate")
	_, err = fmt.Scanln() // wait for Enter Key
	if err != nil {
		return err
	}
	testnet.Cleanup()

	return nil
}

// EnsureDir ensures the given directory exists, creating it if necessary.
// Errors if the path already exists as a non-directory.
func EnsureDir(dir string, mode os.FileMode) error {
	err := os.MkdirAll(dir, mode)
	if err != nil {
		return fmt.Errorf("could not create directory %q: %w", dir, err)
	}
	return nil
}

func copyFile(sourceFile string, destinationFile string) error {
	input, err := os.ReadFile(sourceFile)
	if err != nil {
		return err
	}
	err = os.WriteFile(destinationFile, input, 0644)
	if err != nil {
		return err
	}
	return nil
}

// GenAppStateFromConfig gets the genesis app state from the config
func GenAppStateFromConfig(cdc codec.JSONCodec, txEncodingConfig client.TxEncodingConfig,
	config *tmconfig.Config, initCfg genutiltypes.InitConfig, genesis *genutiltypes.AppGenesis, genBalIterator genutiltypes.GenesisBalancesIterator,
	validator genutiltypes.MessageValidator, valAddrCodec runtime.ValidatorAddressCodec,
	genFile string,
) (appState json.RawMessage, err error) {
	// process genesis transactions, else create default genesis.json
	appGenTxs, persistentPeers, err := genutil.CollectTxs(
		cdc, txEncodingConfig.TxJSONDecoder(), config.Moniker, initCfg.GenTxsDir, genesis, genBalIterator, validator, valAddrCodec)
	if err != nil {
		return appState, err
	}

	config.P2P.PersistentPeers = persistentPeers
	tmconfig.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

	// if there are no gen txs to be processed, return the default empty state
	if len(appGenTxs) == 0 {
		return appState, errors.New("there must be at least one genesis tx")
	}

	// create the app state
	appGenesisState, err := genutiltypes.GenesisStateFromAppGenesis(genesis)
	if err != nil {
		return appState, err
	}

	appGenesisState, err = genutil.SetGenTxsInAppGenesisState(cdc, txEncodingConfig.TxJSONEncoder(), appGenesisState, appGenTxs)
	if err != nil {
		return appState, err
	}

	appState, err = json.MarshalIndent(appGenesisState, "", "  ")
	if err != nil {
		return appState, err
	}

	genesis.AppState = appState
	err = genutil.ExportGenesisFile(genesis, genFile)

	return appState, err
}

func createMockAppCreator(wasmVmMeta memc.IWasmVmMeta, defaultNodeHome string) (*mcfg.MultiChainApp, func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp) {
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	appOpts := multichain.DefaultAppOptions{}
	appOpts.Set(flags.FlagHome, defaultNodeHome)
	// we set this so it does not try to read a genesis file
	appOpts.Set(flags.FlagChainID, mcfg.MYTHOS_CHAIN_ID_TESTNET)
	appOpts.Set(sdkserver.FlagInvCheckPeriod, 0)
	appOpts.Set(sdkserver.FlagUnsafeSkipUpgrades, 0)
	appOpts.Set(sdkserver.FlagMinGasPrices, "")
	appOpts.Set(sdkserver.FlagPruning, pruningtypes.PruningOptionDefault)
	g, goctx, _ := multichain.GetTestCtx(logger, true)
	return app.NewAppCreator(wasmVmMeta, logger, db, nil, appOpts, g, goctx, &multichain.MockApiCtx{})
}
