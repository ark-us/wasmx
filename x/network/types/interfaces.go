package types

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BaseApp interface {
	Name() string
	CreateQueryContext(height int64, prove bool) (sdk.Context, error)
	CommitMultiStore() storetypes.CommitMultiStore
	GetContextForCheckTx(txBytes []byte) sdk.Context
	GetContextForFinalizeBlock(txBytes []byte) sdk.Context
}
