package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"cosmossdk.io/log"

	// "github.com/cosmos/cosmos-sdk/baseapp"

	// runapp "github.com/cosmos/cosmos-sdk/runtime"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino" // Import amino.proto file for reflection

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/merkle"
	"github.com/cometbft/cometbft/libs/bytes"
	cmttypes "github.com/cometbft/cometbft/types"

	mcfg "mythos/v1/config"
	mctx "mythos/v1/context"
	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

func InitChainAndCommitBlock(
	app servertypes.Application,
	req *abci.RequestInitChain,
	logger log.Logger,
) (*abci.ResponseInitChain, *abci.ResponseInfo, error) {
	// TODO catch panic so other chains remain unaffected
	res, err := app.Info(types.RequestInfo)
	if err != nil {
		return nil, nil, fmt.Errorf("error calling Info: %v", err)
	}
	blockHeight := res.LastBlockHeight
	if blockHeight < 0 {
		return nil, res, fmt.Errorf("got a negative last block height (%d) from the app", blockHeight)
	}

	logger.Info("ABCI Handshake App Info",
		"height", blockHeight,
		"hash", fmt.Sprintf("%X", res.LastBlockAppHash),
		"software-version", res.Version,
		"protocol-version", res.AppVersion,
	)

	resInit, err := app.InitChain(req)
	if err != nil {
		return nil, res, err
	}

	freq := &abci.RequestFinalizeBlock{
		Height: req.InitialHeight,
		Time:   req.Time,
	}
	_, err = app.FinalizeBlock(freq)
	if err != nil {
		return nil, res, err
	}

	_, err = app.Commit()
	if err != nil {
		return nil, res, err
	}

	// If the app returned validators or consensus params, update the state.
	validatorsUpdates := resInit.Validators
	if len(validatorsUpdates) == 0 {
		validatorsUpdates = req.Validators
	}
	if len(validatorsUpdates) == 0 {
		// If validator set is not set in genesis and still empty after InitChain, exit.
		return nil, res, fmt.Errorf("validator set is nil in genesis and still empty after InitChain")
	}
	return resInit, res, nil
}

func InitConsensusContract(
	mythosapp mcfg.MythosApp,
	consensusLogger log.Logger,
	networkServer mcfg.NetworkKeeper,
	appHash []byte,
	consensusParams *cmttypes.ConsensusParams,
	appVersion uint64,
	pubAddr bytes.HexBytes,
	pubKey, privKey []byte,
	nodeIndex int32,
	peers []string,
	nodePorts mctx.NodePorts,
) error {
	version := types.Version{
		Software: "",
		Consensus: types.Consensus{
			App:   appVersion,
			Block: 0,
		},
	}
	// TODO ?
	// version.Consensus.App = consensusParams.Version.App

	initChainSetup := &types.InitChainSetup{
		ChainID:         mythosapp.GetBaseApp().ChainID(),
		ConsensusParams: consensusParams,
		AppHash:         appHash,
		// We update the last results hash with the empty hash, to conform with RFC-6962.
		LastResultsHash:  merkle.HashFromByteSlices(nil),
		Version:          version,
		ValidatorAddress: pubAddr, // hexbytes
		ValidatorPrivKey: privKey, // consensus privkey
		ValidatorPubKey:  pubKey,  // consensus pubkey
		Peers:            peers,
		NodeIndex:        nodeIndex,
		InitialPorts:     nodePorts,
	}

	// TODO check if app block height is same as network block height
	// https://github.com/cometbft/cometbft/blob/9cccc8c463f204b210b2a290c2066445188dc681/internal/consensus/replay.go#L360

	// setup the consensus contract
	err := SetupNode(mythosapp, consensusLogger, networkServer, initChainSetup)
	if err != nil {
		return err
	}
	return nil
}

func StartNode(mythosapp mcfg.MythosApp, logger log.Logger, networkServer mcfg.NetworkKeeper) error {
	cb := func(goctx context.Context) (any, error) {
		ctx := sdk.UnwrapSDKContext(goctx)
		msg := []byte(fmt.Sprintf(`{"RunHook":{"hook":"%s","data":""}}`, wasmxtypes.HOOK_START_NODE))
		res, err := networkServer.ExecuteContract(ctx, &types.MsgExecuteContract{
			Sender:   wasmxtypes.ROLE_HOOKS_NONC,
			Contract: wasmxtypes.ROLE_HOOKS_NONC,
			Msg:      msg,
		})
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	actionExecutor := mythosapp.GetActionExecutor()

	_, err := actionExecutor.Execute(mythosapp.GetGoContextParent(), mythosapp.GetBaseApp().LastBlockHeight(), cb)
	if err != nil {
		return err
	}
	return nil
}

func SetupNode(mythosapp mcfg.MythosApp, logger log.Logger, networkServer mcfg.NetworkKeeper, initChainSetup *types.InitChainSetup) error {
	cb := func(goctx context.Context) (any, error) {
		ctx := sdk.UnwrapSDKContext(goctx)

		initbz, err := json.Marshal(initChainSetup)
		if err != nil {
			return nil, err
		}
		initData := base64.StdEncoding.EncodeToString(initbz)
		msg := []byte(fmt.Sprintf(`{"RunHook":{"hook":"%s","data":"%s"}}`, wasmxtypes.HOOK_SETUP_NODE, initData))
		res, err := networkServer.ExecuteContract(ctx, &types.MsgExecuteContract{
			Sender:   wasmxtypes.ROLE_HOOKS_NONC,
			Contract: wasmxtypes.ROLE_HOOKS_NONC,
			Msg:      msg,
		})
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	actionExecutor := mythosapp.GetActionExecutor()
	_, err := actionExecutor.Execute(mythosapp.GetGoContextParent(), mythosapp.GetBaseApp().LastBlockHeight(), cb)
	if err != nil {
		return err
	}
	return nil
}
