package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	abciserver "github.com/cometbft/cometbft/abci/server"
	tcmd "github.com/cometbft/cometbft/cmd/cometbft/commands"
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	cmttypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"

	pruningtypes "cosmossdk.io/store/pruning/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	sdkserverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servergrpc "github.com/cosmos/cosmos-sdk/server/grpc"
	servercmtlog "github.com/cosmos/cosmos-sdk/server/log"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	srvconfig "github.com/loredanacirstea/wasmx/server/config"
	srvflags "github.com/loredanacirstea/wasmx/server/flags"
	websrv "github.com/loredanacirstea/wasmx/x/websrv/server"
	websrvconfig "github.com/loredanacirstea/wasmx/x/websrv/server/config"
	websrvflags "github.com/loredanacirstea/wasmx/x/websrv/server/flags"

	jsonrpc "github.com/loredanacirstea/wasmx/x/wasmx/server"
	jsonrpcconfig "github.com/loredanacirstea/wasmx/x/wasmx/server/config"
	jsonrpcflags "github.com/loredanacirstea/wasmx/x/wasmx/server/flags"

	networkgrpc "github.com/loredanacirstea/wasmx/x/network/keeper"
	networkserver "github.com/loredanacirstea/wasmx/x/network/server"
	networkconfig "github.com/loredanacirstea/wasmx/x/network/server/config"
	networkflags "github.com/loredanacirstea/wasmx/x/network/server/flags"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	"github.com/loredanacirstea/wasmx/x/network/vmp2p"

	mapp "github.com/loredanacirstea/wasmx/app"
	mcfg "github.com/loredanacirstea/wasmx/config"
	mctx "github.com/loredanacirstea/wasmx/context"
	menc "github.com/loredanacirstea/wasmx/encoding"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

var flagSameMachineNodeIndex = "same-machine-node-index"

