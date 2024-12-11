package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	mcodec "mythos/v1/codec"
	"mythos/v1/x/wasmx/types"
)

func TestGenesisState_Validate(t *testing.T) {
	addrCodec := mcodec.NewAccBech32Codec("myth", mcodec.NewAddressPrefixedFromAcc).(mcodec.AccBech32Codec)
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesisState(addrCodec, "feecollector", "mint", "bootstrap", 1, false, "{}"),
			valid:    true,
		},
		{
			desc:     "valid genesis state",
			genState: &types.GenesisState{

				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
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
