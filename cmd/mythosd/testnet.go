package main

// DONTCOVER

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	tmconfig "github.com/cometbft/cometbft/config"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cometbft/cometbft/types"
	tmtime "github.com/cometbft/cometbft/types/time"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p/core/crypto"

	pvm "github.com/cometbft/cometbft/privval"

	app "mythos/v1/app"
	mcfg "mythos/v1/config"
	config "mythos/v1/server/config"
	websrvconfig "mythos/v1/x/websrv/server/config"
	websrvflags "mythos/v1/x/websrv/server/flags"

	jsonrpcconfig "mythos/v1/x/wasmx/server/config"
	jsonrpcflags "mythos/v1/x/wasmx/server/flags"

	cosmosmodtypes "mythos/v1/x/cosmosmod/types"
	// "mythos/v1/testutil/network"
)

var (
	flagNodeDirPrefix     = "node-dir-prefix"
	flagNumValidators     = "v"
	flagOutputDir         = "output-dir"
	flagNodeDaemonHome    = "node-daemon-home"
	flagStartingIPAddress = "starting-ip-address"
	flagSameMachine       = "same-machine"
	flagNoCors            = "nocors"
	flagEnableLogging     = "enable-logging"
	flagRPCAddress        = "rpc.address"
	flagAPIAddress        = "api.address"
	flagPrintMnemonic     = "print-mnemonic"
	flagP2P               = "libp2p"
)

type initArgs struct {
	algo              string
	chainID           string
	keyringBackend    string
	minGasPrices      string
	nodeDaemonHome    string
	nodeDirPrefix     string
	numValidators     int
	outputDir         string
	startingIPAddress string
	sameMachine       bool
	noCors            bool
	p2p               bool
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
}

