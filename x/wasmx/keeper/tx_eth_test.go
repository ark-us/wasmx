package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"

	testdata "mythos/v1/x/wasmx/keeper/testdata/classic"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestSendEthTx() {
	priv, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err)
	sender := sdk.AccAddress(priv.PubKey().Address().Bytes())
	initBalance := sdk.NewInt(1000_000_000)
	// getHex := `6d4ce63c`
	setHex := `60fe47b1`

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	evmcode, err := hex.DecodeString(testdata.SimpleStorage)
	s.Require().NoError(err)

	initvalue := "0000000000000000000000000000000000000000000000000000000000000009"
	initvaluebz, err := hex.DecodeString(initvalue)
	s.Require().NoError(err)

	databz := append(evmcode, initvaluebz...)
	res := appA.SendEthTx(priv, nil, databz, uint64(1000000), big.NewInt(10000), nil)

	contractAddressStr := appA.GetContractAddressFromLog(res.GetLog())
	contractAddress := sdk.MustAccAddressFromBech32(contractAddressStr)

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, hex.EncodeToString(queryres))

	initvalue = "0000000000000000000000000000000000000000000000000000000000000006"
	databz = appA.Hex2bz(setHex + initvalue)
	to := wasmxtypes.EvmAddressFromAcc(contractAddress)
	res = appA.SendEthTx(priv, &to, databz, uint64(1000000), big.NewInt(10000), nil)

	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, hex.EncodeToString(queryres))
}
