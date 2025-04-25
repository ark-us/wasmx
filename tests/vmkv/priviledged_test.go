package keeper_test

import (
	_ "embed"
	"encoding/json"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/loredanacirstea/mythos-tests/vmkv/testdata"
	mcodec "github.com/loredanacirstea/wasmx/codec"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	vmkv "github.com/loredanacirstea/wasmx/x/vmkv"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestPriviledgedAPI() {
	// test that contracts without a role requiring a protected host API, only use the mocked host API
	wasmbin := testdata.WasmxTestKvDB
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	// test that a contract with a priviledged API, without the proper role is executed using the Mocked API
	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "sqltest", nil)

	cmdQuery := &Calldata{Get: &vmkv.KvGetRequest{
		Id:  "conn1",
		Key: []byte{1, 2},
	}}
	data, err := json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	// response from mocked host
	suite.Require().Equal(`{"error":"","value":"null"}`, string(qres))

	// now set a role for this priviledged contract
	suite.registerRole("somerole", contractAddress, sender)

	cmdQuery = &Calldata{Get: &vmkv.KvGetRequest{
		Id:  "conn1",
		Key: []byte{1, 2},
	}}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(`{"error":"kv db connection not found","value":"null"}`, string(qres))
}

func (suite *KeeperTestSuite) registerRole(rolename string, contractAddress mcodec.AccAddressPrefixed, sender simulation.Account) {
	title := "Register " + rolename
	description := "Register " + rolename
	appA := s.AppContext()
	rolesAddr := appA.AccBech32Codec().BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_ROLES))

	valAccount := simulation.Account{
		PrivKey: s.Chain().SenderPrivKey,
		PubKey:  s.Chain().SenderPrivKey.PubKey(),
		Address: s.Chain().SenderAccount.GetAddress(),
	}
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(500000)
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(valAccount.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	// register new role
	msg := []byte(fmt.Sprintf(`{"SetRole":{"role":{"role":"%s","storage_type":0,"primary":0,"multiple":false,"labels":["%s"],"addresses":["%s"]}}}`, rolename, rolename, contractAddress))
	msgbz, err := json.Marshal(&types.WasmxExecutionMessage{Data: msg})
	s.Require().NoError(err)
	exec := &types.MsgExecuteContract{
		Sender:   appA.App.WasmxKeeper.GetAuthority(),
		Contract: rolesAddr.String(),
		Msg:      msgbz,
	}
	appA.PassGovProposal(valAccount, sender, []sdk.Msg{exec}, "", title, description, false)

	resp := appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), contractAddress)
	s.Require().Equal(rolename, resp)

	role := appA.App.WasmxKeeper.GetRoleByLabel(appA.Context(), rolename)
	s.Require().Equal(contractAddress.String(), role.Addresses[0])
	s.Require().Equal(rolename, role.Labels[0])
	s.Require().Equal(rolename, role.Role)
}
