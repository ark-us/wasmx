package types

import (
	context "context"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

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

	Info(*abci.RequestInfo) (*abci.ResponseInfo, error)
	Query(context.Context, *abci.RequestQuery) (*abci.ResponseQuery, error)
}
