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

	srvconfig "mythos/v1/server/config"
	srvflags "mythos/v1/server/flags"
	websrv "mythos/v1/x/websrv/server"
	websrvconfig "mythos/v1/x/websrv/server/config"
	websrvflags "mythos/v1/x/websrv/server/flags"

	jsonrpc "mythos/v1/x/wasmx/server"
	jsonrpcconfig "mythos/v1/x/wasmx/server/config"
	jsonrpcflags "mythos/v1/x/wasmx/server/flags"

	networkgrpc "mythos/v1/x/network/keeper"
	networkserver "mythos/v1/x/network/server"
	networkconfig "mythos/v1/x/network/server/config"
	networkflags "mythos/v1/x/network/server/flags"

	mapp "mythos/v1/app"
	mcfg "mythos/v1/config"
)

// StartCmd runs the service passed in, either stand-alone or in-process with
// Tendermint.
func StartCmd(appCreator servertypes.AppCreator, defaultNodeHome string) *cobra.Command {
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

			serverCtx.Logger.Info("starting ABCI with Tendermint")

			// amino is needed here for backwards compatibility of REST routes
			return startInProcess(serverCtx, clientCtx, appCreator)
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

	// add support for all Tendermint-specific command line options
	tcmd.AddNodeFlags(cmd)
	return cmd
}

