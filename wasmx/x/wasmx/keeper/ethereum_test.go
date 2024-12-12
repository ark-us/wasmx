package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"math/big"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/loredanacirstea/wasmx/v1/crypto/ethsecp256k1"
	testdata "github.com/loredanacirstea/wasmx/v1/x/wasmx/keeper/testdata/classic"
	"github.com/loredanacirstea/wasmx/v1/x/wasmx/types"
	cch "github.com/loredanacirstea/wasmx/v1/x/wasmx/types/contract_handler"
	"github.com/loredanacirstea/wasmx/v1/x/wasmx/types/contract_handler/alias"
	"github.com/loredanacirstea/wasmx/v1/x/wasmx/vm/precompiles"
)

func (suite *KeeperTestSuite) TestSendEthTx() {
	priv, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err)
	sender := sdk.AccAddress(priv.PubKey().Address().Bytes())
	initBalance := sdkmath.NewInt(1000_000_000)
	// getHex := `6d4ce63c`
	setHex := `60fe47b1`

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	evmcode, err := hex.DecodeString(testdata.SimpleStorage)
	s.Require().NoError(err)

	initvalue := "0000000000000000000000000000000000000000000000000000000000000009"
	initvaluebz, err := hex.DecodeString(initvalue)
	s.Require().NoError(err)

	databz := append(evmcode, initvaluebz...)
	res := appA.SendEthTx(priv, nil, databz, nil, uint64(1000000), big.NewInt(10000), nil)

	contractAddressStr := appA.GetContractAddressFromEvents(res.GetEvents())
	contractAddress, err := appA.AddressStringToAccAddressPrefixed(contractAddressStr)
	s.Require().NoError(err)

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, hex.EncodeToString(queryres))

	initvalue = "0000000000000000000000000000000000000000000000000000000000000006"
	databz = appA.Hex2bz(setHex + initvalue)
	to := types.EvmAddressFromAcc(contractAddress.Bytes())
	res = appA.SendEthTx(priv, &to, databz, nil, uint64(1000000), big.NewInt(10000), nil)

	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, hex.EncodeToString(queryres))
}

func (suite *KeeperTestSuite) TestAliasContract() {
	appA := s.AppContext()
	wasmbin := precompiles.GetPrecompileByLabel(appA.AddressCodec(), "alias_eth")
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)

	priv, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err)
	senderEth := sdk.AccAddress(priv.PubKey().Address().Bytes())
	senderEthHex := types.EvmAddressFromAcc(senderEth)
	senderHex := types.EvmAddressFromAcc(sender.Address)

	// "a074b0e1": "getCosmosAddress(address)",
	// "337e8247": "getRegisterHash(address,address)",
	// "2f6da104": "getRegisterMessage(address,address)",
	// "24b8fbf6": "register(address,bytes)"
	// registerHex := "24b8fbf6"
	registerHex := "4420e486"
	getCosmosAddressHex := "a074b0e1"
	getRegisterHash := "337e8247"
	// getRegisterMessage := "2f6da104"

	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	aliasAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "alias", nil)

	calld := getRegisterHash + senderHex.Hex()[2:] + senderEthHex.Hex()[2:]
	qres := appA.WasmxQueryRaw(sender, aliasAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)

	// calld = getRegisterMessage + senderHex.Hash().Hex()[2:] + senderEthHex.Hash().Hex()[2:]
	// msgData := appA.WasmxQueryRaw(sender, aliasAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	// msgHash := appA.SignHash191(msgData)
	// s.Require().Equal(msgHash.Hex()[2:], hex.EncodeToString(qres))

	signature := appA.SignEthMessage(priv, common.BytesToHash(qres))
	registerCalld := registerHex + senderEthHex.Hex()[2:] + hex.EncodeToString(signature)

	return // TODO

	appA.ExecuteContract(sender, aliasAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(registerCalld)}, nil, nil)

	calld = getCosmosAddressHex + senderEthHex.Hex()[2:]
	res := appA.WasmxQuery(sender, aliasAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	cosmosAddr := res[24:64]
	s.Require().Equal(strings.ToLower(senderHex.Hex()[2:]), cosmosAddr)
}

func (suite *KeeperTestSuite) TestAliasContractHandler() {
	sender := suite.GetRandomAccount()
	priv, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err)
	senderEth := sdk.AccAddress(priv.PubKey().Address().Bytes())
	senderEthHex := types.EvmAddressFromAcc(senderEth)
	senderHex := types.EvmAddressFromAcc(sender.Address)

	appA := s.AppContext()

	handler := appA.App.WasmxKeeper.ContractHandler()

	cwhandler := appA.App.WasmxKeeper.WasmVMResponseHandler()
	s.Require().NotNil(cwhandler)

	handler.Register(types.ROLE_ALIAS, alias.NewAliasHandler())
	msg := alias.RegisterRequest{EthAddress: senderEthHex, CoinType: uint32(60)}
	_, err = handler.Execute(appA.Context(), cch.ContractHandlerMessage{
		Role:   types.ROLE_ALIAS,
		Method: "register",
		Sender: sender.Address,
		Msg:    msg,
	})
	s.Require().NoError(err)

	querymsg := alias.GetCosmosAddressRequest{EthAddress: senderEthHex, CoinType: uint32(60)}
	resp, err := handler.Query(appA.Context(), cch.ContractHandlerMessage{
		Role:   types.ROLE_ALIAS,
		Method: "getCosmosAddress",
		Sender: sdk.AccAddress([]byte("modulename")),
		Msg:    querymsg,
	})
	s.Require().NoError(err)

	qres := resp.([]interface{})[0].(alias.GetCosmosAddressResponse)
	s.Require().True(qres.Found)
	s.Require().Equal(senderHex, qres.CosmAddress)
}

