package vmmc

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	tndtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	mcfg "mythos/v1/config"
	menc "mythos/v1/encoding"
	networkserver "mythos/v1/x/network/server"
)

func InitApp(ctx *Context, req *abci.RequestInitChain, cfg *menc.ChainConfig, nodeAddress, nodePubKey, nodePrivKey []byte, peers []string) (*abci.ResponseInitChain, error) {
	logger := ctx.Logger(ctx.Ctx)
	multichainapp, err := mcfg.GetMultiChainApp(ctx.GoContextParent)
	if err != nil {
		return nil, err
	}
	app := multichainapp.NewApp(req.ChainId, cfg)
	sapp, ok := app.(servertypes.Application)
	if !ok {
		return nil, fmt.Errorf("cannot convert MythosApp to server Application")
	}

	resInit, res, err := networkserver.InitChainAndCommitBlock(sapp, req, logger)
	if err != nil {
		return nil, err
	}
	appHash := res.LastBlockAppHash
	if len(resInit.AppHash) > 0 {
		appHash = resInit.AppHash
	}

	consensusParams := *req.ConsensusParams
	if resInit.ConsensusParams != nil {
		consensusParams = UpdateParams(consensusParams, resInit.ConsensusParams)
	}
	constp := cmttypes.ConsensusParamsFromProto(consensusParams)

	// InitChain
	err = networkserver.InitConsensusContract(app, logger, app.GetNetworkKeeper(), appHash, &constp, res.AppVersion, nodeAddress, nodePubKey, nodePrivKey, 0, peers)
	if err != nil {
		return resInit, err
	}

	// TODO level0 start node hook starts the subchains
	// start chain - StartNode

	err = networkserver.StartNode(app, logger, app.GetNetworkKeeper())
	if err != nil {
		return resInit, err
	}
	return resInit, nil
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
