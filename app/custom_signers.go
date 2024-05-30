package app

import (
	"cosmossdk.io/x/tx/signing"

	networktypes "mythos/v1/x/network/types"
)

func GetCustomSigners() []signing.CustomGetSigner {
	return []signing.CustomGetSigner{
		networktypes.ProvideExecuteAtomicTxGetSigners(),
	}
}