// StartCmd runs the service passed in, either stand-alone or in-process with
// Tendermint.
func StartCmd(wasmVmMeta memc.IWasmVmMeta, appCreator servertypes.AppCreator, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the full node",
		Long: `Run the full node application with Tendermint in or out of process. By
default, the application will run with Tendermint in process.

Pruning options can be provided via the '--pruning' flag or alternatively with '--pruning-keep-recent',
'pruning-keep-every', and 'pruning-interval' together.

For '--pruning' the options are as follows:

default: the last 100 states are kept in addition to every 500th state; pruning at 10 block intervals
nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
everything: all saved states will be deleted, storing only the current and previous state; pruning at 10 block intervals
custom: allow pruning options to be manually specified through 'pruning-keep-recent', 'pruning-keep-every', and 'pruning-interval'

Node halting configurations exist in the form of two flags: '--halt-height' and '--halt-time'. During
the ABCI Commit phase, the node will check if the current block height is greater than or equal to
the halt-height or if the current block time is greater than or equal to the halt-time. If so, the
node will attempt to gracefully shutdown and the block will not be committed. In addition, the node
will not be able to commit subsequent blocks.

For profiling and benchmarking purposes, CPU profiling can be enabled via the '--cpu-profile' flag
which accepts a path for the resulting pprof file.

The node may be started in a 'query only' mode where only the gRPC and JSON HTTP
API services are enabled via the 'grpc-only' flag. In this mode, Tendermint is
bypassed and can be used when legacy queries are needed after an on-chain upgrade
is performed. Note, when enabled, gRPC will also be automatically enabled.
`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			serverCtx := GetServerContextFromCmd(cmd)

			// Bind flags to the Context's Viper so the app construction can set
			// options accordingly.
			err := serverCtx.Viper.BindPFlags(cmd.Flags())
			if err != nil {
				return err
			}

			_, err = server.GetPruningOptionsFromFlags(serverCtx.Viper)
			return err
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			serverCtx := GetServerContextFromCmd(cmd)
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			withTM, _ := cmd.Flags().GetBool(srvflags.WithTendermint)
			if !withTM {
				serverCtx.Logger.Info("starting ABCI without Tendermint")
				// return startStandAlone(serverCtx, appCreator)
				// TODO replace this
				return nil
			}

			serverCtx.Logger.Info("Unlocking keyring")

			// fire unlock precess for keyring
			keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
			if keyringBackend == keyring.BackendFile {
				_, err = clientCtx.Keyring.List()
				if err != nil {
					return err
				}
			}

			serverCtx.Logger.Info("starting ABCI with WasmX Tendermint")

			// amino is needed here for backwards compatibility of REST routes
			return startInProcess(wasmVmMeta, serverCtx, clientCtx, appCreator)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().Bool(srvflags.WithTendermint, true, "Run abci app embedded in-process with tendermint")
	cmd.Flags().String(srvflags.Address, "tcp://0.0.0.0:26658", "Listen address")
	cmd.Flags().String(srvflags.Transport, "socket", "Transport protocol: socket, grpc")
	cmd.Flags().String(srvflags.TraceStore, "", "Enable KVStore tracing to an output file")
	cmd.Flags().String(server.FlagMinGasPrices, "", "Minimum gas prices to accept for transactions; Any fee in a tx must meet this minimum (e.g. 0.01photon;0.0001stake)") //nolint:lll
	cmd.Flags().IntSlice(server.FlagUnsafeSkipUpgrades, []int{}, "Skip a set of upgrade heights to continue the old binary")
	cmd.Flags().Uint64(server.FlagHaltHeight, 0, "Block height at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Uint64(server.FlagHaltTime, 0, "Minimum block time (in Unix seconds) at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Bool(server.FlagInterBlockCache, true, "Enable inter-block caching")
	cmd.Flags().String(srvflags.CPUProfile, "", "Enable CPU profiling and write to the provided file")
	cmd.Flags().Bool(server.FlagTrace, false, "Provide full stack traces for errors in ABCI Log")
	cmd.Flags().String(server.FlagPruning, pruningtypes.PruningOptionDefault, "Pruning strategy (default|nothing|everything|custom)")
	cmd.Flags().Uint64(server.FlagPruningKeepRecent, 0, "Number of recent heights to keep on disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(server.FlagPruningInterval, 0, "Height interval at which pruned heights are removed from disk (ignored if pruning is not 'custom')") //nolint:lll
	cmd.Flags().Uint(server.FlagInvCheckPeriod, 0, "Assert registered invariants every N blocks")
	cmd.Flags().Uint64(server.FlagMinRetainBlocks, 0, "Minimum block height offset during ABCI commit to prune Tendermint blocks")
	cmd.Flags().String(srvflags.AppDBBackend, "", "The type of database for application and snapshots databases")

	cmd.Flags().Bool(srvflags.GRPCOnly, false, "Start the node in gRPC query only mode without Tendermint process")
	cmd.Flags().Bool(srvflags.GRPCEnable, true, "Define if the gRPC server should be enabled")
	cmd.Flags().String(srvflags.GRPCAddress, sdkserverconfig.DefaultGRPCAddress, "the gRPC server address to listen on")
	cmd.Flags().Bool(srvflags.GRPCWebEnable, true, "Define if the gRPC-Web server should be enabled. (Note: gRPC must also be enabled.)")

	cmd.Flags().Bool(srvflags.RPCEnable, false, "Defines if Cosmos-sdk REST server should be enabled")
	cmd.Flags().Bool(srvflags.EnabledUnsafeCors, false, "Defines if CORS should be enabled (unsafe - use it at your own risk)")

	cmd.Flags().Bool(websrvflags.WebsrvEnable, true, "Define if the websrv web server should be enabled")
	cmd.Flags().Bool(websrvflags.WebsrvEnableOAuth, true, "Define if the websrv oauth server should be enabled. (Note: websrv.enable must also be enabled.)")
	cmd.Flags().String(websrvflags.WebsrvAddress, websrvconfig.DefaultWebsrvAddress, "the Websrv web server address to listen on")
	cmd.Flags().Int(websrvflags.WebsrvMaxOpenConnections, websrvconfig.DefaultMaxOpenConnections, "Sets the maximum number of simultaneous connections for the server listener") //nolint:lll
	cmd.Flags().StringSlice(websrvflags.WebsrvCORSAllowedOrigins, websrvconfig.DefaultCORSAllowedOrigins, "Sets the allowed origins for the server listener")
	cmd.Flags().StringSlice(websrvflags.WebsrvCORSAllowedMethods, websrvconfig.DefaultCORSAllowedMethods, "Sets the allowed methods for the server listener")
	cmd.Flags().StringSlice(websrvflags.WebsrvCORSAllowedHeaders, websrvconfig.DefaultCORSAllowedHeaders, "Sets the allowed headers for the server listener")

	cmd.Flags().Bool(jsonrpcflags.JsonRpcEnable, true, "Define if the json-rpc server should be enabled")
	cmd.Flags().StringSlice(jsonrpcflags.JsonRpcApi, jsonrpcconfig.GetDefaultAPINamespaces(), "Defines a list of JSON-RPC namespaces that should be enabled")
	cmd.Flags().String(jsonrpcflags.JsonRpcAddress, jsonrpcconfig.DefaultJsonRpcAddress, "the json-rpc server address to listen on")
	cmd.Flags().String(jsonrpcflags.JsonRpcWsAddress, jsonrpcconfig.DefaultJsonRpcWsAddress, "the json-rpc websocket server address to listen on")
	cmd.Flags().Duration(jsonrpcflags.JsonRpcEVMTimeout, jsonrpcconfig.DefaultEVMTimeout, "Sets a timeout used for eth_call (0=infinite)")
	cmd.Flags().Duration(jsonrpcflags.JsonRpcHTTPTimeout, jsonrpcconfig.DefaultHTTPTimeout, "Sets a read/write timeout for json-rpc http server (0=infinite)")
	cmd.Flags().Duration(jsonrpcflags.JsonRpcHTTPIdleTimeout, jsonrpcconfig.DefaultHTTPIdleTimeout, "Sets a idle timeout for json-rpc http server (0=infinite)")
	cmd.Flags().Bool(jsonrpcflags.JsonRpcAllowUnprotectedTxs, jsonrpcconfig.DefaultAllowUnprotectedTxs, "Allow for unprotected (non EIP155 signed) transactions to be submitted via the node's RPC when the global parameter is disabled")
	cmd.Flags().Int(jsonrpcflags.JsonRpcMaxOpenConnections, jsonrpcconfig.DefaultMaxOpenConnections, "Sets the maximum number of simultaneous connections for the server listener")

	cmd.Flags().Bool(networkflags.NetworkEnable, true, "Define if the network grpc server should be enabled")
	cmd.Flags().Bool(networkflags.NetworkLeader, false, "Set node as leader. Temporary.")
	cmd.Flags().String(networkflags.NetworkIps, "localhost:8090", "Set node ips. Temporary.")
	cmd.Flags().String(networkflags.NetworkNodeId, "0", "This node's index in the array of validators")
	cmd.Flags().String(networkflags.NetworkAddress, networkconfig.DefaultNetworkAddress, "the network grpc server address to listen on")
	cmd.Flags().Int(networkflags.NetworkMaxOpenConnections, networkconfig.DefaultMaxOpenConnections, "Sets the maximum number of simultaneous connections for the server listener") //nolint:lll

	cmd.Flags().String(srvflags.TLSCertPath, "", "the cert.pem file path for the server TLS configuration")
	cmd.Flags().String(srvflags.TLSKeyPath, "", "the key.pem file path for the server TLS configuration")

	cmd.Flags().Uint64(server.FlagStateSyncSnapshotInterval, 0, "State sync snapshot interval")
	cmd.Flags().Uint32(server.FlagStateSyncSnapshotKeepRecent, 2, "State sync snapshot to keep")

	cmd.Flags().Uint32(flagSameMachineNodeIndex, 0, "delta for assigning port numbers; usually equal to the number of nodes ran on the machine")

	// add support for all Tendermint-specific command line options
	tcmd.AddNodeFlags(cmd)
	return cmd
}

func startStandAlone(wasmVmMeta memc.IWasmVmMeta, svrCtx *server.Context, _ servertypes.AppCreator) error {
	addr := svrCtx.Viper.GetString(srvflags.Address)
	transport := svrCtx.Viper.GetString(srvflags.Transport)
	home := svrCtx.Viper.GetString(flags.FlagHome)

	g, ctx := getCtx(svrCtx, true)

	db, err := openDB(home, server.GetAppDBBackend(svrCtx.Viper))
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			svrCtx.Logger.With("error", err).Error("error closing db")
		}
	}()

	traceWriterFile := svrCtx.Viper.GetString(srvflags.TraceStore)
	traceWriter, err := openTraceWriter(traceWriterFile)
	if err != nil {
		return err
	}

	apictx := &APICtx{
		GoRoutineGroup:  g,
		GoContextParent: ctx,
		SvrCtx:          svrCtx,
		ClientCtx:       client.Context{},
	}

	bapps, appCreator := mapp.NewAppCreator(
		wasmVmMeta,
		svrCtx.Logger,
		db,
		traceWriter,
		svrCtx.Viper,
		g, ctx,
		apictx,
	)
	apictx.AppCreator = appCreator
	apictx.Multiapp = bapps

	config, err := srvconfig.GetConfig(svrCtx.Viper)
	if err != nil {
		svrCtx.Logger.Error("failed to get server config", "error", err.Error())
		return err
	}

	if err := config.ValidateBasic(); err != nil {
		return err
	}

	for _, chainId := range config.Network.InitialChains {
		chainCfg, err := mcfg.GetChainConfig(chainId)
		if err != nil {
			panic(err)
		}

		appCreator(chainId, chainCfg)
	}

	iapp_, err := bapps.GetApp(cast.ToString(svrCtx.Viper.Get(flags.FlagChainID)))
	if err != nil {
		return err
	}
	app_, ok := iapp_.(*mapp.App)
	if !ok {
		return fmt.Errorf("error App interface from multichainapp")
	}
	app := servertypes.Application(app_)

	cmtApp := server.NewCometABCIWrapper(app)
	svr, err := abciserver.NewServer(addr, transport, cmtApp)
	if err != nil {
		return fmt.Errorf("error creating listener: %v", err)
	}

	svr.SetLogger(servercmtlog.CometLoggerWrapper{Logger: svrCtx.Logger.With(log.ModuleKey, "abci-server")})

	g.Go(func() error {
		if err := svr.Start(); err != nil {
			svrCtx.Logger.Error("failed to start out-of-process ABCI server", "err", err)
			return err
		}

		// Wait for the calling process to be canceled or close the provided context,
		// so we can gracefully stop the ABCI server.
		<-ctx.Done()
		svrCtx.Logger.Info("stopping the ABCI server...")
		return svr.Stop()
	})

	return g.Wait()
}

