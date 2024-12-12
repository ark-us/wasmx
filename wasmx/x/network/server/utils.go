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

	mcodec "github.com/loredanacirstea/wasmx/v1/codec"
	mcfg "github.com/loredanacirstea/wasmx/v1/config"
	mctx "github.com/loredanacirstea/wasmx/v1/context"
	"github.com/loredanacirstea/wasmx/v1/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/v1/x/wasmx/types"
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

	_, err := actionExecutor.ExecuteWithMockHeader(mythosapp.GetGoContextParent(), cb)
	if err != nil {
		return err
	}

	// we update checkState on the baseapp
	bapp := mythosapp.GetBaseApp()
	lastHeight := bapp.LastBlockHeight()
	if lastHeight > 0 {
		logger.Info("setting checkTx header from last known block", "height", lastHeight)
		header, err := networkServer.GetHeaderByHeight(mythosapp, logger, lastHeight, false)
		if err != nil || header == nil {
			// we do not throw an error now because this will happen right after bootstrap, when we do not have block data
			// TODO reconsider this when we have block sync too, so we can sync at least the last block
			logger.Info("StartNode: could not get header by height for checkTx: %s", err.Error())
			return nil
		}
		bapp.SetCheckStateHeader(*header)
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
	_, err := actionExecutor.ExecuteWithMockHeader(mythosapp.GetGoContextParent(), cb)
	if err != nil {
		return err
	}
	return nil
}

func InitializeSingleConsensusContracts(mythosapp mcfg.MythosApp, logger log.Logger, networkServer mcfg.NetworkKeeper) error {
	cb := func(goctx context.Context) (any, error) {
		ctx := sdk.UnwrapSDKContext(goctx)
		var erreur error

		mythosapp.GetWasmxKeeper().IterateContractInfo(ctx, func(addr sdk.AccAddress, cinfo wasmxtypes.ContractInfo) bool {
			// TODO metaconsensus state sync as an extension and remove from here
			if cinfo.StorageType != wasmxtypes.ContractStorageType_SingleConsensus && cinfo.StorageType != wasmxtypes.ContractStorageType_MetaConsensus {
				return false
			}
			// we only need to run contract instantiation
			// so contracts persist instantiation storage
			bootstrapAddr, err := mcodec.AccAddressPrefixedFromBech32(cinfo.Creator)
			if err != nil {
				erreur = err
				return true
			}
			contractAddressBech32, err := mythosapp.AddressCodec().BytesToString(addr)
			if err != nil {
				erreur = err
				return true
			}
			contractAddress, err := mcodec.AccAddressPrefixedFromBech32(contractAddressBech32)
			if err != nil {
				erreur = err
				return true
			}
			codeInfo := mythosapp.GetWasmxKeeper().GetCodeInfo(ctx, cinfo.CodeId)
			if codeInfo == nil {
				erreur = fmt.Errorf("no code info found for codeID %d", cinfo.CodeId)
				return true
			}

			_, _, err = mythosapp.GetWasmxKeeper().ExecuteContractInstantiationInternal(
				ctx,
				cinfo.CodeId,
				codeInfo,
				bootstrapAddr,
				contractAddress,
				cinfo.StorageType,
				cinfo.InitMessage,
				nil,
				cinfo.Label,
			)
			if err != nil {
				erreur = err
				return true
			}

			return false
		})
		if erreur != nil {
			return nil, erreur
		}
		return nil, nil
	}

	actionExecutor := mythosapp.GetActionExecutor()
	_, err := actionExecutor.ExecuteWithMockHeader(mythosapp.GetGoContextParent(), cb)
	if err != nil {
		return err
	}
	return nil
}

func ConsensusTx(mythosapp mcfg.MythosApp, logger log.Logger, networkServer mcfg.NetworkKeeper, msg []byte) error {
	cb := func(goctx context.Context) (any, error) {
		ctx := sdk.UnwrapSDKContext(goctx)
		res, err := networkServer.ExecuteContract(ctx, &types.MsgExecuteContract{
			Sender:   wasmxtypes.ROLE_CONSENSUS,
			Contract: wasmxtypes.ROLE_CONSENSUS,
			Msg:      msg,
		})
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	actionExecutor := mythosapp.GetActionExecutor()

	_, err := actionExecutor.ExecuteWithMockHeader(mythosapp.GetGoContextParent(), cb)
	if err != nil {
		return err
	}
	return nil
}

func ConsensusQuery(mythosapp mcfg.MythosApp, logger log.Logger, networkServer mcfg.NetworkKeeper, msg []byte) (*types.MsgQueryContractResponse, error) {
	cb := func(goctx context.Context) (any, error) {
		ctx := sdk.UnwrapSDKContext(goctx)
		res, err := networkServer.QueryContract(ctx, &types.MsgQueryContract{
			Sender:   wasmxtypes.ROLE_CONSENSUS,
			Contract: wasmxtypes.ROLE_CONSENSUS,
			Msg:      msg,
		})
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	actionExecutor := mythosapp.GetActionExecutor()
	resp, err := actionExecutor.ExecuteWithMockHeader(mythosapp.GetGoContextParent(), cb)
	if err != nil {
		return nil, err
	}
	return resp.(*types.MsgQueryContractResponse), nil
}
