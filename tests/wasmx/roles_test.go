package keeper_test

import (
	_ "embed"
	"encoding/hex"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	testdata "github.com/loredanacirstea/mythos-tests/testdata/classic"
	wasmxtest "github.com/loredanacirstea/mythos-tests/testdata/wasmx"
)

func (suite *KeeperTestSuite) TestEwasmCallToPriviledged() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)
	evmcode, err := hex.DecodeString(testdata.CallGeneral)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, contractAddress1 := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "callgeneralwasm1", nil)

	storagebin := wasmxtest.WasmxSimpleStorage
	codeId := appA.StoreCode(sender, storagebin, nil)
	contractAddress2 := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

	msg := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	// res := appA.ExecuteContract(sender, contractAddress2, types.WasmxExecutionMessage{Data: data}, nil, nil)

	// Execute nested calls
	// value := `0000000000000000000000000000000000000000000000000000000000000009`
	// data := `0000000000000000000000000000000000000000000000000000000000000002` + `000000000000000000000000` + hex.EncodeToString(contractAccount2.Bytes()) + `000000000000000000000000` + hex.EncodeToString(contractAccount3.Bytes()) + value

	data := `0000000000000000000000000000000000000000000000000000000000000001` + `000000000000000000000000` + hex.EncodeToString(contractAddress2.Bytes()) + hex.EncodeToString(msg)

	deps := []string{types.EvmAddressFromAcc(contractAddress1.Bytes()).Hex(), types.EvmAddressFromAcc(contractAddress2.Bytes()).Hex()}
	res := appA.ExecuteContract(sender, contractAddress1, types.WasmxExecutionMessage{Data: appA.Hex2bz(data)}, nil, deps)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000003", string(res.Data))

	queryres, err := appA.App.WasmxKeeper.QuerySmart(
		appA.Context(),
		contractAddress2,
		[]byte(`{"get":{"key":"hello"}}`),
	)
	suite.Require().NoError(err)
	suite.Require().Equal("", hex.EncodeToString(queryres))

	// 	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	// 	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAccount1, keybz)
	// 	suite.Require().Equal(value, hex.EncodeToString(queryres))

	// 	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAccount2, keybz)
	// 	suite.Require().Equal(value, hex.EncodeToString(queryres))

	// 	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAccount3, keybz)
	// 	suite.Require().Equal(value, hex.EncodeToString(queryres))

}

// func (suite *KeeperTestSuite) TestRolesCallCore() {
// 	sender := suite.GetRandomAccount()
// 	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)
// 	valAccount := simulation.Account{
// 		PrivKey: s.Chain().SenderPrivKey,
// 		PubKey:  s.Chain().SenderPrivKey.PubKey(),
// 		Address: s.Chain().SenderAccount.GetAddress(),
// 	}

// 	appA := s.AppContext()
// 	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
// 	suite.Commit()
// 	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(valAccount.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
// 	suite.Commit()

// 	wasmbin := precompiles.GetPrecompileByLabel(appA.AddressCodec(), types.INTERPRETER_EVM_SHANGHAI)
// 	codeId := appA.StoreCode(sender, wasmbin, nil)
// 	interpreterAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "newinterpreter", nil)

// 	newlabel := types.INTERPRETER_EVM_SHANGHAI + "2"

// 	// Register contract role proposal
// 	title := "Register interpreter"
// 	description := "Register interpreter"
// 	authority := appA.MustAccAddressToString(authtypes.NewModuleAddress(types.ROLE_GOVERNANCE))
// 	interpreterAddressStr := interpreterAddress.String()
// 	proposal := &types.MsgRegisterRole{Authority: authority, Title: title, Description: description, Role: "interpreter", Label: newlabel, ContractAddress: interpreterAddressStr}
// 	appA.PassGovProposal(valAccount, sender, []sdk.Msg{proposal}, "", title, description, false)

// 	resp := appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), interpreterAddress.Bytes())
// 	s.Require().Equal(newlabel, resp)

// 	role := appA.App.WasmxKeeper.GetRoleByLabel(appA.Context(), newlabel)
// 	s.Require().Equal(interpreterAddressStr, role.ContractAddress)
// 	s.Require().Equal(newlabel, role.Label)
// 	s.Require().Equal("interpreter", role.Role)

// 	// use this interpreter to execute contract
// 	setHex := `60fe47b1`
// 	evmcode, err := hex.DecodeString(testdata.SimpleStorage)
// 	s.Require().NoError(err)

// 	initvalue := "0000000000000000000000000000000000000000000000000000000000000009"
// 	initvaluebz, err := hex.DecodeString(initvalue)
// 	s.Require().NoError(err)
// 	_, contractAddress := appA.Deploy(sender, evmcode, []string{newlabel}, types.WasmxExecutionMessage{Data: initvaluebz}, nil, "simpleStorage", nil)

// 	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
// 	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
// 	suite.Require().Equal(initvalue, hex.EncodeToString(queryres))

// 	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(setHex + "0000000000000000000000000000000000000000000000000000000000000006")}, nil, nil)

// 	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
// 	suite.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", hex.EncodeToString(queryres))

// }
