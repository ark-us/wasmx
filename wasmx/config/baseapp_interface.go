package config

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
)

type GenesisDocProvider func(string) (*cmttypes.GenesisDoc, error)

type BaseApp interface {
	Name() string
	ChainID() string
	CreateQueryContext(height int64, prove bool) (sdk.Context, error)
	CommitMultiStore() storetypes.CommitMultiStore
	GetContextForCheckTx(txBytes []byte) sdk.Context
	GetContextForFinalizeBlock(txBytes []byte) sdk.Context
	NewUncachedContext(isCheckTx bool, header cmtproto.Header) sdk.Context
	LastBlockHeight() int64
	LastCommitID() storetypes.CommitID
	TxDecode(txBytes []byte) (sdk.Tx, error)
	TxEncode(tx sdk.Tx) ([]byte, error)

	Info(*abci.RequestInfo) (*abci.ResponseInfo, error)
	Query(context.Context, *abci.RequestQuery) (*abci.ResponseQuery, error)

	ApplySnapshotChunk(req *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error)
	LoadSnapshotChunk(req *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error)
	OfferSnapshot(req *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error)
	ListSnapshots(req *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error)

	CheckTx(req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error)
	GetCheckStateContext() sdk.Context
}