// legacyAminoCdc is used for the legacy REST API
func startInProcess(wasmVmMeta memc.IWasmVmMeta, svrCtx *server.Context, clientCtx client.Context, _ servertypes.AppCreator) (err error) {
	tndcfg := svrCtx.Config
	home := tndcfg.RootDir
	logger := svrCtx.Logger
	nodeOffset := svrCtx.Viper.GetUint32(flagSameMachineNodeIndex)

	g, ctx := getCtx(svrCtx, true)

	if cpuProfile := svrCtx.Viper.GetString(srvflags.CPUProfile); cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			return err
		}
		memProfile := cpuProfile + "_mem.prof"
		fm, err := os.Create(memProfile)
		if err != nil {
			return err
		}
		gorProfile := cpuProfile + "_gor.prof"
		fg, err := os.Create(gorProfile)
		if err != nil {
			return err
		}
		allocProfile := cpuProfile + "_alloc.prof"
		fa, err := os.Create(allocProfile)
		if err != nil {
			return err
		}

		logger.Info("starting CPU profiler", "profile", cpuProfile)
		if err := pprof.StartCPUProfile(f); err != nil {
			return err
		}
		if err := pprof.WriteHeapProfile(fm); err != nil {
			return err
		}
		if err := pprof.Lookup("goroutine").WriteTo(fg, 0); err != nil {
			return err
		}
		if err := pprof.Lookup("allocs").WriteTo(fa, 0); err != nil {
			return err
		}

		// "goroutine":    goroutineProfile,
		// "threadcreate": threadcreateProfile,
		// "heap":         heapProfile,
		// "allocs":       allocsProfile,
		// "block":        blockProfile,
		// "mutex":        mutexProfile,

		defer func() {
			logger.Info("stopping CPU profiler", "profile", cpuProfile)
			pprof.StopCPUProfile()
			if err := f.Close(); err != nil {
				logger.Error("failed to close CPU profiler file", "error", err.Error())
			}
		}()
		defer func() {
			logger.Info("stopping memory profiler", "profile", memProfile)
			runtime.GC() // get up-to-date statistics
			if err := fm.Close(); err != nil {
				logger.Error("failed to close memory profiler file", "error", err.Error())
			}
		}()
		defer func() {
			logger.Info("stopping goroutine profiler", "profile", gorProfile)
			if err := fg.Close(); err != nil {
				logger.Error("failed to close goroutine profiler file", "error", err.Error())
			}
		}()
		defer func() {
			logger.Info("stopping allocs profiler", "profile", allocProfile)
			if err := fa.Close(); err != nil {
				logger.Error("failed to close allocs profiler file", "error", err.Error())
			}
		}()
	}

	traceWriterFile := svrCtx.Viper.GetString(srvflags.TraceStore)
	db, err := openDB(home, server.GetAppDBBackend(svrCtx.Viper))
	if err != nil {
		logger.Error("failed to open DB", "error", err.Error())
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.With("error", err).Error("error closing db")
		}
	}()

	traceWriter, err := openTraceWriter(traceWriterFile)
	if err != nil {
		logger.Error("failed to open trace writer", "error", err.Error())
		return err
	}

	msrvconfig, err := srvconfig.GetConfig(svrCtx.Viper)
	if err != nil {
		logger.Error("failed to get server config", "error", err.Error())
		return err
	}

	if err := msrvconfig.ValidateBasic(); err != nil {
		return err
	}

	genDocProvider := getGenDocProvider(tndcfg)
	genDoc, err := genDocProvider()
	if err != nil {
		return err
	}
	nodeKey, err := p2p.LoadOrGenNodeKey(tndcfg.NodeKeyFile())
	if err != nil {
		return err
	}

	// var rpcClient client.CometRPC
	// var (
	// 	tmNode *node.Node
	// 	gRPCOnly  = svrCtx.Viper.GetBool(srvflags.GRPCOnly)
	// 	cleanupFn func()
	// )

	clientCtx = clientCtx.
		WithHomeDir(home).
		WithChainID(genDoc.ChainID)

	metrics, err := startTelemetry(msrvconfig.Config)
	if err != nil {
		return err
	}
	metricsProvider := node.DefaultMetricsProvider(tndcfg.Instrumentation)

	apictx := &APICtx{
		GoRoutineGroup:  g,
		GoContextParent: ctx,
		SvrCtx:          svrCtx,
		ClientCtx:       clientCtx,
		SrvCfg:          msrvconfig,
		TndCfg:          tndcfg,
		MetricsProvider: metricsProvider,
		Metrics:         metrics,
	}

	multiapp, appCreator := mapp.NewAppCreator(
		wasmVmMeta,
		logger,
		db,
		traceWriter,
		svrCtx.Viper,
		g, ctx,
		apictx,
	)
	apictx.AppCreator = appCreator
	apictx.Multiapp = multiapp

	// create apps for initial chains
	initialChains := msrvconfig.Network.InitialChains
	portOffset := int32(nodeOffset * uint32(len(initialChains)))
	chainsToStart := []string{}
	startStateSyncProviders := []func(){}
	for i, chainId := range initialChains {
		chainCfg, err := mcfg.GetChainConfig(chainId)
		if err != nil {
			panic(err)
		}
		// TODO use ports set in toml file; family of ports
		ports := mctx.GetChainNodePorts(int32(i), portOffset)
		// ports for subchains
		initialPorts := mctx.GetChainNodePortsInitial(int32(i), portOffset)

		// Run state sync
		// TODO for any multichain
		if tndcfg.StateSync.Enable {
			mythosapp, csvrCtx, _, cmsrvconfig, ctndcfg, rpcClient, err := apictx.BuildConfigs(chainId, chainCfg, ports)
			if err != nil {
				panic(err)
			}

			peers := networktypes.GetPeersFromConfigIps(chainId, cmsrvconfig.Network.Ips)
			currentNodeId, err := networktypes.GetCurrentNodeIdFromConfig(chainId, cmsrvconfig.Network.Id)
			if err != nil {
				panic(err)
			}
			if len(peers) >= 2 {
				mythosapp.NonDeterministicSetNodePortsInitial(initialPorts)

				privValidator := pvm.LoadOrGenFilePV(ctndcfg.PrivValidatorKeyFile(), ctndcfg.PrivValidatorStateFile())

				err = startStateSync(mythosapp.GetGoContextParent(), mythosapp.GetGoRoutineGroup(), csvrCtx, cmsrvconfig, ctndcfg, chainId, *chainCfg, mythosapp, rpcClient, privValidator.Key.PrivKey.Bytes(), peers, currentNodeId)
				if err != nil {
					panic(err)
				}
				// fixme
				continue
				// return g.Wait()
			} else {
				csvrCtx.Logger.Info("not starting statesync", "chain_id", chainId, "reason", "app.toml ips has < 2 ips: statesync provider must be the first ip in the list")
			}
		}

		chainsToStart = append(chainsToStart, chainId)

		mythosapp_, csvrCtx, _, cmsrvconfig, ctndcfg, rpcClient, err := apictx.StartChainApis(chainId, chainCfg, ports)
		if err != nil {
			panic(err)
		}
		mythosapp, ok := mythosapp_.(*mapp.App)
		if !ok {
			panic(fmt.Errorf("cannot convert MythosApp to App"))
		}
		app := servertypes.Application(mythosapp)
		bapp := mythosapp.GetBaseApp()
		privValidator := pvm.LoadOrGenFilePV(ctndcfg.PrivValidatorKeyFile(), ctndcfg.PrivValidatorStateFile())
		genesisDocProvider := getGenDocProvider2(ctndcfg)

		mythosapp.SetServerConfig(cmsrvconfig)
		mythosapp.SetTendermintConfig(ctndcfg)
		mythosapp.SetRpcClient(rpcClient)

		// initialize chain if this is block 0
		// init all chains first and start them afterwards
		// InitChain runs multiple contract executions that are not under ActionExecutor control; starting the chains while InitChain is not finished will start delayed executions that will intersect with InitChain executions
		if bapp.LastBlockHeight() == 0 {
			_, err := networkgrpc.InitChain(csvrCtx.Logger, cmsrvconfig, app, privValidator, nodeKey, genesisDocProvider, chainId, mythosapp.GetNetworkKeeper(), initialPorts)
			if err != nil {
				return err
			}
		}

		if !strings.Contains(chainId, "level0") {
			ssfn := func(mythosapp *mapp.App, csvrCtx *server.Context, cmsrvconfig *srvconfig.Config, ctndcfg *cmtcfg.Config, chainId string, rpcClient client.CometRPC, privValidator *pvm.FilePV) func() {
				return func() {
					startStateSyncProvider(mythosapp.GetGoContextParent(), mythosapp.GetGoRoutineGroup(), csvrCtx, cmsrvconfig, ctndcfg, chainId, *chainCfg, mythosapp, rpcClient, privValidator.Key.PrivKey.Bytes())
				}
			}(mythosapp, csvrCtx, cmsrvconfig, ctndcfg, chainId, rpcClient, privValidator)
			startStateSyncProviders = append(startStateSyncProviders, ssfn)
		}
	}

	// start nodes for all chains
	// should be all chains, taken from level0
	for _, chainId := range chainsToStart {
		iapp, _ := multiapp.GetApp(chainId)
		app, ok := iapp.(mcfg.MythosApp)
		if !ok {
			return fmt.Errorf("cannot get MythosApp")
		}
		logger := svrCtx.Logger.With("chain_id", chainId)

		// start the node
		// TODO send the configs to the smart contract
		// TODO send cmsrvconfig, ctndcfg with StartNode hook
		err = networkserver.StartNode(app, logger, app.GetNetworkKeeper())
		if err != nil {
			return err
		}
	}

	for _, ssfn := range startStateSyncProviders {
		ssfn()
	}

	// TODO - do we need this? (see simap)
	// if opts.PostSetup != nil {
	// 	if err := opts.PostSetup(svrCtx, clientCtx, ctx, g); err != nil {
	// 		return err
	// 	}
	// }

	// TODO do we need this?
	// var rosettaSrv crgserver.Server
	// if config.Rosetta.Enable {
	// 	offlineMode := config.Rosetta.Offline
	// 	if !config.GRPC.Enable { // If GRPC is not enabled rosetta cannot work in online mode, so it works in offline mode.
	// 		offlineMode = true
	// 	}

	// 	conf := &rosetta.Config{
	// 		Blockchain:    config.Rosetta.Blockchain,
	// 		Network:       config.Rosetta.Network,
	// 		TendermintRPC: ctx.Config.RPC.ListenAddress,
	// 		GRPCEndpoint:  config.GRPC.Address,
	// 		Addr:          config.Rosetta.Address,
	// 		Retries:       config.Rosetta.Retries,
	// 		Offline:       offlineMode,
	// 	}
	// 	conf.WithCodec(clientCtx.InterfaceRegistry, clientCtx.Codec.(*codec.ProtoCodec))

	// 	rosettaSrv, err = rosetta.ServerFromConfig(conf)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	errCh := make(chan error)
	// 	go func() {
	// 		if err := rosettaSrv.Start(); err != nil {
	// 			errCh <- err
	// 		}
	// 	}()

	// 	select {
	// 	case err := <-errCh:
	// 		return err
	// 	case <-time.After(types.ServerStartTime): // assume server started successfully
	// 	}
	// }

	// wait for signal capture and gracefully return
	// we are guaranteed to be waiting for the "ListenForQuitSignals" goroutine.
	return g.Wait()
}