func (suite *KeeperTestSuite) TestAliasedAccount() {
	sender := suite.GetRandomAccount()
	// senderHex := types.EvmAddressFromAcc(sender.Address)
	priv, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err)
	senderEth := sdk.AccAddress(priv.PubKey().Address().Bytes())
	senderEthHex := types.EvmAddressFromAcc(senderEth)

	receiver := suite.GetRandomAccount()
	receiverHex := types.EvmAddressFromAcc(receiver.Address)
	// priv2, err := ethsecp256k1.GenerateKey()
	// s.Require().NoError(err)
	// receiverEth := sdk.AccAddress(priv2.PubKey().Address().Bytes())
	// receiverEthHex := types.EvmAddressFromAcc(receiverEth)

	initBalance := sdkmath.NewInt(1000_000_000_000)

	appA := s.AppContext()
	aliasEthAddr := appA.BytesToAccAddressPrefixed(appA.Hex2bz(types.ADDR_ALIAS_ETH))

	// We only fund sender
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	handler := alias.NewAliasHandler()
	// Alias sender
	msg := alias.RegisterRequest{EthAddress: senderEthHex, CoinType: uint32(60)}
	execMsg, err := handler.Encode(cch.ContractHandlerMessage{
		Method: "register",
		Msg:    msg,
	})
	s.Require().NoError(err)
	appA.ExecuteContract(sender, aliasEthAddr, *execMsg, nil, nil)
	suite.Commit()

	senderBalance := appA.App.BankKeeper.GetBalance(appA.Context(), sender.Address, appA.Chain.Config.BaseDenom)

	// Send coins with tx signed by alias account
	value := big.NewInt(100000)
	appA.SendEthTx(priv, &receiverHex, []byte{}, value, 300000, nil, nil)
	suite.Commit()

	receiverBalance := appA.App.BankKeeper.GetBalance(appA.Context(), receiver.Address, appA.Chain.Config.BaseDenom)
	s.Require().Equal(receiverBalance.Amount.BigInt(), value)

	senderBalance2 := appA.App.BankKeeper.GetBalance(appA.Context(), sender.Address, appA.Chain.Config.BaseDenom)
	diff := big.NewInt(0).Sub(senderBalance.Amount.BigInt(), value)
	s.Require().Equal(senderBalance2.Amount.BigInt(), diff)
	senderBalance = senderBalance2

	senderEthBalance := appA.App.BankKeeper.GetBalance(appA.Context(), senderEth, appA.Chain.Config.BaseDenom)
	s.Require().Equal(senderEthBalance.Amount.BigInt(), big.NewInt(0))

	// test contract
	// "f8b2cb4f": "getBalance(address)",
	// "40dcae86": "sendCoins(address)"
	evmcode, err := hex.DecodeString(testdata.MythosAliasTest)
	s.Require().NoError(err)
	_, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "MythosAliasTest", nil)
	appA.Faucet.Fund(appA.Context(), contractAddress, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	sendCoinsHex := "40dcae86"
	// getBalanceHex := "f8b2cb4f"

	return // TODO

	calld := sendCoinsHex + receiverHex.Hex()[2:]
	appA.SendEthTx(priv, nil, appA.Hex2bz(calld), big.NewInt(10), uint64(1000000), big.NewInt(10000), nil)
	valueSent := big.NewInt(10)

	receiverBalance2 := appA.App.BankKeeper.GetBalance(appA.Context(), receiver.Address, appA.Chain.Config.BaseDenom)
	diff = big.NewInt(0).Add(receiverBalance.Amount.BigInt(), valueSent)
	s.Require().Equal(diff, receiverBalance2.Amount.BigInt())

	senderBalance2 = appA.App.BankKeeper.GetBalance(appA.Context(), sender.Address, appA.Chain.Config.BaseDenom)
	diff = big.NewInt(0).Sub(senderBalance.Amount.BigInt(), value)
	s.Require().Equal(senderBalance2.Amount.BigInt(), diff)

	senderEthBalance = appA.App.BankKeeper.GetBalance(appA.Context(), senderEth, appA.Chain.Config.BaseDenom)
	s.Require().Equal(senderEthBalance.Amount.BigInt(), big.NewInt(0))

	// res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(sendCoinsHex)}, nil, nil)

	// Alias receiver
	// msg = alias.RegisterRequest{EthAddress: receiverEthHex, CoinType: uint32(60)}
	// execMsg, err = handler.Encode(cch.ContractHandlerMessage{
	// 	Method: "register",
	// 	Msg:    msg,
	// })
	// s.Require().NoError(err)
	// appA.ExecuteContract(receiver, aliasEthAddr, *execMsg, nil, nil)
	// suite.Commit()

	// receiverEthBalance, err := appA.App.BankKeeper.Balance(appA.Context(), &banktypes.QueryBalanceRequest{Address: appA.MustAccAddressToString(receiverEth), Denom: appA.Chain.Config.BaseDenom})
	// s.Require().NoError(err)
	// s.Require().Equal(receiverEthBalance.GetBalance().Amount.BigInt(), big.NewInt(0))

}
