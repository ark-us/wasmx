package keeper_test

import (
	_ "embed"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	testdata "mythos/v1/x/wasmx/keeper/testdata/classic"
	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm/precompiles"
)

func (suite *KeeperTestSuite) TestDynamicInterpreter() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1_000_000_000_000_000_000)
	valAccount := simulation.Account{
		PrivKey: s.chainA.SenderPrivKey,
		PubKey:  s.chainA.SenderPrivKey.PubKey(),
		Address: s.chainA.SenderAccount.GetAddress(),
	}

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), valAccount.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	wasmbin := precompiles.GetPrecompileByLabel(types.INTERPRETER_EVM_SHANGHAI)
	codeId := appA.StoreCode(sender, wasmbin, nil)
	interpreterAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "newinterpreter", nil)

	newlabel := types.INTERPRETER_EVM_SHANGHAI + "2"

	// Register contract role proposal
	proposal := types.NewRegisterRoleProposal("Register interpreter", "Register interpreter", "interpreter", newlabel, interpreterAddress.String())

	appA.PassGovProposal(valAccount, sender, proposal)

	resp := appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), interpreterAddress)
	s.Require().Equal(newlabel, resp)

	role := appA.App.WasmxKeeper.GetRoleByLabel(appA.Context(), newlabel)
	s.Require().Equal(interpreterAddress.String(), role.ContractAddress)
	s.Require().Equal(newlabel, role.Label)
	s.Require().Equal("interpreter", role.Role)

	// use this interpreter to execute contract
	setHex := `60fe47b1`
	evmcode, err := hex.DecodeString(testdata.SimpleStorage)
	s.Require().NoError(err)

	initvalue := "0000000000000000000000000000000000000000000000000000000000000009"
	initvaluebz, err := hex.DecodeString(initvalue)
	s.Require().NoError(err)
	_, contractAddress := appA.Deploy(sender, evmcode, []string{newlabel}, types.WasmxExecutionMessage{Data: initvaluebz}, nil, "simpleStorage")

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, hex.EncodeToString(queryres))

	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(setHex + "0000000000000000000000000000000000000000000000000000000000000006")}, nil, nil)

	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", hex.EncodeToString(queryres))

}