type APICtx struct {
	GoRoutineGroup  *errgroup.Group
	GoContextParent context.Context
	SvrCtx          *server.Context
	ClientCtx       client.Context
	SrvCfg          srvconfig.Config
	TndCfg          *cmtcfg.Config
	AppCreator      func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp
	MetricsProvider node.MetricsProvider
	Metrics         *telemetry.Metrics
	Multiapp        *mcfg.MultiChainApp
}

func NewAPICtx(
	g *errgroup.Group,
	ctx context.Context,
	svrCtx *server.Context,
	clientCtx client.Context,
	msrvconfig srvconfig.Config,
	tndcfg *cmtcfg.Config,
	appCreator func(chainId string, chainCfg *menc.ChainConfig) mcfg.MythosApp,
	metricsProvider node.MetricsProvider,
	metrics *telemetry.Metrics,
	multiapp *mcfg.MultiChainApp,
) *APICtx {
	return &APICtx{
		GoRoutineGroup:  g,
		GoContextParent: ctx,
		SvrCtx:          svrCtx,
		ClientCtx:       clientCtx,
		SrvCfg:          msrvconfig,
		TndCfg:          tndcfg,
		AppCreator:      appCreator,
		MetricsProvider: metricsProvider,
		Metrics:         metrics,
		Multiapp:        multiapp,
	}
}

