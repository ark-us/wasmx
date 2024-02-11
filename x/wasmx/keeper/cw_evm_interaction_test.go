package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	aabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	testdata "mythos/v1/x/wasmx/keeper/testdata/classic"
	interfacesTestdata "mythos/v1/x/wasmx/keeper/testdata/interfaces"
	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm"
)

type AtomicSwapBalance struct {
	Addr   common.Address `json:"addr"`
	Amount *big.Int       `json:"amount"`
}

func (suite *KeeperTestSuite) TestProxyInterfacesPrecompile() {
	wasmbin := cw20_base
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)
	proxyAddressBz, err := hex.DecodeString("0000000000000000000000000000000000000025")
	s.Require().NoError(err)
	proxyAddress := sdk.AccAddress(proxyAddressBz)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	expectedDeps := []string{types.CW_ENV_8}

	metadata := types.CodeMetadata{
		Abi: interfacesTestdata.Erc20Str,
	}
	codeId := appA.StoreCodeWithMetadata(sender, wasmbin, nil, metadata)
	codeInfo := appA.App.WasmxKeeper.GetCodeInfo(appA.Context(), codeId)
	s.Require().ElementsMatch(expectedDeps, codeInfo.Deps, "wrong deps")

	instantiateMsg := CW20InstantiateMsg{
		Name:     "cw20",
		Symbol:   "TKN",
		Decimals: 18,
		InitialBalances: []Cw20Coin{
			{Address: sender.Address.String(), Amount: "10000000000000000"},
		},
	}
	calld, err := json.Marshal(instantiateMsg)
	s.Require().NoError(err)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: calld}, "cw20", nil)
	contractAddressEvm := types.EvmAddressFromAcc(contractAddress)

	// EvmToJson
	calldbz, err := hex.DecodeString("0000000000000000000000003defca2d10c7540621fd8ad553e7f987571b712d0000000000000000000000000000000000000000000000000000000000000064")
	s.Require().NoError(err)
	databz, err := vm.ProxyInterfacesAbi.Pack("EvmToJson", contractAddressEvm, "transfer", calldbz)
	s.Require().NoError(err)

	qres := appA.WasmxQueryRaw(sender, proxyAddress, types.WasmxExecutionMessage{Data: databz}, nil, nil)
	s.Require().Equal(`{"transfer":{"amount":"100","recipient":"mythos18hhu5tgsca2qvg0a3t248elesat3kufdvfudaj"}}`, string(qres))

	// EvmToJsonCall
	recipient := suite.GetRandomAccount()
	recipientHex := types.EvmAddressFromAcc(recipient.Address)
	calldbz, err = interfacesTestdata.Erc20Abi.Methods["transfer"].Inputs.Pack(recipientHex, big.NewInt(100))
	s.Require().NoError(err)
	databz, err = vm.ProxyInterfacesAbi.Pack("EvmToJsonCall", contractAddressEvm, "transfer", calldbz)
	s.Require().NoError(err)

	calld = []byte(fmt.Sprintf(`{"balance":{"address":"%s"}}`, sender.Address.String()))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: calld}, nil, nil)
	suite.Require().Equal(`{"balance":"10000000000000000"}`, string(qres))

	appA.ExecuteContract(sender, proxyAddress, types.WasmxExecutionMessage{Data: databz}, nil, nil)

	calld = []byte(fmt.Sprintf(`{"balance":{"address":"%s"}}`, recipient.Address.String()))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: calld}, nil, nil)
	suite.Require().Equal(`{"balance":"100"}`, string(qres))
}

