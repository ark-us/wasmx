package keeper

import (
	"bytes"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cometkvstore "github.com/cometbft/cometbft/abci/example/kvstore"
	comettypes "github.com/cometbft/cometbft/abci/types"
)

// GetValidators set the params
func (m msgServer) GetValidators_(ctx sdk.Context) (validators []comettypes.ValidatorUpdate) {
	itr, err := m.DB.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}
	for ; itr.Valid(); itr.Next() {
		// fmt.Println("-GetValidators-", itr.Key(), string(itr.Key()), itr.Value(), string(itr.Value()))
		if isValidatorTx(itr.Key()) {
			validator := new(comettypes.ValidatorUpdate)
			err := comettypes.ReadMessage(bytes.NewBuffer(itr.Value()), validator)
			if err != nil {
				panic(err)
			}
			validators = append(validators, *validator)
		}
	}
	if err = itr.Error(); err != nil {
		panic(err)
	}
	return
}

func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), cometkvstore.ValidatorPrefix)
}
