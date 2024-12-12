package app

import (
	"cosmossdk.io/x/tx/signing"

	networktypes "github.com/loredanacirstea/wasmx/v1/x/network/types"
)

func GetCustomSigners() []signing.CustomGetSigner {
	return []signing.CustomGetSigner{
		networktypes.ProvideExecuteAtomicTxGetSigners(),
	}
}
