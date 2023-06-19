package testutil

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	//nolint

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	govtypes1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	ibcgotesting "github.com/cosmos/ibc-go/v6/testing"

	app "mythos/v1/app"
	wasmxkeeper "mythos/v1/x/wasmx/keeper"
	"mythos/v1/x/wasmx/types"
)

var (
	// TestAccAddress defines a resuable bech32 address for testing purposes
	// TODO: update crypto.AddressHash() when sdk uses address.Module()
	// TestAccAddress = icatypes.GenerateAddress(sdk.AccAddress(crypto.AddressHash([]byte(icatypes.ModuleName))), ibcgotesting.FirstConnectionID, TestPortID)
	// TestOwnerAddress defines a reusable bech32 address for testing purposes
	TestOwnerAddress = "cosmos1fjx8p8uzx3h5qszqnwvelulzd659j8ua5qvaep"
	// TestPortID defines a reusable port identifier for testing purposes
	TestPortID, _ = icatypes.NewControllerPortID(TestOwnerAddress)
	// TestVersion defines a reusable interchainaccounts version string for testing purposes
	TestVersion = string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: ibcgotesting.FirstConnectionID,
		HostConnectionId:       ibcgotesting.FirstConnectionID,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))
)

type AppContext struct {
	S KeeperTestSuite

	App   *app.App
	Chain *ibcgotesting.TestChain

	// for generate test tx
	ClientCtx client.Context

	Denom  string
	Faucet *wasmxkeeper.TestFaucet
}

func (s AppContext) Context() sdk.Context {
	return s.Chain.GetContext()
}

func (suite *AppContext) RegisterInterTxAccount(endpoint *ibcgotesting.Endpoint, owner string) error {
	// types.New
	return nil
}

var DEFAULT_GAS_PRICE = "0.05amyt"
var DEFAULT_GAS_LIMIT = uint64(15_000_000)

func (s AppContext) prepareCosmosTx(account simulation.Account, msgs []sdk.Msg, gasLimit *uint64, gasPrice *string) []byte {
	encodingConfig := app.MakeEncodingConfig()
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	var parsedGasPrices sdk.DecCoins
	var err error

	if gasLimit != nil {
		txBuilder.SetGasLimit(*gasLimit)
	} else {
		txBuilder.SetGasLimit(DEFAULT_GAS_LIMIT)
	}

	if gasPrice != nil {
		parsedGasPrices, err = sdk.ParseDecCoins(*gasPrice)
	} else {
		parsedGasPrices, err = sdk.ParseDecCoins(DEFAULT_GAS_PRICE)
	}
	s.S.Require().NoError(err)
	feeAmount := parsedGasPrices.AmountOf("amyt").MulInt64(int64(DEFAULT_GAS_LIMIT)).RoundInt()

	fees := &sdk.Coins{{Denom: s.Denom, Amount: feeAmount}}
	txBuilder.SetFeeAmount(*fees)
	err = txBuilder.SetMsgs(msgs...)
	s.S.Require().NoError(err)

	seq, err := s.App.AccountKeeper.GetSequence(s.Context(), account.Address)
	s.S.Require().NoError(err)

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: account.PubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seq,
	}

	err = txBuilder.SetSignatures(sigV2)
	s.S.Require().NoError(err)

	// Second round: all signer infos are set, so each signer can sign.
	accNumber := s.App.AccountKeeper.GetAccount(s.Context(), account.Address).GetAccountNumber()
	signerData := authsigning.SignerData{
		ChainID:       s.Context().ChainID(),
		AccountNumber: accNumber,
		Sequence:      seq,
	}
	sigV2, err = tx.SignWithPrivKey(
		encodingConfig.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, account.PrivKey, encodingConfig.TxConfig,
		seq,
	)
	s.S.Require().NoError(err)

	err = txBuilder.SetSignatures(sigV2)
	s.S.Require().NoError(err)

	// bz are bytes to be broadcasted over the network
	bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.S.Require().NoError(err)
	return bz
}

func (s AppContext) DeliverTx(account simulation.Account, msgs ...sdk.Msg) abci.ResponseDeliverTx {
	bz := s.prepareCosmosTx(account, msgs, nil, nil)
	req := abci.RequestDeliverTx{Tx: bz}
	res := s.App.BaseApp.DeliverTx(req)
	return res
}

