package keeper_test

import (
	_ "embed"
	"encoding/json"
	"fmt"

	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simulation "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"mythos/v1/app"
	mcfg "mythos/v1/config"
	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestMultiChainExec() {
	chainId := mcfg.MYTHOS_CHAIN_ID_TEST
	config, err := mcfg.GetChainConfig(chainId)
	s.Require().NoError(err)
	sender := suite.GetRandomAccount()
	newacc := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)
	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	bankAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_BANK)

	msg := fmt.Sprintf(`{"SendCoins":{"from_address":"%s","to_address":"%s","amount":[{"denom":"%s","amount":"0x1000"}]}}`, sender.Address.String(), newacc.Address.String(), config.BaseDenom)
	suite.broadcastMultiChainExec([]byte(msg), sender, bankAddress, chainId)

	qmsg := fmt.Sprintf(`{"GetBalance":{"address":"%s","denom":"%s"}}`, newacc.Address.String(), config.BaseDenom)
	res := suite.queryMultiChainCall(appA.App, []byte(qmsg), sender, bankAddress, chainId)

	balance := &banktypes.QueryBalanceResponse{}
	err = json.Unmarshal(res, balance)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(appA.Denom, sdkmath.NewInt(0x1000)), *balance.Balance)
	// TODO try again query client - this time with conn.defer() in the test
}

func (suite *KeeperTestSuite) queryMultiChainCall(mapp *app.App, msg []byte, sender simulation.Account, contractAddress sdk.AccAddress, chainId string) []byte {
	msgwrap := &wasmxtypes.WasmxExecutionMessage{Data: msg}
	msgbz, err := json.Marshal(msgwrap)
	suite.Require().NoError(err)
	multimsg := &types.QueryContractCallRequest{
		MultiChainId: chainId,
		Sender:       sender.Address.String(),
		Address:      contractAddress.String(),
		QueryData:    msgbz,
	}
	appA := s.AppContext()
	res, err := mapp.NetworkKeeper.ContractCall(appA.Context(), multimsg)
	suite.Require().NoError(err)

	wres := &wasmxtypes.WasmxExecutionMessage{}
	err = json.Unmarshal(res.Data, wres)
	suite.Require().NoError(err)
	return wres.Data
}

func (suite *KeeperTestSuite) broadcastMultiChainExec(msg []byte, sender simulation.Account, contractAddress sdk.AccAddress, chainId string) (*abci.ExecTxResult, error) {
	appA := s.AppContext()
	msgwrap := &wasmxtypes.WasmxExecutionMessage{Data: msg}
	msgbz, err := json.Marshal(msgwrap)
	suite.Require().NoError(err)
	msgexec := &wasmxtypes.MsgExecuteContract{
		Sender:   sender.Address.String(),
		Contract: contractAddress.String(),
		Msg:      msgbz,
	}
	msgAny, err := codectypes.NewAnyWithValue(msgexec)
	suite.Require().NoError(err)
	multimsg := &types.MsgMultiChainWrap{
		MultiChainId: chainId,
		Sender:       sender.Address.String(),
		Data:         msgAny,
	}
	return appA.BroadcastTxAsync(sender, multimsg)
}
