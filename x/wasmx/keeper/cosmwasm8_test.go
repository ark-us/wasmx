package keeper_test

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"mythos/v1/x/wasmx/types"
)

var (
	//go:embed testdata/cw8/simple_contract.wasm
	cwSimpleContract []byte

	//go:embed testdata/cw8/cw20_atomic_swap.wasm
	cw20_atomic_swap []byte
)

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

	data = fmt.Sprintf(`{"release":{"id":"swap1","preimage":"%s"}}`, preimage)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(data)}, coins, nil)
}