func (suite *KeeperTestSuite) TestProxyInterfacesAtomicSwap() {
	wasmbin := cw20_base
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	sender2AddressEvm := types.EvmAddressFromAcc(sender2.Address)
	initBalance := sdkmath.NewInt(1000_000_000).MulRaw(1000000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), sender2.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	expectedDeps := []string{types.CW_ENV_8}

	// Deploy CW20
	metadata := types.CodeMetadata{
		Abi: interfacesTestdata.Erc20Str,
	}
	codeId := appA.StoreCodeWithMetadata(sender, wasmbin, nil, metadata)
	codeInfo := appA.App.WasmxKeeper.GetCodeInfo(appA.Context(), codeId)
	s.Require().ElementsMatch(expectedDeps, codeInfo.Deps, "wrong deps")

	instantiateMsg := CW20InstantiateMsg{
		Name:     "cw20",
		Symbol:   "TKN",
		Decimals: 18,
		InitialBalances: []Cw20Coin{
			{Address: sender.Address.String(), Amount: "10000000000000000"},
		},
	}
	calld, err := json.Marshal(instantiateMsg)
	s.Require().NoError(err)
	contractAddressCw20 := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: calld}, "cw20", nil)
	contractAddressCw20Evm := types.EvmAddressFromAcc(contractAddressCw20)

	// Deploy ERC20
	evmcodeErc20, err := hex.DecodeString(testdata.Erc20AtomicSwap)
	s.Require().NoError(err)
	initvaluebz, err := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000005746f6b656e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003544b4e0000000000000000000000000000000000000000000000000000000000")
	s.Require().NoError(err)
	_, contractAddressErc20 := appA.DeployEvm(sender, evmcodeErc20, types.WasmxExecutionMessage{Data: initvaluebz}, nil, "ERC20", nil)
	contractAddressErc20Evm := types.EvmAddressFromAcc(contractAddressErc20)

	// Mint Erc20 for sender2
	calld, err = interfacesTestdata.Erc20Abi.Pack("mint", sender2AddressEvm, big.NewInt(200))
	s.Require().NoError(err)
	appA.ExecuteContract(sender2, contractAddressErc20, types.WasmxExecutionMessage{Data: calld}, sdk.NewCoins(sdk.NewCoin(appA.Denom, sdkmath.NewInt(200))), nil)

	// Deploy AtomicSwap
	evmcodeSwap, err := hex.DecodeString(testdata.AtomicSwap)
	s.Require().NoError(err)
	_, contractAddressSwap := appA.DeployEvm(sender, evmcodeSwap, types.WasmxExecutionMessage{Data: []byte{}}, nil, "AtomicSwap", nil)
	atomicSwapAbi, err := aabi.JSON(strings.NewReader(testdata.AtomicSwapAbiStr))
	s.Require().NoError(err)

	// Register tokens with atomic swap contract
	calld, err = atomicSwapAbi.Pack("register", contractAddressCw20Evm, uint8(2))
	s.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddressSwap, types.WasmxExecutionMessage{Data: calld}, nil, nil)

	calld, err = atomicSwapAbi.Pack("register", contractAddressErc20Evm, uint8(1))
	s.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddressSwap, types.WasmxExecutionMessage{Data: calld}, nil, nil)

	// Approve AtomicSwap contract in erc20
	contractAddressSwapEvm := types.EvmAddressFromAcc(contractAddressSwap)
	calld, err = interfacesTestdata.Erc20Abi.Pack("approve", contractAddressSwapEvm, big.NewInt(200))
	s.Require().NoError(err)
	appA.ExecuteContract(sender2, contractAddressErc20, types.WasmxExecutionMessage{Data: calld}, nil, nil)

	// Approve AtomicSwap contract in cw20
	appA.ExecuteContract(sender, contractAddressCw20, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"increase_allowance":{"spender":"%s","amount":"200"}}`, contractAddressSwap.String()))}, nil, nil)

	// Create swap
	calld, err = atomicSwapAbi.Pack("create", "swap1", big.NewInt(0), []AtomicSwapBalance{
		{Addr: contractAddressCw20Evm, Amount: big.NewInt(100)},
	}, []AtomicSwapBalance{
		{Addr: contractAddressErc20Evm, Amount: big.NewInt(100)},
	})
	s.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddressSwap, types.WasmxExecutionMessage{Data: calld}, nil, nil)

	// Check CW20 balance
	calld = []byte(fmt.Sprintf(`{"balance":{"address":"%s"}}`, contractAddressSwap.String()))
	qres := appA.WasmxQueryRaw(sender, contractAddressCw20, types.WasmxExecutionMessage{Data: calld}, nil, nil)
	suite.Require().Equal(`{"balance":"100"}`, string(qres))

	// Fulfill swap
	calld, err = atomicSwapAbi.Pack("fulfill", "swap1")
	s.Require().NoError(err)
	appA.ExecuteContract(sender2, contractAddressSwap, types.WasmxExecutionMessage{Data: calld}, nil, nil)

	// Check CW20 balance AtomicSwap
	calld = []byte(fmt.Sprintf(`{"balance":{"address":"%s"}}`, contractAddressSwap.String()))
	qres = appA.WasmxQueryRaw(sender, contractAddressCw20, types.WasmxExecutionMessage{Data: calld}, nil, nil)
	suite.Require().Equal(`{"balance":"0"}`, string(qres))

	// Check CW20 balance sender2
	calld = []byte(fmt.Sprintf(`{"balance":{"address":"%s"}}`, sender2.Address.String()))
	qres = appA.WasmxQueryRaw(sender2, contractAddressCw20, types.WasmxExecutionMessage{Data: calld}, nil, nil)
	suite.Require().Equal(`{"balance":"100"}`, string(qres))

	// Check Erc20 balance sender
	senderAddressEvm := types.EvmAddressFromAcc(sender.Address)
	calld, err = interfacesTestdata.Erc20Abi.Pack("balanceOf", senderAddressEvm)
	s.Require().NoError(err)
	qres2 := appA.WasmxQuery(sender, contractAddressErc20, types.WasmxExecutionMessage{Data: calld}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000064", qres2)
}
