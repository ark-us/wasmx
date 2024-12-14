package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cometbft/cometbft/libs/rand"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	mcfg "github.com/loredanacirstea/wasmx/config"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func TestGenesisState_Validate(t *testing.T) {
	addrCodec := mcodec.NewAccBech32Codec("myth", mcodec.NewAddressPrefixedFromAcc).(mcodec.AccBech32Codec)
	bootstrapAccount, err := addrCodec.BytesToString(sdk.AccAddress(rand.Bytes(address.Len)))
	require.NoError(t, err)
	feeCollector, err := addrCodec.BytesToString(authtypes.NewModuleAddress(mcfg.FEE_COLLECTOR))
	require.NoError(t, err)
	mintAddress, err := addrCodec.BytesToString(authtypes.NewModuleAddress("mint"))
	require.NoError(t, err)
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesisState(addrCodec, bootstrapAccount, feeCollector, mintAddress, 1, false, "{}"),
			valid:    true,
		},
		{
			desc:     "valid genesis state",
			genState: &types.GenesisState{},
			valid:    false,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