func (ac *APICtx) BuildConfigs(
	chainId string,
	chainCfg *menc.ChainConfig,
	ports mctx.NodePorts,
) (mcfg.MythosApp, *server.Context, client.Context, *srvconfig.Config, *cmtcfg.Config, client.CometRPC, error) {
	cmsrvconfig, ctndcfg, err := cloneConfigs(&ac.SrvCfg, ac.TndCfg)
	if err != nil {
		return nil, nil, ac.ClientCtx, nil, nil, nil, err
	}

	// TODO extract the port base ; define port family range in the toml files
	cmsrvconfig.Config.API.Address = strings.Replace(cmsrvconfig.Config.API.Address, "1317", fmt.Sprintf("%d", ports.CosmosRestApi), 1)
	cmsrvconfig.Config.GRPC.Address = strings.Replace(cmsrvconfig.Config.GRPC.Address, "9090", fmt.Sprintf("%d", ports.CosmosGrpc), 1)
	cmsrvconfig.Network.Address = strings.Replace(cmsrvconfig.Network.Address, "8090", fmt.Sprintf("%d", ports.WasmxNetworkGrpc), 1)
	cmsrvconfig.Websrv.Address = strings.Replace(cmsrvconfig.Websrv.Address, "9999", fmt.Sprintf("%d", ports.WebsrvWebServer), 1)
	cmsrvconfig.JsonRpc.Address = strings.Replace(cmsrvconfig.JsonRpc.Address, "8545", fmt.Sprintf("%d", ports.EvmJsonRpc), 1)
	cmsrvconfig.JsonRpc.WsAddress = strings.Replace(cmsrvconfig.JsonRpc.WsAddress, "8546", fmt.Sprintf("%d", ports.EvmJsonRpcWs), 1)
	ctndcfg.RPC.ListenAddress = strings.Replace(ctndcfg.RPC.ListenAddress, "26657", fmt.Sprintf("%d", ports.TendermintRpc), 1)

	var mythosapp_ mcfg.MythosApp
	found := false
	iapp, err := ac.Multiapp.GetApp(chainId)
	if err == nil {
		mythosapp_, found = iapp.(mcfg.MythosApp)
	}
	if !found {
		mythosapp_ = ac.AppCreator(chainId, chainCfg)
	}
	mythosapp, ok := mythosapp_.(*mapp.App)
	if !ok {
		return nil, nil, ac.ClientCtx, nil, nil, nil, fmt.Errorf("cannot convert MythosApp to App")
	}
	bapp := mythosapp.GetBaseApp()

	mythosapp.Logger().Info("starting chain api servers and clients", "chain_id", chainId)

	cclientCtx := ac.ClientCtx.WithChainID(chainId)

	csvrCtx := &server.Context{
		Viper:  ac.SvrCtx.Viper,
		Logger: ac.SvrCtx.Logger.With("chain_id", chainId),
		Config: ctndcfg,
	}

	rpcClient := networkgrpc.NewABCIClient(mythosapp, bapp, csvrCtx.Logger, mythosapp.GetNetworkKeeper(), csvrCtx.Config, cmsrvconfig, mythosapp.GetActionExecutor().(*networkgrpc.ActionExecutor))
	cclientCtx = cclientCtx.WithClient(rpcClient)

	mythosapp.SetServerConfig(cmsrvconfig)
	mythosapp.SetTendermintConfig(ctndcfg)
	mythosapp.SetRpcClient(rpcClient)
	mythosapp.NonDeterministicSetNodePorts(ports)

	return mythosapp, csvrCtx, cclientCtx, cmsrvconfig, ctndcfg, rpcClient, nil
}