func startStandAlone(svrCtx *server.Context, _ servertypes.AppCreator) error {
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

	bapps, appCreator := mapp.NewAppCreator(
		svrCtx.Logger,
		db,
		traceWriter,
		svrCtx.Viper,
		g, ctx,
	)

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

	svr.SetLogger(servercmtlog.CometLoggerWrapper{Logger: svrCtx.Logger.With("module", "abci-server")})

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
func startInProcess(svrCtx *server.Context, clientCtx client.Context, _ servertypes.AppCreator) (err error) {
	tndcfg := svrCtx.Config
	home := tndcfg.RootDir
	logger := svrCtx.Logger

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

	multiapp, appCreator := mapp.NewAppCreator(
		logger,
		db,
		traceWriter,
		svrCtx.Viper,
		g, ctx,
	)

	clientCtx = clientCtx.
		WithHomeDir(home).
		WithChainID(genDoc.ChainID)

	metrics, err := startTelemetry(msrvconfig.Config)
	if err != nil {
		return err
	}
	metricsProvider := node.DefaultMetricsProvider(tndcfg.Instrumentation)

	// create apps for initial chains
	initialChains := msrvconfig.Network.InitialChains
	for i, chainId := range initialChains {
		chainCfg, err := mcfg.GetChainConfig(chainId)
		if err != nil {
			panic(err)
		}

		mythosapp_ := appCreator(chainId, chainCfg)
		mythosapp, ok := mythosapp_.(*mapp.App)
		if !ok {
			return fmt.Errorf("cannot convert MythosApp to App")
		}
		app := servertypes.Application(mythosapp)
		bapp := mythosapp.GetBaseApp()

		cclientCtx := clientCtx.WithChainID(chainId)

		cmsrvconfig, ctndcfg, err := cloneConfigs(&msrvconfig, tndcfg)
		if err != nil {
			panic(err)
		}
		fmt.Println("---chain---", chainId)
		fmt.Println("---cmsrvconfig---", cmsrvconfig)
		fmt.Println("---ctndcfg---", ctndcfg)

		// TODO extract the port base ; define port family range in the toml files
		cmsrvconfig.Config.API.Address = strings.Replace(cmsrvconfig.Config.API.Address, "1317", fmt.Sprintf("%d", 1317+i), 1)
		cmsrvconfig.Config.GRPC.Address = strings.Replace(cmsrvconfig.Config.GRPC.Address, "9090", fmt.Sprintf("%d", 9090+i), 1)
		cmsrvconfig.Network.Address = strings.Replace(cmsrvconfig.Network.Address, "8090", fmt.Sprintf("%d", 8090+i), 1)
		cmsrvconfig.Websrv.Address = strings.Replace(cmsrvconfig.Websrv.Address, "9900", fmt.Sprintf("%d", 9900+i), 1)
		cmsrvconfig.JsonRpc.Address = strings.Replace(cmsrvconfig.JsonRpc.Address, "8555", fmt.Sprintf("%d", 8555+i*2), 1)
		cmsrvconfig.JsonRpc.WsAddress = strings.Replace(cmsrvconfig.JsonRpc.WsAddress, "8555", fmt.Sprintf("%d", 8556+i), 1)

		ctndcfg.RPC.ListenAddress = strings.Replace(ctndcfg.RPC.ListenAddress, "26657", fmt.Sprintf("%d", 26657+i), 1)

		csvrCtx := &server.Context{
			Viper:  svrCtx.Viper,
			Logger: svrCtx.Logger.With("chain_id", chainId),
			Config: ctndcfg,
		}

		privValidator := pvm.LoadOrGenFilePV(ctndcfg.PrivValidatorKeyFile(), ctndcfg.PrivValidatorStateFile())
		genesisDocProvider := getGenDocProvider2(ctndcfg)

		rpcClient := networkgrpc.NewABCIClient(mythosapp, bapp, csvrCtx.Logger, mythosapp.GetNetworkKeeper(), csvrCtx.Config, cmsrvconfig, mythosapp.GetActionExecutor().(*networkgrpc.ActionExecutor))
		cclientCtx = cclientCtx.WithClient(rpcClient)

		// Start the gRPC server in a goroutine. Note, the provided ctx will ensure
		// that the server is gracefully shut down.
		g.Go(func() error {
			_, err = networkgrpc.StartGRPCServer(
				csvrCtx,
				cclientCtx,
				ctx,
				cmsrvconfig,
				app,
				privValidator,
				nodeKey,
				genesisDocProvider,
				metricsProvider,
				rpcClient,
			)
			if err != nil {
				csvrCtx.Logger.Error(err.Error())
			}
			return err
		})

		// Add the tx service to the gRPC router. We only need to register this
		// service if API or gRPC or JSONRPC is enabled, and avoid doing so in the general
		// case, because it spawns a new local tendermint RPC client.
		// if cmsrvconfig.API.Enable || cmsrvconfig.GRPC.Enable || cmsrvconfig.Websrv.Enable || cmsrvconfig.JsonRpc.Enable {
		// Re-assign for making the client available below do not use := to avoid
		// shadowing the clientCtx variable.
		app.RegisterTxService(cclientCtx)
		app.RegisterTendermintService(cclientCtx)
		app.RegisterNodeService(cclientCtx, cmsrvconfig.Config)
		// }

		grpcSrv, cclientCtx, err := startGrpcServer(ctx, g, cmsrvconfig.Config.GRPC, cclientCtx, csvrCtx, app)
		if err != nil {
			return err
		}

		err = startAPIServer(ctx, g, cmsrvconfig.Config, cclientCtx, csvrCtx, app, grpcSrv, metrics)
		if err != nil {
			return err
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
			g.Go(func() error {
				// httpSrv, httpSrvDone, err
				_, _, err = jsonrpc.StartJsonRpc(csvrCtx, cclientCtx, ctx, tmRPCAddr, tmEndpoint, cmsrvconfig)
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
			g.Go(func() error {
				// httpSrv, httpSrvDone, err
				_, _, err = websrv.StartWebsrv(csvrCtx, cclientCtx, ctx, &cmsrvconfig.Websrv)
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

		// initialize chain if this is block 0
		// init all chains first and start them afterwards
		// InitChain runs multiple contract executions that are not under ActionExecutor control; starting the chains while InitChain is not finished will start delayed executions that will intersect with InitChain executions
		if bapp.LastBlockHeight() == 0 {
			_, err := networkgrpc.InitChain(csvrCtx.Logger, &msrvconfig, app, privValidator, nodeKey, genesisDocProvider, chainId, mythosapp.GetNetworkKeeper())
			if err != nil {
				return err
			}
		}
	}

	// start nodes for all chains
	// should be all chains, taken from level0
	for _, chainId := range multiapp.ChainIds {
		iapp, _ := multiapp.GetApp(chainId)
		app, ok := iapp.(mcfg.MythosApp)
		if !ok {
			return fmt.Errorf("cannot get MythosApp")
		}
		logger := svrCtx.Logger.With("chain_id", chainId)

		// start the node
		err = networkserver.StartNode(app, logger, app.GetNetworkKeeper())
		if err != nil {
			return err
		}
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
		return servergrpc.StartGRPCServer(ctx, svrCtx.Logger.With("module", "grpc-server"), config, grpcSrv)
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

	apiSrv := api.New(clientCtx, svrCtx.Logger.With("module", "api-server"), grpcSrv)
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
