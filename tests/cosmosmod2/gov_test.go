package keeper_test

import (
	"encoding/json"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simulation "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	wasmxtest "github.com/loredanacirstea/mythos-tests/testdata/wasmx"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestTendermintP2PToRaftToTendermintP2PMigration() {
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)
	chain := s.Chain()
	appA := s.AppContext()
	denom := appA.Chain.Config.BaseDenom
	valAccount := simulation.Account{
		PrivKey: chain.SenderPrivKey,
		PubKey:  chain.SenderPrivKey.PubKey(),
		Address: chain.SenderAccount.GetAddress(),
	}
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(denom, initBalance))
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender2.Address), sdk.NewCoin(denom, initBalance))
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(valAccount.Address), sdk.NewCoin(denom, initBalance))
	raftContract := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_CONSENSUS_RAFTP2P)
	raftContractBech32 := appA.BytesToAccAddressPrefixed(raftContract.Bytes())

	var msg1 []byte
	var err error

	msg1 = []byte(`{"getContextValue":{"key":"validatorNodesInfo"}}`)
	qresp, err := s.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_CONSENSUS,
		Contract: wasmxtypes.ROLE_CONSENSUS,
		Msg:      msg1,
	})
	s.Require().NoError(err)
	nodesInfo := appA.QueryDecode(qresp.Data)

	// Register contract role proposal
	newlabel := wasmxtypes.CONSENSUS_RAFTP2P
	title := "Register consensus"
	description := "Register consensus"
	authority := appA.MustAccAddressToString(authtypes.NewModuleAddress(wasmxtypes.ROLE_GOVERNANCE))

	msg := []byte(fmt.Sprintf(`{"SetContractForRoleGov":{"role":"consensus","label":"%s","contract_address":"%s","action_type":0}}`, newlabel, raftContractBech32.String()))
	msgbz, err := json.Marshal(&wasmxtypes.WasmxExecutionMessage{Data: msg})
	s.Require().NoError(err)
	proposal := &wasmxtypes.MsgExecuteContract{
		Sender:   authority,
		Contract: wasmxtypes.ROLE_ROLES,
		Msg:      msgbz,
	}
	appA.PassGovProposal(valAccount, sender, []sdk.Msg{proposal}, "", title, description, false)

	resp := appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), raftContractBech32)
	s.Require().Equal(newlabel, resp)

	role := appA.App.WasmxKeeper.GetRoleByLabel(appA.Context(), newlabel)
	s.Require().GreaterOrEqual(len(role.Addresses), 2)
	s.Require().Equal(len(role.Addresses), len(role.Labels))
	s.Require().Equal(raftContractBech32.String(), role.Addresses[role.Primary])
	s.Require().Equal(newlabel, role.Labels[role.Primary])
	s.Require().Equal("consensus", role.Role)

	time.Sleep(time.Second * 4)

	// check that the setup was done on the new contract
	state := chain.GetCurrentState(appA.Context())
	s.Require().Contains(state, "Leader.active")

	// Check each simulated node has the correct context:
	msg1 = []byte(`{"getContextValue":{"key":"validatorNodesInfo"}}`)
	qresp, err = s.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   raftContractBech32.String(),
		Contract: raftContractBech32.String(),
		Msg:      msg1,
	})
	s.Require().NoError(err)
	qrespbz := appA.QueryDecode(qresp.Data)
	s.Require().Equal(string(qrespbz), string(nodesInfo))

	msg1 = []byte(`{"getContextValue":{"key":"currentNodeId"}}`)
	qresp, err = s.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   raftContractBech32.String(),
		Contract: raftContractBech32.String(),
		Msg:      msg1,
	})
	s.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	s.Require().Equal(string(qrespbz), `0`)

	// make a transaction with the new consensus algorithm
	codeId := appA.StoreCode(sender, wasmxtest.WasmxSimpleStorage, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, wasmxtypes.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

	data := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	appA.ExecuteContract(sender, contractAddress, wasmxtypes.WasmxExecutionMessage{Data: data}, nil, nil)

	// migrate consensus back to tendermint
	tendermintContract := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_CONSENSUS_TENDERMINTP2P)
	tendermintContractBech32 := appA.BytesToAccAddressPrefixed(tendermintContract.Bytes())
	contractInfo, _, _, err := appA.App.WasmxKeeper.ContractInstance(appA.Context(), tendermintContractBech32)
	s.Require().NoError(err)

	// we need to instantiate a new contract from the tendermint codeId
	newlabel = wasmxtypes.CONSENSUS_TENDERMINTP2P + "2"
	initmsg := wasmxtypes.WasmxExecutionMessage{}
	err = json.Unmarshal([]byte(contractInfo.InitMessage), &initmsg)
	s.Require().NoError(err)
	tendermintContractBech32 = appA.InstantiateCode(sender, contractInfo.CodeId, initmsg, newlabel, nil)

	// Register contract role proposal
	title = "Register consensus"
	description = "Register consensus"

	msg = []byte(fmt.Sprintf(`{"SetContractForRoleGov":{"role":"consensus","label":"%s","contract_address":"%s","action_type":0}}`, newlabel, tendermintContractBech32.String()))
	msgbz, err = json.Marshal(&wasmxtypes.WasmxExecutionMessage{Data: msg})
	s.Require().NoError(err)
	proposal = &wasmxtypes.MsgExecuteContract{
		Sender:   authority,
		Contract: wasmxtypes.ROLE_ROLES,
		Msg:      msgbz,
	}
	appA.PassGovProposal(valAccount, sender, []sdk.Msg{proposal}, "", title, description, false)

	resp = appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), tendermintContractBech32)
	s.Require().Equal(newlabel, resp)

	role = appA.App.WasmxKeeper.GetRoleByLabel(appA.Context(), newlabel)
	s.Require().GreaterOrEqual(len(role.Addresses), 2)
	s.Require().Equal(len(role.Addresses), len(role.Labels))
	s.Require().Equal(tendermintContractBech32.String(), role.Addresses[role.Primary])
	s.Require().Equal(newlabel, role.Labels[role.Primary])
	s.Require().Equal("consensus", role.Role)

	time.Sleep(time.Second * 4)

	// check that the setup was done on the new contract
	state = chain.GetCurrentState(appA.Context())
	s.Require().Contains(state, "started.Validator")

	// Check each simulated node has the correct context:
	msg1 = []byte(`{"getContextValue":{"key":"validatorNodesInfo"}}`)
	qresp, err = s.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   tendermintContractBech32.String(),
		Contract: tendermintContractBech32.String(),
		Msg:      msg1,
	})
	s.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	s.Require().Equal(string(qrespbz), string(nodesInfo))

	msg1 = []byte(`{"getContextValue":{"key":"currentNodeId"}}`)
	qresp, err = s.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   tendermintContractBech32.String(),
		Contract: tendermintContractBech32.String(),
		Msg:      msg1,
	})
	s.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	s.Require().Equal(string(qrespbz), `0`)

	// make a transaction with the new consensus algorithm
	data = []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	appA.ExecuteContract(sender, contractAddress, wasmxtypes.WasmxExecutionMessage{Data: data}, nil, nil)
}