func (ac *APICtx) StartChainApis(
	chainId string,
	chainCfg *menc.ChainConfig,
	ports mctx.NodePorts,
) (mcfg.MythosApp, *server.Context, client.Context, *srvconfig.Config, *cmtcfg.Config, client.CometRPC, error) {
	mythosapp_, csvrCtx, cclientCtx, cmsrvconfig, ctndcfg, rpcClient, err := ac.BuildConfigs(chainId, chainCfg, ports)
	if err != nil {
		return nil, nil, ac.ClientCtx, nil, nil, nil, err
	}
	mythosapp := mythosapp_.(*mapp.App)
	app := servertypes.Application(mythosapp)

	// Add the tx service to the gRPC router. We only need to register this
	// service if API or gRPC or JSONRPC is enabled, and avoid doing so in the general
	// case, because it spawns a new local tendermint RPC client.
	// if cmsrvconfig.API.Enable || cmsrvconfig.GRPC.Enable || cmsrvconfig.Websrv.Enable || cmsrvconfig.JsonRpc.Enable {
	// Re-assign for making the client available below do not use := to avoid
	// shadowing the clientCtx variable.
	mythosapp.Logger().Info("registering chain services", "chain_id", chainId)
	app.RegisterTxService(cclientCtx)
	app.RegisterTendermintService(cclientCtx)
	app.RegisterNodeService(cclientCtx, cmsrvconfig.Config)
	// }

	// Start the gRPC server in a goroutine. Note, the provided ctx will ensure
	// that the server is gracefully shut down.
	ac.GoRoutineGroup.Go(func() error {
		_, err = networkgrpc.StartGRPCServer(
			csvrCtx,
			cclientCtx,
			ac.GoContextParent,
			cmsrvconfig,
			app,
			ac.MetricsProvider,
			rpcClient,
		)
		if err != nil {
			csvrCtx.Logger.Error(err.Error())
		}
		return err
	})

	grpcSrv, cclientCtx, err := startGrpcServer(ac.GoContextParent, ac.GoRoutineGroup, cmsrvconfig.Config.GRPC, cclientCtx, csvrCtx, app)
	if err != nil {
		return nil, nil, ac.ClientCtx, nil, nil, nil, err
	}

	err = startAPIServer(ac.GoContextParent, ac.GoRoutineGroup, cmsrvconfig.Config, cclientCtx, csvrCtx, app, grpcSrv, ac.Metrics)
	if err != nil {
		return nil, nil, ac.ClientCtx, nil, nil, nil, err
	}

	// var (
	// 	httpSrv     *http.Server
	// 	httpSrvDone chan struct{}
	// )

	if cmsrvconfig.JsonRpc.Enable {
		tmEndpoint := "/websocket"
		tmRPCAddr := ctndcfg.RPC.ListenAddress

		// Start the gRPC server in a goroutine. Note, the provided ctx will ensure
		// that the server is gracefully shut down.
		ac.GoRoutineGroup.Go(func() error {
			// httpSrv, httpSrvDone, err
			_, _, err = jsonrpc.StartJsonRpc(csvrCtx, cclientCtx, ac.GoContextParent, tmRPCAddr, tmEndpoint, cmsrvconfig, chainId, *chainCfg)
			if err != nil {
				csvrCtx.Logger.Error(err.Error())
			}
			return err
		})
		// defer func() {
		// 	shutdownCtx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
		// 	defer cancelFn()
		// 	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		// 		logger.Error("HTTP server shutdown produced a warning", "error", err.Error())
		// 	} else {
		// 		logger.Info("HTTP server shut down, waiting 5 sec")
		// 		select {
		// 		case <-time.Tick(5 * time.Second):
		// 		case <-httpSrvDone:
		// 		}
		// 	}
		// }()
	}

	if cmsrvconfig.Websrv.Enable {
		ac.GoRoutineGroup.Go(func() error {
			// httpSrv, httpSrvDone, err
			_, _, err = websrv.StartWebsrv(csvrCtx, cclientCtx, ac.GoContextParent, &cmsrvconfig.Websrv)
			if err != nil {
				csvrCtx.Logger.Error(err.Error())
			}
			return err
		})
		// defer func() {
		// 	shutdownCtx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
		// 	defer cancelFn()
		// 	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		// 		logger.Error("HTTP server shutdown produced a warning", "error", err.Error())
		// 	} else {
		// 		logger.Info("HTTP server shut down, waiting 5 sec")
		// 		select {
		// 		case <-time.Tick(5 * time.Second):
		// 		case <-httpSrvDone:
		// 		}
		// 	}
		// }()
	}

	return mythosapp, csvrCtx, cclientCtx, cmsrvconfig, ctndcfg, rpcClient, nil
}

