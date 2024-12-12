package types_test

import (
	"math/big"
	"testing"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	"github.com/stretchr/testify/require"
)

func TestChainIdValidate(t *testing.T) {
	for _, tc := range []struct {
		chainId    string
		evmChainId *big.Int
		valid      bool
	}{
		{
			chainId:    "mythos_8000-1",
			evmChainId: big.NewInt(8000),
			valid:      true,
		},
		{
			chainId:    "level0_0_1000-1",
			evmChainId: big.NewInt(1000),
			valid:      true,
		},
		{
			chainId:    "leveln_2_1000-1",
			evmChainId: big.NewInt(1000),
			valid:      true,
		},
		{
			chainId:    "chain0_1_10001-1",
			evmChainId: big.NewInt(10001),
			valid:      true,
		},
		{
			chainId:    "level11_11_900001-1",
			evmChainId: big.NewInt(900001),
			valid:      true,
		},
	} {
		t.Run(tc.chainId, func(t *testing.T) {
			valid := types.IsValidChainID(tc.chainId)
			require.Equal(t, tc.valid, valid)

			if valid {
				evmId, err := types.ParseEvmChainID(tc.chainId)
				require.NoError(t, err)
				require.Equal(t, tc.evmChainId, evmId)
			}
		})
	}
}