func (s AppContext) DeliverTxWithOpts(account simulation.Account, msg sdk.Msg, gasLimit uint64, gasPrice *string) abci.ResponseDeliverTx {
	bz := s.prepareCosmosTx(account, []sdk.Msg{msg}, &gasLimit, gasPrice)
	req := abci.RequestDeliverTx{Tx: bz}
	res := s.App.BaseApp.DeliverTx(req)
	return res
}

func (s AppContext) StoreCode(sender simulation.Account, wasmbin []byte, deps []string) uint64 {
	storeCodeMsg := &types.MsgStoreCode{
		Sender:   sender.Address.String(),
		ByteCode: wasmbin,
		Deps:     deps,
	}

	res := s.DeliverTx(sender, storeCodeMsg)
	s.S.Require().True(res.IsOK(), res.GetLog())
	s.S.Commit()

	codeId := s.GetCodeIdFromLog(res.GetLog())

	bytecode, err := s.App.WasmxKeeper.GetByteCode(s.Context(), codeId)
	s.S.Require().NoError(err)
	s.S.Require().Equal(bytecode, wasmbin)
	return codeId
}

func (s AppContext) StoreCodeWasmx1(sender simulation.Account, wasmbin []byte) uint64 {
	deps := []string{types.WASMX_WASMX_1}
	return s.StoreCode(sender, wasmbin, deps)
}

func (s AppContext) StoreCodeEwasmEnv1(sender simulation.Account, wasmbin []byte) uint64 {
	deps := []string{types.EWASM_ENV_1}
	return s.StoreCode(sender, wasmbin, deps)
}

func (s AppContext) Deploy(sender simulation.Account, code []byte, deps []string, instantiateMsg types.WasmxExecutionMessage, funds sdk.Coins, label string) (uint64, sdk.AccAddress) {
	msgbz, err := json.Marshal(instantiateMsg)
	s.S.Require().NoError(err)
	storeCodeMsg := &types.MsgDeployCode{
		Sender:   sender.Address.String(),
		ByteCode: code,
		Deps:     deps,
		Metadata: types.CodeMetadata{Name: "mycontract"},
		Msg:      msgbz,
		Funds:    funds,
		// Label:    label,
	}

	res := s.DeliverTx(sender, storeCodeMsg)
	s.S.Require().True(res.IsOK(), res.GetLog())
	s.S.Commit()

	codeId := s.GetCodeIdFromLog(res.GetLog())
	contractAddressStr := s.GetContractAddressFromLog(res.GetLog())
	contractAddress := sdk.MustAccAddressFromBech32(contractAddressStr)
	return codeId, contractAddress
}

func (s AppContext) DeployEvm(sender simulation.Account, evmcode []byte, initMsg types.WasmxExecutionMessage, funds sdk.Coins, label string) (uint64, sdk.AccAddress) {
	// return s.Deploy(sender, evmcode, []string{types.WASMX_WASMX_2, types.INTERPRETER_EVM_SHANGHAI}, initMsg, funds, label)
	return s.Deploy(sender, evmcode, []string{types.EWASM_ENV_1, types.INTERPRETER_EVM_SHANGHAI}, initMsg, funds, label)
}

func (s AppContext) InstantiateCode(sender simulation.Account, codeId uint64, instantiateMsg types.WasmxExecutionMessage, label string, funds sdk.Coins) sdk.AccAddress {
	msgbz, err := json.Marshal(instantiateMsg)
	s.S.Require().NoError(err)
	instantiateContractMsg := &types.MsgInstantiateContract{
		Sender: sender.Address.String(),
		CodeId: codeId,
		Label:  label,
		Msg:    msgbz,
		Funds:  funds,
	}
	res := s.DeliverTxWithOpts(sender, instantiateContractMsg, 5000000, nil)
	s.S.Require().True(res.IsOK(), res.GetLog())
	s.S.Commit()
	contractAddressStr := s.GetContractAddressFromLog(res.GetLog())
	contractAddress := sdk.MustAccAddressFromBech32(contractAddressStr)
	return contractAddress
}

