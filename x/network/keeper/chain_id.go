package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	config "mythos/v1/config"
)

func (k *Keeper) SetGlobalChainConfig(ctx sdk.Context, chainId string) error {
	// TODO get chain address prefixes from contract
	config.SetGlobalChainConfig(chainId)

	// find the proper baseapp

	return nil
}