func startStateSync(
	goContextParent context.Context,
	goRoutineGroup *errgroup.Group,
	csvrCtx *server.Context,
	cmsrvconfig *srvconfig.Config,
	ctndcfg *cmtcfg.Config,
	chainId string,
	chainCfg menc.ChainConfig,
	app mcfg.MythosApp,
	rpcClient client.CometRPC,
	privateKey []byte,
	peers []string,
	currentNodeId int,
) error {
	if len(peers) <= currentNodeId {
		return fmt.Errorf("peers index out of bounds")
	}
	if len(peers) < 2 {
		return fmt.Errorf("need at least 2 peers")
	}
	// mythos1q77zrfhdctzgugutmnypyp0z2mg657e2hdwpqz@/ip4/127.0.0.1/tcp/5001/p2p/12D3KooWRRtnJfsJbRDMrMQQd5wopPDsjM9urKsLb9VzA1Y49udr
	port := strings.Split(peers[currentNodeId], "/")[4]

	peerId := 0
	if currentNodeId == 0 {
		peerId = 1
	}
	peeraddress := strings.Split(peers[peerId], "@")[1]

	currentIdStr := "0"
	nodeids := strings.Split(cmsrvconfig.Network.Id, ";")
	for _, nodeid := range nodeids {
		chainIdPair := strings.Split(nodeid, ":")
		if len(chainIdPair) == 1 {
			currentIdStr = chainIdPair[0]
		} else if len(chainIdPair) > 1 && chainIdPair[0] == chainId {
			currentIdStr = chainIdPair[1]
			break
		}
	}
	currentId, err := strconv.Atoi(currentIdStr)
	if err != nil {
		return err
	}

	ssctx, err := vmp2p.InitializeStateSyncWithPeer(goContextParent, goRoutineGroup, csvrCtx.Logger, ctndcfg, chainId, chainCfg, app, rpcClient, mcfg.GetStateSyncProtocolId(chainId), peeraddress, privateKey, port, peers, int32(currentId))
	if err != nil {
		return err
	}

	err = node.StartStateSync(ssctx.StateSyncReactor, ssctx.BcReactor, ssctx.StateSyncProvider, ctndcfg.StateSync, ssctx.StateStore, nil, ssctx.StateSyncGenesis)
	if err != nil {
		return fmt.Errorf("failed to start state sync: %w", err)
	}
	return nil
}

func startStateSyncProvider(
	goContextParent context.Context,
	goRoutineGroup *errgroup.Group,
	csvrCtx *server.Context,
	cmsrvconfig *srvconfig.Config,
	ctndcfg *cmtcfg.Config,
	chainId string,
	chainCfg menc.ChainConfig,
	app mcfg.MythosApp,
	rpcClient client.CometRPC,
	privateKey []byte,
) {
	peers := networktypes.GetPeersFromConfigIps(chainId, cmsrvconfig.Network.Ips)
	port := strings.Split(peers[0], "/")[4]
	go func() {
		err := vmp2p.InitializeStateSyncProvider(goContextParent, goRoutineGroup, csvrCtx.Logger, ctndcfg, chainId, chainCfg, app, rpcClient, mcfg.GetStateSyncProtocolId(chainId), privateKey, port)
		if err != nil {
			csvrCtx.Logger.Error("InitializeStateSyncProvider", "error", err.Error())
		}
	}()
}

func cloneConfigs(msrvconfig *srvconfig.Config, tndcfg *cmtcfg.Config) (*srvconfig.Config, *cmtcfg.Config, error) {
	newmsrvconfig := &srvconfig.Config{}
	newtndcfg := &cmtcfg.Config{}
	msrvconfigbz, err := json.Marshal(msrvconfig)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(msrvconfigbz, newmsrvconfig)
	if err != nil {
		return nil, nil, err
	}

	tndcfgbz, err := json.Marshal(tndcfg)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(tndcfgbz, newtndcfg)
	if err != nil {
		return nil, nil, err
	}
	return newmsrvconfig, newtndcfg, nil
}