func (s AppContext) ExecuteContract(sender simulation.Account, contractAddress sdk.AccAddress, executeMsg types.WasmxExecutionMessage, funds sdk.Coins, dependencies []string) abci.ResponseDeliverTx {
	return s.ExecuteContractWithGas(sender, contractAddress, executeMsg, funds, dependencies, 1500000, nil)
}

func (s AppContext) ExecuteContractWithGas(sender simulation.Account, contractAddress sdk.AccAddress, executeMsg types.WasmxExecutionMessage, funds sdk.Coins, dependencies []string, gasLimit uint64, gasPrice *string) abci.ResponseDeliverTx {
	res := s.ExecuteContractNoCheck(sender, contractAddress, executeMsg, funds, dependencies, gasLimit, gasPrice)
	s.S.Require().True(res.IsOK(), res.GetLog())
	s.S.Require().NotContains(res.GetLog(), "failed to execute message", res.GetLog())
	s.S.Commit()
	return res
}

func (s AppContext) ExecuteContractNoCheck(sender simulation.Account, contractAddress sdk.AccAddress, executeMsg types.WasmxExecutionMessage, funds sdk.Coins, dependencies []string, gasLimit uint64, gasPrice *string) abci.ResponseDeliverTx {
	msgbz, err := json.Marshal(executeMsg)
	s.S.Require().NoError(err)
	executeContractMsg := &types.MsgExecuteContract{
		Sender:       sender.Address.String(),
		Contract:     contractAddress.String(),
		Msg:          msgbz,
		Funds:        funds,
		Dependencies: dependencies,
	}
	return s.DeliverTxWithOpts(sender, executeContractMsg, gasLimit, gasPrice)
}

func (s AppContext) WasmxQuery(account simulation.Account, contract sdk.AccAddress, executeMsg types.WasmxExecutionMessage, funds sdk.Coins, dependencies []string) string {
	result := s.WasmxQueryRaw(account, contract, executeMsg, funds, dependencies)
	return hex.EncodeToString(result)
}

func (s AppContext) WasmxQueryRaw(account simulation.Account, contract sdk.AccAddress, executeMsg types.WasmxExecutionMessage, funds sdk.Coins, dependencies []string) []byte {
	msgbz, err := json.Marshal(executeMsg)
	s.S.Require().NoError(err)
	query := types.QuerySmartContractCallRequest{
		Sender:       account.Address.String(),
		Address:      contract.String(),
		QueryData:    msgbz,
		Funds:        funds,
		Dependencies: dependencies,
	}
	bz, err := query.Marshal()
	s.S.Require().NoError(err)

	req := abci.RequestQuery{Data: bz, Path: "/mythos.wasmx.v1.Query/SmartContractCall"}
	abcires := s.App.BaseApp.Query(req)
	var resp types.QuerySmartContractCallResponse
	err = resp.Unmarshal(abcires.Value)
	s.S.Require().NoError(err)

	var data types.WasmxQueryResponse
	err = json.Unmarshal(resp.Data, &data)
	s.S.Require().NoError(err, abcires)
	return data.Data
}

func (s AppContext) SubmitGovProposal(sender simulation.Account, content v1beta1.Content, deposit sdk.Coins) abci.ResponseDeliverTx {
	proposalMsg, err := v1beta1.NewMsgSubmitProposal(content, deposit, sender.Address)
	s.S.Require().NoError(err)
	resp := s.DeliverTx(sender, proposalMsg)
	s.S.Require().True(resp.IsOK(), resp.GetLog())
	s.S.Commit()
	return resp
}

