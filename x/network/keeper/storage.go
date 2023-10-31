package keeper

import (
	"strings"

	cometkvstore "github.com/cometbft/cometbft/abci/example/kvstore"
)

func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), cometkvstore.ValidatorPrefix)
}