// TODO: Move nodeKey into being created within the function.
func startCmtNode(
	ctx context.Context,
	cfg *cmtcfg.Config,
	app servertypes.Application,
	svrCtx *server.Context,
) (tmNode *node.Node, cleanupFn func(), err error) {
	nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	if err != nil {
		return nil, cleanupFn, err
	}

	cmtApp := server.NewCometABCIWrapper(app)
	tmNode, err = node.NewNodeWithContext(
		ctx,
		cfg,
		pvm.LoadOrGenFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile()),
		nodeKey,
		proxy.NewLocalClientCreator(cmtApp),
		getGenDocProvider(cfg),
		cmtcfg.DefaultDBProvider,
		node.DefaultMetricsProvider(cfg.Instrumentation),
		servercmtlog.CometLoggerWrapper{Logger: svrCtx.Logger},
	)

	if err != nil {
		return tmNode, cleanupFn, err
	}

	if err := tmNode.Start(); err != nil {
		return tmNode, cleanupFn, err
	}

	cleanupFn = func() {
		if tmNode != nil && tmNode.IsRunning() {
			_ = tmNode.Stop()
		}
	}

	return tmNode, cleanupFn, nil
}

// returns a function which returns the genesis doc from the genesis file.
func getGenDocProvider(cfg *cmtcfg.Config) func() (*cmttypes.GenesisDoc, error) {
	return func() (*cmttypes.GenesisDoc, error) {
		genFile := cfg.GenesisFile()
		appGenesis, err := genutiltypes.AppGenesisFromFile(genFile)
		if err != nil {
			return nil, err
		}

		return appGenesis.ToGenesisDoc()
	}
}

// returns a function which returns the genesis doc from the genesis file.
func getGenDocProvider2(cfg *cmtcfg.Config) func(chainId string) (*cmttypes.GenesisDoc, error) {
	return func(chainId string) (*cmttypes.GenesisDoc, error) {
		genFile := cfg.GenesisFile()
		if chainId != "" {
			genFile = strings.Replace(genFile, ".json", "_"+chainId+".json", 1)
		}
		appGenesis, err := genutiltypes.AppGenesisFromFile(genFile)
		if err != nil {
			return nil, err
		}

		return appGenesis.ToGenesisDoc()
	}
}

func startGrpcServer(
	ctx context.Context,
	g *errgroup.Group,
	config sdkserverconfig.GRPCConfig,
	clientCtx client.Context,
	svrCtx *server.Context,
	app servertypes.Application,
) (*grpc.Server, client.Context, error) {
	if !config.Enable {
		// return grpcServer as nil if gRPC is disabled
		return nil, clientCtx, nil
	}
	_, port, err := net.SplitHostPort(config.Address)
	if err != nil {
		return nil, clientCtx, err
	}

	maxSendMsgSize := config.MaxSendMsgSize
	if maxSendMsgSize == 0 {
		maxSendMsgSize = sdkserverconfig.DefaultGRPCMaxSendMsgSize
	}

	maxRecvMsgSize := config.MaxRecvMsgSize
	if maxRecvMsgSize == 0 {
		maxRecvMsgSize = sdkserverconfig.DefaultGRPCMaxRecvMsgSize
	}

	grpcAddress := fmt.Sprintf("127.0.0.1:%s", port)

	// if gRPC is enabled, configure gRPC client for gRPC gateway
	grpcClient, err := grpc.Dial(
		grpcAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.ForceCodec(codec.NewProtoCodec(clientCtx.InterfaceRegistry).GRPCCodec()),
			grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
			grpc.MaxCallSendMsgSize(maxSendMsgSize),
		),
	)
	if err != nil {
		return nil, clientCtx, err
	}

	clientCtx = clientCtx.WithGRPCClient(grpcClient)
	svrCtx.Logger.Debug("gRPC client assigned to client context", "target", grpcAddress)

	grpcSrv, err := servergrpc.NewGRPCServer(clientCtx, app, config)
	if err != nil {
		return nil, clientCtx, err
	}

	// Start the gRPC server in a goroutine. Note, the provided ctx will ensure
	// that the server is gracefully shut down.
	g.Go(func() error {
		return servergrpc.StartGRPCServer(ctx, svrCtx.Logger.With(log.ModuleKey, "grpc-server"), config, grpcSrv)
	})
	return grpcSrv, clientCtx, nil
}

func startAPIServer(
	ctx context.Context,
	g *errgroup.Group,
	svrCfg sdkserverconfig.Config,
	clientCtx client.Context,
	svrCtx *server.Context,
	app servertypes.Application,
	grpcSrv *grpc.Server,
	metrics *telemetry.Metrics,
) error {
	if !svrCfg.API.Enable {
		return nil
	}

	apiSrv := api.New(clientCtx, svrCtx.Logger.With(log.ModuleKey, "api-server"), grpcSrv)
	app.RegisterAPIRoutes(apiSrv, svrCfg.API)

	if svrCfg.Telemetry.Enabled {
		apiSrv.SetTelemetry(metrics)
	}

	g.Go(func() error {
		return apiSrv.Start(ctx, svrCfg)
	})
	return nil
}

func startTelemetry(cfg sdkserverconfig.Config) (*telemetry.Metrics, error) {
	if !cfg.Telemetry.Enabled {
		return nil, nil
	}

	return telemetry.New(cfg.Telemetry)
}

func getCtx(svrCtx *server.Context, block bool) (*errgroup.Group, context.Context) {
	ctx, cancelFn := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	// listen for quit signals so the calling parent process can gracefully exit
	server.ListenForQuitSignals(g, block, cancelFn, svrCtx.Logger)
	return g, ctx
}

func openDB(rootDir string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB("application", backendType, dataDir)
}

// OpenIndexerDB opens the custom eth indexer db, using the same db backend as the main app
func OpenIndexerDB(rootDir string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB("evmindexer", backendType, dataDir)
}

func openTraceWriter(traceWriterFile string) (w io.Writer, err error) {
	if traceWriterFile == "" {
		return
	}

	filePath := filepath.Clean(traceWriterFile)
	return os.OpenFile(
		filePath,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0o600,
	)
}
