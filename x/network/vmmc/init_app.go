package vmmc

import (
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	tndtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	mcfg "mythos/v1/config"
	networkserver "mythos/v1/x/network/server"
)

type GenutilGenesisState struct {
	GenTxs [][]byte `json:"gen_txs"`
}

func InitApp(ctx *Context, req *InitSubChainMsg) (*abci.ResponseInitChain, error) {
	// TODO catch the panic! so other chains don't stop
	if req == nil {
		return nil, fmt.Errorf("cannot initialize chain with empty request")
	}
	logger := ctx.Logger(ctx.Ctx).With("subchain_id", req.InitChainRequest.ChainId)

	multichainapp, err := mcfg.GetMultiChainApp(ctx.GoContextParent)
	if err != nil {
		return nil, err
	}
	app := multichainapp.NewApp(req.InitChainRequest.ChainId, &req.ChainConfig)
	sapp, ok := app.(servertypes.Application)
	if !ok {
		return nil, fmt.Errorf("cannot convert MythosApp to server Application")
	}
	genesisStateWasmx := map[string][]byte{}
	err = json.Unmarshal(req.InitChainRequest.AppStateBytes, &genesisStateWasmx)
	if err != nil {
		return nil, err
	}

	genesisState := map[string]json.RawMessage{}
	for key, value := range genesisStateWasmx {
		genesisState[key] = value

		// TODO remove this when we replace genutil
		if key == genutiltypes.ModuleName {
			genutilGenesis := &GenutilGenesisState{}
			err = json.Unmarshal(value, genutilGenesis)
			if err != nil {
				return nil, err
			}
			cdc := app.AppCodec()
			genTx := make([]json.RawMessage, len(genutilGenesis.GenTxs))
			for i, txbz := range genutilGenesis.GenTxs {
				genTx[i] = txbz
			}
			newGenutilGenesis := &genutiltypes.GenesisState{GenTxs: genTx}
			newvalue, err := cdc.MarshalJSON(newGenutilGenesis)
			if err != nil {
				return nil, err
			}
			genesisState[key] = newvalue
		}
	}
	stateBytes, err := json.Marshal(&genesisState)
	if err != nil {
		return nil, err
	}
	req.InitChainRequest.AppStateBytes = stateBytes

	resInit, res, err := networkserver.InitChainAndCommitBlock(sapp, &req.InitChainRequest, logger)
	if err != nil {
		return nil, err
	}
	appHash := res.LastBlockAppHash
	if len(resInit.AppHash) > 0 {
		appHash = resInit.AppHash
	}

	consensusParams := *req.InitChainRequest.ConsensusParams
	if resInit.ConsensusParams != nil {
		consensusParams = UpdateParams(consensusParams, resInit.ConsensusParams)
	}
	constp := cmttypes.ConsensusParamsFromProto(consensusParams)

	// InitChain
	err = networkserver.InitConsensusContract(app, logger, app.GetNetworkKeeper(), appHash, &constp, res.AppVersion, req.ValidatorAddress, req.ValidatorPubKey, req.ValidatorPrivKey, req.CurrentNodeId, req.Peers, req.InitialPorts)
	if err != nil {
		return resInit, err
	}
	return resInit, nil
}

func StartApp(ctx *Context, req *StartSubChainMsg) error {
	// TODO catch the panic! so other chains don't stop
	if req == nil {
		return fmt.Errorf("cannot start node with empty request")
	}
	logger := ctx.Logger(ctx.Ctx)
	logger.Info("starting subchain", "subchain_id", req.ChainId)
	multichainapp, err := mcfg.GetMultiChainApp(ctx.GoContextParent)
	if err != nil {
		return err
	}
	mcfg.CacheChainConfig(req.ChainId, req.ChainConfig)
	var app mcfg.MythosApp
	found := false
	iapp, err := multichainapp.GetApp(req.ChainId)
	if err == nil {
		app_, ok := iapp.(mcfg.MythosApp)
		if ok {
			app = app_
			found = true
		}
	}
	if !found {
		app = multichainapp.NewApp(req.ChainId, &req.ChainConfig)
	}

	// start API servers
	_, _, _, _, _, _, err = multichainapp.APICtx.StartChainApis(req.ChainId, &req.ChainConfig, req.NodePorts)
	if err != nil {
		return err
	}
	err = networkserver.StartNode(app, logger, app.GetNetworkKeeper())
	if err != nil {
		return err
	}

	InitializeStateSyncProvider(
		ctx.GoContextParent,
		ctx.GoRoutineGroup,
		logger,
		app.GetTendermintConfig(),
		req.ChainId,
		app,
		app.GetRpcClient(),
		mcfg.GetStateSyncProtocolId(req.ChainId),
		fmt.Sprintf("%d", req.NodePorts.WasmxNetworkP2P),
	)

	return nil
}

// Update returns a copy of the params with updates from the non-zero fields of p2.
// NOTE: note: must not modify the original
func UpdateParams(params tndtypes.ConsensusParams, params2 *tndtypes.ConsensusParams) tndtypes.ConsensusParams {
	res := params // explicit copy

	if params2 == nil {
		return res
	}

	// we must defensively consider any structs may be nil
	if params2.Block != nil {
		res.Block.MaxBytes = params2.Block.MaxBytes
		res.Block.MaxGas = params2.Block.MaxGas
	}
	if params2.Evidence != nil {
		res.Evidence.MaxAgeNumBlocks = params2.Evidence.MaxAgeNumBlocks
		res.Evidence.MaxAgeDuration = params2.Evidence.MaxAgeDuration
		res.Evidence.MaxBytes = params2.Evidence.MaxBytes
	}
	if params2.Validator != nil {
		// Copy params2.Validator.PubkeyTypes, and set result's value to the copy.
		// This avoids having to initialize the slice to 0 values, and then write to it again.
		res.Validator.PubKeyTypes = append([]string{}, params2.Validator.PubKeyTypes...)
	}
	if params2.Version != nil {
		res.Version.App = params2.Version.App
	}
	if params2.Abci != nil {
		res.Abci.VoteExtensionsEnableHeight = params2.Abci.GetVoteExtensionsEnableHeight()
	}
	return res
}