func (s AppContext) PassGovProposal(valAccount, sender simulation.Account, content v1beta1.Content) {
	deposit := sdk.NewCoins(sdk.NewCoin(s.Denom, sdk.NewInt(1_000_000_000_000)))
	resp := s.SubmitGovProposal(sender, content, deposit)

	proposalId, err := s.GetProposalIdFromLog(resp.GetLog())
	s.S.Require().NoError(err)
	proposal, found := s.App.GovKeeper.GetProposal(s.Context(), proposalId)
	s.S.Require().True(found)
	s.S.Require().Equal(govtypes1.StatusVotingPeriod, proposal.Status)

	// msgs, err := s.ParseProposal(proposal)
	// s.S.Require().NoError(err)
	// msg3, ok := msgs[0].(*govtypes1.MsgExecLegacyContent)
	// s.S.Require().True(ok)
	// textProp, ok := msg3.Content.GetCachedValue().(*v1beta1.TextProposal)
	// s.S.Require().True(ok)
	// s.S.Require().Equal(content.GetTitle(), textProp.Title)
	// s.S.Require().Equal(content.GetDescription(), textProp.Description)

	voteMsg := v1beta1.NewMsgVote(valAccount.Address, proposalId, v1beta1.OptionYes)
	resp = s.DeliverTx(valAccount, voteMsg)
	s.S.Require().True(resp.IsOK(), resp.GetLog())
	s.S.Commit()

	votingParams := s.App.GovKeeper.GetVotingParams(s.Context())
	voteEnd := *votingParams.VotingPeriod + time.Hour
	s.S.CommitNBlocks(s.Chain, uint64(voteEnd.Seconds()/5))
	s.S.Commit()
}

func (s AppContext) ParseProposal(proposal govtypes1.Proposal) ([]sdk.Msg, error) {
	msgs := make([]sdk.Msg, len(proposal.Messages))
	for i, anyJSON := range proposal.Messages {
		var msg sdk.Msg
		err := s.App.AppCodec().UnpackAny(anyJSON, &msg)
		if err != nil {
			return msgs, err
		}
		msgs[i] = msg
	}
	return msgs, nil
}

func (s AppContext) Hex2bz(hexd string) []byte {
	if hexd[:2] == "0x" {
		hexd = hexd[2:]
	}
	bz, err := hex.DecodeString(hexd)
	s.S.Require().NoError(err)
	return bz
}

type Attribute struct {
	Key   string
	Value string
}

type Event struct {
	Type       string
	Attributes *[]Attribute
}

type Log struct {
	MsgIndex uint64
	Events   []Event
}

func (s AppContext) GetFromLog(logstr string, logtype string) *[]Attribute {
	var logs []Log
	err := json.Unmarshal([]byte(logstr), &logs)
	s.S.Require().NoError(err)
	for _, log := range logs {
		for _, ev := range log.Events {
			if ev.Type == logtype {
				return ev.Attributes
			}
		}
	}
	return nil
}

func (s AppContext) GetCodeIdFromLog(logstr string) uint64 {
	attrs := s.GetFromLog(logstr, "store_code")
	if attrs == nil {
		return 0
	}
	for _, attr := range *attrs {
		if attr.Key == "code_id" {
			ui64, err := strconv.ParseUint(attr.Value, 10, 64)
			s.S.Require().NoError(err)
			return ui64
		}
	}
	return 0
}

func (s AppContext) GetContractAddressFromLog(logstr string) string {
	attrs := s.GetFromLog(logstr, "instantiate")
	s.S.Require().NotNil(attrs)
	for _, attr := range *attrs {
		if attr.Key == "contract_address" {
			return attr.Value
		}
	}
	s.S.Require().True(false, "no contract address found in log")
	return ""
}

func (s AppContext) GetProposalIdFromLog(logstr string) (uint64, error) {
	attrs := s.GetFromLog(logstr, "submit_proposal")
	// attrs := s.getSubmitProposalFromLog(logstr)
	for _, attr := range *attrs {
		if attr.Key == "proposal_id" {
			val, err := strconv.ParseInt(attr.Value, 10, 64)
			if err != nil {
				return 0, err
			}
			return uint64(val), nil
		}
	}
	return 0, errors.New("not found")
}

func GetSdkEvents(events []abci.Event, evtype string) []sdk.Event {
	sdkEvents := make([]sdk.Event, len(events))
	for _, ev := range events {
		if ev.GetType() != evtype {
			continue
		}

		attributes := make([]sdk.Attribute, len(ev.Attributes))
		for _, attr := range ev.Attributes {
			attributes = append(attributes, sdk.NewAttribute(string(attr.Key), string(attr.Value)))
		}

		sdkev := sdk.NewEvent(ev.Type, attributes...)
		sdkEvents = append(sdkEvents, sdkev)
	}
	return sdkEvents
}
