package keeper_test

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	cw8types "mythos/v1/x/wasmx/cw8/types"
	"mythos/v1/x/wasmx/types"
)

var (
	//go:embed testdata/cw8/simple_contract.wasm
	cwSimpleContract []byte

	//go:embed testdata/cw8/cw20_atomic_swap.wasm
	cw20_atomic_swap []byte

	// taken from cosmwasm/contracts/reflect
	//go:embed testdata/cw8/reflect-aarch64.wasm
	wasm_reflect []byte
)

type ReflectMsg struct {
	Msgs []cw8types.CosmosMsg `json:"msgs"`
}

type ReflectSubMsg struct {
	Msgs []cw8types.SubMsg `json:"msgs"`
}

type ReflectExecuteMsg struct {
	ReflectMsg ReflectMsg `json:"reflect_msg"`
}

type ReflectExecuteMsg2 struct {
	ReflectSubMsg ReflectSubMsg `json:"reflect_sub_msg,omitempty"`
}

func (suite *KeeperTestSuite) TestWasmxSimpleContract() {
	wasmbin := cwSimpleContract
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	expectedDeps := []string{types.CW_ENV_8}

	codeId := appA.StoreCode(sender, wasmbin, nil)
	codeInfo := appA.App.WasmxKeeper.GetCodeInfo(appA.Context(), codeId)
	s.Require().ElementsMatch(expectedDeps, codeInfo.Deps, "wrong deps")

	value := 2
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{}`)}, "cwSimpleContract", nil)

	data := []byte(`{"increase":{}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	value += 1

	keybz := []byte("counter")
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(fmt.Sprintf("%d", value), string(queryres))

	data = []byte(`{"value":{}}`)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(`{"value":3}`, string(qres))
}

func (suite *KeeperTestSuite) TestWasmxCwAtomicSwap() {
	wasmbin := cw20_atomic_swap
	sender := suite.GetRandomAccount()
	recipient := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	expectedDeps := []string{types.CW_ENV_8}

	codeId := appA.StoreCode(sender, wasmbin, nil)
	codeInfo := appA.App.WasmxKeeper.GetCodeInfo(appA.Context(), codeId)
	s.Require().ElementsMatch(expectedDeps, codeInfo.Deps, "wrong deps")

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{}`)}, "cwSimpleContract", nil)

	preimage := "983dbea1affedeee253d5921804d11ce119058ba35f397adc02f69162025a0d5"
	preimageBz, err := hex.DecodeString(preimage)
	s.Require().NoError(err)
	h := sha256.New()
	h.Write(preimageBz)
	hashBz := h.Sum(nil)
	hashHex := hex.EncodeToString(hashBz)
	coins := sdk.NewCoins(sdk.NewCoin(appA.Denom, sdk.NewInt(10000)))

	balance1, err := appA.App.BankKeeper.Balance(appA.Context(), &banktypes.QueryBalanceRequest{Address: sender.Address.String(), Denom: appA.Denom})
	s.Require().NoError(err)

	data := fmt.Sprintf(`{"create":{"id":"swap1","hash":"%s","recipient":"%s","expires":{"at_height":10000}}}`, hashHex, recipient.Address.String())
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(data)}, coins, nil)

	balance2, err := appA.App.BankKeeper.Balance(appA.Context(), &banktypes.QueryBalanceRequest{Address: sender.Address.String(), Denom: appA.Denom})
	s.Require().NoError(err)

	balanceContract, err := appA.App.BankKeeper.Balance(appA.Context(), &banktypes.QueryBalanceRequest{Address: contractAddress.String(), Denom: appA.Denom})
	s.Require().NoError(err)

	fmt.Println("--balance1--", balance1)
	fmt.Println("--balance2--", balance2)

	// s.Require().Equal(balance1.Balance.Sub(coins[0]), balance2.Balance)
	s.Require().Equal(coins[0].Amount, balanceContract.Balance.Amount)

	balance3, err := appA.App.BankKeeper.Balance(appA.Context(), &banktypes.QueryBalanceRequest{Address: recipient.Address.String(), Denom: appA.Denom})
	s.Require().NoError(err)
	fmt.Println("--balance3--", balance3)

	data = fmt.Sprintf(`{"release":{"id":"swap1","preimage":"%s"}}`, preimage)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(data)}, coins, nil)

	balanceContract, err = appA.App.BankKeeper.Balance(appA.Context(), &banktypes.QueryBalanceRequest{Address: contractAddress.String(), Denom: appA.Denom})
	s.Require().NoError(err)

	balance3, err = appA.App.BankKeeper.Balance(appA.Context(), &banktypes.QueryBalanceRequest{Address: recipient.Address.String(), Denom: appA.Denom})
	s.Require().NoError(err)
	fmt.Println("--balance3--", balance3)
	fmt.Println("--balance contract--", balanceContract)
	fmt.Println("--sender.Address--", sender.Address.String())
	fmt.Println("--contractAddress--", contractAddress.String())
	s.Require().Equal(balance2.Balance.Add(coins[0]), balance3.Balance)
}

func (suite *KeeperTestSuite) TestWasmxCwReflect() {
	wasmbin := wasm_reflect
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	expectedDeps := []string{types.CW_ENV_8}

	codeId := appA.StoreCode(sender, wasmbin, nil)
	codeInfo := appA.App.WasmxKeeper.GetCodeInfo(appA.Context(), codeId)
	s.Require().ElementsMatch(expectedDeps, codeInfo.Deps, "wrong deps")

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{}`)}, "wasm_reflect", nil)

	codeIdCounter := appA.StoreCode(sender, cwSimpleContract, nil)
	contractAddressCounter := appA.InstantiateCode(sender, codeIdCounter, types.WasmxExecutionMessage{Data: []byte(`{}`)}, "cwSimpleContract", nil)

	msgCounter := types.WasmxExecutionMessage{
		Data: []byte(`{"increase":{}}`),
	}
	msgbz, err := json.Marshal(msgCounter)
	s.Require().NoError(err)

	msgs := make([]cw8types.CosmosMsg, 1)
	msgs[0] = cw8types.CosmosMsg{
		Wasm: &cw8types.WasmMsg{
			Execute: &cw8types.ExecuteMsg{
				ContractAddr: contractAddressCounter.String(),
				Msg:          msgbz,
				Funds:        make(cw8types.Coins, 0),
			},
		},
	}
	msgsToReflect := ReflectExecuteMsg{ReflectMsg: ReflectMsg{Msgs: msgs}}
	msgsToReflectBz, err := json.Marshal(msgsToReflect)
	s.Require().NoError(err)

	qres := appA.WasmxQueryRaw(sender, contractAddressCounter, types.WasmxExecutionMessage{Data: []byte(`{"value":{}}`)}, nil, nil)
	suite.Require().Equal(`{"value":2}`, string(qres))

	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: msgsToReflectBz}, nil, nil)

	qres = appA.WasmxQueryRaw(sender, contractAddressCounter, types.WasmxExecutionMessage{Data: []byte(`{"value":{}}`)}, nil, nil)
	suite.Require().Equal(`{"value":3}`, string(qres))

	// SubMessages with Reply
	submsgs := make([]cw8types.SubMsg, 1)
	gasLimit := uint64(1000000)
	submsgs[0] = cw8types.SubMsg{
		ID:       2,
		ReplyOn:  cw8types.ReplyAlways,
		GasLimit: &gasLimit,
		Msg: cw8types.CosmosMsg{
			Wasm: &cw8types.WasmMsg{
				Execute: &cw8types.ExecuteMsg{
					ContractAddr: contractAddressCounter.String(),
					Msg:          msgbz,
					Funds:        make(cw8types.Coins, 0),
				},
			},
		},
	}
	msgsToReflect2 := ReflectExecuteMsg2{ReflectSubMsg: ReflectSubMsg{Msgs: submsgs}}
	msgsToReflectBz2, err := json.Marshal(msgsToReflect2)
	s.Require().NoError(err)

	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: msgsToReflectBz2}, nil, nil)

	qres = appA.WasmxQueryRaw(sender, contractAddressCounter, types.WasmxExecutionMessage{Data: []byte(`{"value":{}}`)}, nil, nil)
	suite.Require().Equal(`{"value":4}`, string(qres))

	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(`{"sub_msg_result":{"id":2}}`)}, nil, nil)
	expectedReply := cw8types.Reply{
		ID: 2,
		Result: cw8types.SubMsgResult{
			Ok: &cw8types.SubMsgResponse{
				Events: []cw8types.Event{{Type: "execute", Attributes: cw8types.EventAttributes{cw8types.EventAttribute{Key: "contract_address", Value: contractAddressCounter.String()}}}},
				Data:   []byte{10, 8, 0, 0, 0, 0, 0, 0, 0, 4},
			},
		},
	}
	expectedReplyBz, err := json.Marshal(expectedReply)
	s.Require().NoError(err)
	// expectedRes := fmt.Sprintf(`{"id":2,"result":{"ok":{"events":[{"type":"execute","attributes":[{"key":"contract_address","value":"%s"}]}],"data":"CggAAAAAAAAABA=="}}}`, contractAddressCounter.String())
	suite.Require().Equal(string(expectedReplyBz), string(qres))
}