// NewTestnetCmd creates a root testnet command with subcommands to run an in-process testnet or initialize
// validator configuration files for running a multi-validator testnet in a separate process
func NewTestnetCmd(mbm module.BasicManager, genBalIterator cosmosmodtypes.GenesisBalancesIterator) *cobra.Command {
	testnetCmd := &cobra.Command{
		Use:                        "testnet",
		Short:                      "subcommands for starting or configuring local testnets",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	testnetCmd.AddCommand(testnetStartCmd())
	testnetCmd.AddCommand(testnetInitFilesCmd(mbm, genBalIterator))
	testnetCmd.AddCommand(testnetAddNodeCmd(mbm, genBalIterator))

	return testnetCmd
}

// get cmd to initialize all files for tendermint testnet and application
func testnetInitFilesCmd(mbm module.BasicManager, genBalIterator cosmosmodtypes.GenesisBalancesIterator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init-files",
		Short: "Initialize config directories & files for a multi-validator testnet running locally via separate processes (e.g. Docker Compose or similar)",
		Long: `init-files will setup "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.) for running "v" validator nodes.

Booting up a network with these validator folders is intended to be used with Docker Compose,
or a similar setup where each node has a manually configurable IP address.

Note, strict routability for addresses is turned off in the config file.

Example:
	mythosd testnet init-files --v 4 --output-dir ./.testnets --starting-ip-address 192.168.10.2
	`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			serverCtx := sdkserver.GetServerContextFromCmd(cmd)

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

			return initTestnetFiles(clientCtx, cmd, serverCtx.Config, mbm, genBalIterator, clientCtx.TxConfig.SigningContext().ValidatorAddressCodec(), args)
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

func testnetAddNodeCmd(mbm module.BasicManager, genBalIterator cosmosmodtypes.GenesisBalancesIterator) *cobra.Command {
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

			serverCtx := sdkserver.GetServerContextFromCmd(cmd)

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

			nodeIndex, err := strconv.Atoi(args_[0])
			if err != nil {
				panic(err)
			}
			leaderURI := args_[1]

			return testnetAddNode(clientCtx, cmd, serverCtx.Config, mbm, genBalIterator, clientCtx.TxConfig.SigningContext().ValidatorAddressCodec(), args, nodeIndex, leaderURI)
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

// get cmd to start multi validator in-process testnet
func testnetStartCmd() *cobra.Command {
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

			return startTestnet(cmd, args)
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

const nodeDirPerm = 0o755

// const p2pListenAddress = "26656"
// const p2pListenAddressMulti = 26756
// replace with network module // TODO maybe have network module on 26656
const p2pListenAddress = "8090"
const p2pListenAddressMulti = 8090

// initTestnetFiles initializes testnet files for a testnet to be run in a separate process
func initTestnetFiles(
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *tmconfig.Config,
	mbm module.BasicManager,
	genBalIterator cosmosmodtypes.GenesisBalancesIterator,
	valAddrCodec runtime.ValidatorAddressCodec,
	args initArgs,
) error {
	return initTestnetFilesInternal(clientCtx, cmd, nodeConfig, mbm, genBalIterator, valAddrCodec, args, 0, "")
}

func testnetAddNode(
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *tmconfig.Config,
	mbm module.BasicManager,
	genBalIterator cosmosmodtypes.GenesisBalancesIterator,
	valAddrCodec runtime.ValidatorAddressCodec,
	args initArgs,
	nodeIndex int,
	leaderURI string,
) error {
	args.numValidators = nodeIndex + 1
	args.sameMachine = true
	err := initTestnetFilesInternal(clientCtx, cmd, nodeConfig, mbm, genBalIterator, valAddrCodec, args, nodeIndex, leaderURI)
	if err != nil {
		return err
	}
	// copy genesis from node0 into our node
	nodeDirName := fmt.Sprintf("%s%d", args.nodeDirPrefix, 0)
	nodeDir := filepath.Join(args.outputDir, nodeDirName, args.nodeDaemonHome)
	nodeConfig.SetRoot(nodeDir)
	genesisFile0 := nodeConfig.GenesisFile()

	nodeDirName = fmt.Sprintf("%s%d", args.nodeDirPrefix, nodeIndex)
	nodeDir = filepath.Join(args.outputDir, nodeDirName, args.nodeDaemonHome)
	nodeConfig.SetRoot(nodeDir)
	genesisFileNew := nodeConfig.GenesisFile()

	bz, err := os.ReadFile(genesisFile0)
	if err != nil {
		return err
	}
	err = os.WriteFile(genesisFileNew, bz, 0o644)
	return err
}

// initTestnetFiles initializes testnet files for a testnet to be run in a separate process
func initTestnetFilesInternal(
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *tmconfig.Config,
	mbm module.BasicManager,
	genBalIterator cosmosmodtypes.GenesisBalancesIterator,
	valAddrCodec runtime.ValidatorAddressCodec,
	args initArgs,
	nodeIndexStart int,
	leaderURI string,
) error {
	if args.chainID == "" {
		args.chainID = fmt.Sprintf("mythos_%d-1", tmrand.Int63n(9999999999999)+1)
	}
	var err error
	nodeIDs := make([]string, args.numValidators)
	valPubKeys := make([]cryptotypes.PubKey, args.numValidators)
	nodeIPs := make([]string, args.numValidators)

	appConfig := config.DefaultConfig()
	appConfig.MinGasPrices = args.minGasPrices
	appConfig.API.Enable = true
	appConfig.Telemetry.Enabled = true
	appConfig.Telemetry.PrometheusRetentionTime = 60
	appConfig.Telemetry.EnableHostnameLabel = false
	appConfig.Telemetry.GlobalLabels = [][]string{{"chain_id", args.chainID}}
	if args.noCors {
		appConfig.API.EnableUnsafeCORS = true
	}

	genAccounts := make([]authtypes.GenesisAccount, args.numValidators)
	genBalances := make([]banktypes.Balance, args.numValidators)
	genFiles := make([]string, args.numValidators)

	networkIps := make([]string, args.numValidators)
	valaddr := make([]sdk.AccAddress, args.numValidators)
	nodeDirNames := make([]string, args.numValidators)
	nodeDirs := make([]string, args.numValidators)
	kbs := make([]keyring.Keyring, args.numValidators)
	inBuf := bufio.NewReader(cmd.InOrStdin())
	// generate private keys, node IDs, and initial transactions
	for i := nodeIndexStart; i < args.numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", args.nodeDirPrefix, i)
		nodeDir := filepath.Join(args.outputDir, nodeDirName, args.nodeDaemonHome)

		nodeDirNames[i] = nodeDirName
		nodeDirs[i] = nodeDir

		nodeConfig.SetRoot(nodeDir)
		nodeConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
		// nodeConfig.ProxyApp = "tcp://127.0.0.1:26657"
		if args.sameMachine {
			nodeConfig.RPC.ListenAddress = "tcp://0.0.0.0:" + strconv.Itoa(26657+i)
			// nodeConfig.ProxyApp = "tcp://127.0.0.1:" + strconv.Itoa(26657+i)
		}
		if args.noCors {
			nodeConfig.RPC.CORSAllowedOrigins = []string{"*"}
		}

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

		host := fmt.Sprintf("%s@%s:%s", addr.String(), nodeIPs[i], "8090")
		if args.sameMachine && !args.p2p {
			ipaddr := strings.Replace(appConfig.Network.Address, "8090", strconv.Itoa(8090+i), 1)
			host = fmt.Sprintf("%s@%s", addr.String(), ipaddr)
		}

		if args.p2p {
			valStr := addr.String()

			privValid := pvm.LoadOrGenFilePV(nodeConfig.PrivValidatorKeyFile(), nodeConfig.PrivValidatorStateFile())
			pk := privValid.Key.PrivKey.Bytes()
			pkcrypto, err := crypto.UnmarshalEd25519PrivateKey(pk)
			if err != nil {
				return err
			}
			identity := libp2p.Identity(pkcrypto)
			p2pPort := "5001"

			if args.sameMachine {
				ipaddr := strings.Replace(appConfig.Network.Address, "8090", strconv.Itoa(5001+i), 1)
				ipaddr = strings.Replace(ipaddr, "0.0.0.0", "127.0.0.1", 1)
				parts := strings.Split(ipaddr, ":")
				node, err := libp2p.New(
					libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%s", parts[0], parts[1])),
					libp2p.Ping(false),
					identity,
				)
				if err != nil {
					return err
				}
				p2pid := node.ID()
				host = fmt.Sprintf("%s@/ip4/%s/tcp/%s/p2p/%s", valStr, parts[0], parts[1], p2pid)
			} else {
				ipaddr := nodeIPs[i]
				ipaddr = strings.Replace(ipaddr, "0.0.0.0", "127.0.0.1", 1)
				node, err := libp2p.New(
					libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%s", ipaddr, p2pPort)),
					libp2p.Ping(false),
					identity,
				)
				if err != nil {
					return err
				}
				p2pid := node.ID()
				host = fmt.Sprintf("%s@/ip4/%s/tcp/%s/p2p/%s", valStr, nodeIPs[i], p2pPort, p2pid)
			}
		}

		valaddr[i] = addr
		networkIps[i] = host
	}
	networkIpsStr := strings.Join(networkIps, ",")
	networkIpsStr = strings.Trim(networkIpsStr, ",")
	if leaderURI != "" {
		networkIpsStr = networkIpsStr + "," + leaderURI
	}

	for i := nodeIndexStart; i < args.numValidators; i++ {
		gentxsDir := filepath.Join(args.outputDir, "gentxs")
		nodeDirName := nodeDirNames[i]
		nodeDir := nodeDirs[i]
		addr := valaddr[i]
		kb := kbs[i]
		memo := fmt.Sprintf("%s@%s:%s", nodeIDs[i], nodeIPs[i], p2pListenAddress)
		if args.sameMachine {
			memo = fmt.Sprintf("%s@%s:%s", nodeIDs[i], "0.0.0.0", strconv.Itoa(p2pListenAddressMulti+i))
		}

		accStakingTokens := sdk.TokensFromConsensusPower(5000, mcfg.PowerReduction)
		coins := sdk.Coins{
			sdk.NewCoin(mcfg.BaseDenom, accStakingTokens),
		}

		genBalances[i] = banktypes.Balance{Address: addr.String(), Coins: coins.Sort()}
		genAccounts[i] = authtypes.NewBaseAccount(addr, nil, 0, 0)

		valStr, err := valAddrCodec.BytesToString(sdk.ValAddress(addr))
		if err != nil {
			return err
		}
		valTokens := sdk.TokensFromConsensusPower(100, mcfg.PowerReduction)
		createValMsg, err := stakingtypes.NewMsgCreateValidator(
			valStr,
			valPubKeys[i],
			sdk.NewCoin(mcfg.BaseDenom, valTokens),
			stakingtypes.NewDescription(nodeDirName, "", "", "", ""),
			stakingtypes.NewCommissionRates(math.LegacyOneDec(), math.LegacyOneDec(), math.LegacyOneDec()),
			math.OneInt(),
		)
		if err != nil {
			return err
		}

		txBuilder := clientCtx.TxConfig.NewTxBuilder()
		if err := txBuilder.SetMsgs(createValMsg); err != nil {
			return err
		}

		txBuilder.SetMemo(memo)

		txFactory := tx.Factory{}
		txFactory = txFactory.
			WithChainID(args.chainID).
			WithMemo(memo).
			WithKeybase(kb).
			WithTxConfig(clientCtx.TxConfig)

		if err := tx.Sign(cmd.Context(), txFactory, nodeDirName, txBuilder, true); err != nil {
			return err
		}

		txBz, err := clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
		if err != nil {
			return err
		}

		if err := writeFile(fmt.Sprintf("%v.json", nodeDirName), gentxsDir, txBz); err != nil {
			return err
		}

		customAppTemplate, customAppConfig := config.AppConfig()
		srvconfig.SetConfigTemplate(customAppTemplate)

		if err := sdkserver.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, tmconfig.DefaultConfig()); err != nil {
			return err
		}

		// value copy, not reference!
		appConfigCopy := *appConfig
		if args.sameMachine {
			appConfigCopy.API.Address = strings.Replace(appConfig.API.Address, "1317", strconv.Itoa(1317+i), 1)
			appConfigCopy.GRPC.Address = strings.Replace(appConfig.GRPC.Address, "9090", strconv.Itoa(9090+i), 1)
			appConfigCopy.JsonRpc.Address = strings.Replace(appConfig.JsonRpc.Address, "8545", strconv.Itoa(8555+i*2), 1)
			appConfigCopy.JsonRpc.WsAddress = strings.Replace(appConfig.JsonRpc.WsAddress, "8546", strconv.Itoa(8556+i), 1)
			appConfigCopy.Websrv.Address = strings.Replace(appConfig.Websrv.Address, "9999", strconv.Itoa(9900+i), 1)
			appConfigCopy.Network.Address = strings.Replace(appConfig.Network.Address, "8090", strconv.Itoa(8090+i), 1)
		}
		appConfigCopy.Network.Id = int32(i - nodeIndexStart)
		appConfigCopy.Network.Ips = networkIpsStr

		srvconfig.WriteConfigFile(filepath.Join(nodeDir, "config/app.toml"), appConfigCopy)
	}

	if nodeIndexStart == 0 {
		if err := initGenFiles(clientCtx, mbm, args.chainID, genAccounts, genBalances, genFiles, args.numValidators); err != nil {
			return err
		}
		err = collectGenFiles(
			clientCtx, nodeConfig, args.chainID, nodeIDs, valPubKeys, args.numValidators,
			args.outputDir, args.nodeDirPrefix, args.nodeDaemonHome, genBalIterator, valAddrCodec, args.sameMachine, nodeIndexStart,
		)
		if err != nil {
			return err
		}
	}

	cmd.PrintErrf("Successfully initialized %d node directories\n", args.numValidators)
	return nil
}

func initGenFiles(
	clientCtx client.Context,
	mbm module.BasicManager,
	chainID string,
	genAccounts []authtypes.GenesisAccount,
	genBalances []banktypes.Balance,
	genFiles []string,
	numValidators int,
) error {
	appGenState := mbm.DefaultGenesis(clientCtx.Codec)

	var cosmosmodGenState cosmosmodtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[cosmosmodtypes.ModuleName], &cosmosmodGenState)

	cosmosmodGenState.Bank.Balances = genBalances
	cosmosmodGenState.Staking.Params.BondDenom = mcfg.BondBaseDenom
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

	// var wasmxGenState wasmxtypes.GenesisState
	// clientCtx.Codec.MustUnmarshalJSON(appGenState[wasmxtypes.ModuleName], &wasmxGenState)

	// wasmxGenState.Params.Denom = coinDenom
	// appGenState[wasmxtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&wasmxGenState)

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

func collectGenFiles(
	clientCtx client.Context, nodeConfig *tmconfig.Config, chainID string,
	nodeIDs []string, valPubKeys []cryptotypes.PubKey, numValidators int,
	outputDir, nodeDirPrefix, nodeDaemonHome string, genBalIterator cosmosmodtypes.GenesisBalancesIterator, valAddrCodec runtime.ValidatorAddressCodec, sameMachine bool, nodeIndexStart int,
) error {
	var appState json.RawMessage
	genTime := tmtime.Now()

	for i := nodeIndexStart; i < numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)
		gentxsDir := filepath.Join(outputDir, "gentxs")
		nodeConfig.Moniker = nodeDirName
		nodeConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
		// nodeConfig.ProxyApp = "tcp://127.0.0.1:26657"
		if sameMachine {
			nodeConfig.RPC.ListenAddress = "tcp://0.0.0.0:" + strconv.Itoa(26657+i)
			// nodeConfig.ProxyApp = "tcp://127.0.0.1:" + strconv.Itoa(26657+i)
		}

		nodeConfig.SetRoot(nodeDir)

		nodeID, valPubKey := nodeIDs[i], valPubKeys[i]
		initCfg := genutiltypes.NewInitConfig(chainID, gentxsDir, nodeID, valPubKey)

		appGenesis, err := genutiltypes.AppGenesisFromFile(nodeConfig.GenesisFile())
		if err != nil {
			return err
		}

		nodeAppState, err := genutil.GenAppStateFromConfig(clientCtx.Codec, clientCtx.TxConfig, nodeConfig, initCfg, appGenesis, genBalIterator, genutiltypes.DefaultMessageValidator, valAddrCodec)
		if err != nil {
			return err
		}

		if appState == nil {
			// set the canonical application state (they should not differ)
			appState = nodeAppState
		}

		genFile := nodeConfig.GenesisFile()

		// overwrite each validator's genesis file to have a canonical genesis time
		if err := genutil.ExportGenesisFileWithTime(genFile, chainID, nil, appState, genTime); err != nil {
			return err
		}
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
func startTestnet(cmd *cobra.Command, args startArgs) error {
	networkConfig := network.DefaultConfig(app.NewTestNetworkFixture)

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
