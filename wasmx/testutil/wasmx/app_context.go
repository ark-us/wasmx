package wasmx

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	//nolint

	"github.com/cosmos/gogoproto/proto"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"

	address "cosmossdk.io/core/address"
	sdkerr "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	govtypes1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibcgotesting "github.com/cosmos/ibc-go/v8/testing"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	cryptoeth "github.com/ethereum/go-ethereum/crypto"

	app "github.com/loredanacirstea/wasmx/app"
	mcodec "github.com/loredanacirstea/wasmx/codec"
	"github.com/loredanacirstea/wasmx/crypto/ethsecp256k1"
	msrvcfg "github.com/loredanacirstea/wasmx/server/config"
	network "github.com/loredanacirstea/wasmx/x/network/keeper"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	wasmxutils "github.com/loredanacirstea/wasmx/x/wasmx/rpc/backend"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
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

type ProtoTxProvider interface {
	GetProtoTx() *txtypes.Tx
}

type AppContext struct {
	S *KeeperTestSuite

	App   *app.App
	Chain TestChain

	// for generate test tx
	ClientCtx client.Context
	Faucet    *TestFaucet

	FinalizeBlockHandle func(txs [][]byte) (*abci.ResponseFinalizeBlock, error)
}

func (s *AppContext) ABCIClient() *network.ABCIClient {
	ae := s.App.GetActionExecutor()
	abcicli := network.NewABCIClient(
		s.App,
		s.App.BaseApp,
		s.App.Logger(),
		s.App.GetNetworkKeeper(),
		config.DefaultConfig(),
		&msrvcfg.Config{},
		ae.(*network.ActionExecutor),
	)
	return abcicli.(*network.ABCIClient)
}

func (s *AppContext) AddressCodec() address.Codec {
	return s.App.AccountKeeper.AddressCodec()
}

func (s *AppContext) ValidatorAddressCodec() address.Codec {
	return s.App.AccountKeeper.ValidatorAddressCodec()
}

func (s *AppContext) ConsensusAddressCodec() address.Codec {
	return s.App.AccountKeeper.ConsensusAddressCodec()
}

func (s *AppContext) AccBech32Codec() mcodec.AccBech32Codec {
	return s.App.WasmxKeeper.AccBech32Codec()
}

func (s *AppContext) AddressStringToAccAddress(addr string) (sdk.AccAddress, error) {
	return s.AddressCodec().StringToBytes(addr)
}

func (s *AppContext) BytesToAccAddressPrefixed(addr []byte) mcodec.AccAddressPrefixed {
	return s.AccBech32Codec().BytesToAccAddressPrefixed(addr)
}

func (s *AppContext) AddressStringToAccAddressPrefixed(addr string) (mcodec.AccAddressPrefixed, error) {
	return s.AccBech32Codec().StringToAccAddressPrefixed(addr)
}

func (s *AppContext) AccAddressToString(addr sdk.AccAddress) (string, error) {
	return s.AddressCodec().BytesToString(addr)
}

func (s *AppContext) MustAccAddressToString(addr sdk.AccAddress) string {
	res, err := s.AddressCodec().BytesToString(addr)
	s.S.Require().NoError(err)
	return res
}

func (s *AppContext) MustAddressStringToAccAddress(addr string) sdk.AccAddress {
	res, err := s.AddressCodec().StringToBytes(addr)
	s.S.Require().NoError(err)
	return res
}

func (s *AppContext) ValidatorAddressToAccAddress(addr string) (sdk.AccAddress, error) {
	return s.ValidatorAddressCodec().StringToBytes(addr)
}

func (s *AppContext) ConsensusAddressToAccAddress(addr string) (sdk.AccAddress, error) {
	return s.ConsensusAddressCodec().StringToBytes(addr)
}

func (s *AppContext) Context() sdk.Context {
	return s.Chain.GetContext()
}

func (s *AppContext) RegisterInterTxAccount(endpoint *ibcgotesting.Endpoint, owner string) error {
	// types.New
	return nil
}

var DEFAULT_GAS_PRICE = "10amyt"
var DEFAULT_GAS_LIMIT = uint64(100_000_000)
var DEFAULT_BALANCE = int64(10_000_000_000)

func (s *AppContext) PrepareCosmosSdkTxBuilder(msgs []sdk.Msg, gasLimit *uint64, gasPrice *string, memo string) client.TxBuilder {
	txConfig := s.App.TxConfig()
	txBuilder := txConfig.NewTxBuilder()
	var parsedGasPrices sdk.DecCoins
	var err error

	if gasLimit != nil {
		txBuilder.SetGasLimit(*gasLimit)
	} else {
		txBuilder.SetGasLimit(DEFAULT_GAS_LIMIT)
	}
	txBuilder.SetMemo(memo)

	var feeAmount sdkmath.Int
	denom := s.Chain.Config.BaseDenom
	if gasPrice != nil {
		parsedGasPrices, err = sdk.ParseDecCoins(*gasPrice)
		feeAmount = parsedGasPrices[0].Amount.MulInt64(int64(DEFAULT_GAS_LIMIT)).RoundInt()
		denom = parsedGasPrices[0].Denom
	} else {
		parsedGasPrices, err = sdk.ParseDecCoins(DEFAULT_GAS_PRICE)
		feeAmount = parsedGasPrices.AmountOf("amyt").MulInt64(int64(DEFAULT_GAS_LIMIT)).RoundInt()
	}
	s.S.Require().NoError(err)

	fees := &sdk.Coins{{Denom: denom, Amount: feeAmount}}
	txBuilder.SetFeeAmount(*fees)
	err = txBuilder.SetMsgs(msgs...)
	s.S.Require().NoError(err)
	return txBuilder
}

func (s *AppContext) PrepareCosmosSdkTx(account simulation.Account, msgs []sdk.Msg, gasLimit *uint64, gasPrice *string, memo string) client.TxBuilder {
	txBuilder := s.PrepareCosmosSdkTxBuilder(msgs, gasLimit, gasPrice, memo)
	return s.SignCosmosSdkTx(txBuilder, account)
}

func (s *AppContext) SignCosmosSdkTx(txBuilder client.TxBuilder, account simulation.Account) client.TxBuilder {
	txConfig := s.App.TxConfig()
	accP, err := s.App.AccountKeeper.GetAccountPrefixed(s.Context(), s.BytesToAccAddressPrefixed(account.Address))
	s.S.Require().NoError(err)
	s.S.Require().NotNil(accP, "account must exist")
	seq := accP.GetSequence()

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: account.PubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode(txConfig.SignModeHandler().DefaultMode()),
			Signature: nil,
		},
		Sequence: seq,
	}

	err = txBuilder.SetSignatures(sigV2)
	s.S.Require().NoError(err)

	// Second round: all signer infos are set, so each signer can sign.
	acc, err := s.App.AccountKeeper.GetAccountPrefixed(s.Context(), s.BytesToAccAddressPrefixed(account.Address))
	s.S.Require().NoError(err)
	signerData := authsigning.SignerData{
		ChainID:       s.Context().ChainID(),
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      seq,
		PubKey:        account.PubKey,
		Address:       acc.String(),
	}
	sigV2, err = tx.SignWithPrivKey(
		s.Context().Context(),
		signing.SignMode(txConfig.SignModeHandler().DefaultMode()), signerData,
		txBuilder, account.PrivKey, txConfig,
		seq,
	)
	s.S.Require().NoError(err)

	err = txBuilder.SetSignatures(sigV2)
	s.S.Require().NoError(err)

	return txBuilder
}

func (s *AppContext) PrepareCosmosTx(account simulation.Account, msgs []sdk.Msg, gasLimit *uint64, gasPrice *string, memo string) []byte {
	txConfig := s.App.TxConfig()
	txBuilder := s.PrepareCosmosSdkTx(account, msgs, gasLimit, gasPrice, memo)

	// bz are bytes to be broadcasted over the network
	bz, err := txConfig.TxEncoder()(txBuilder.GetTx())
	s.S.Require().NoError(err)

	// print json wasmxTx
	// txx, err := txConfig.TxDecoder()(bz)
	// newtx, err := types.WasmxTxFromSdkTx(s.Chain.Codec, txx)
	// newtxbz, err := s.Chain.Codec.MarshalJSON(newtx)
	// fmt.Println("--newtxbz-", err, string(newtxbz))

	return bz
}

func (s *AppContext) getNonce(addr sdk.AccAddress) uint64 {
	nonce, err := s.App.AccountKeeper.GetSequence(s.Context(), addr)
	if err != nil {
		return uint64(0)
	}
	return nonce
}

func (s *AppContext) BuildEthTx(
	priv *ethsecp256k1.PrivKey,
	to *common.Address,
	data []byte,
	value *big.Int,
	gasLimit uint64,
	gasPrice *big.Int,
	accesses *ethtypes.AccessList,
) (*types.MsgExecuteEth, sdk.Coins, uint64) {
	chainID, err := types.ParseEvmChainID(s.Context().ChainID())
	s.S.Require().NoError(err)
	cfg := s.Chain.Config
	ethSigner := ethtypes.LatestSignerForChainID(chainID)
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.getNonce(from.Bytes())
	tx := &ethtypes.LegacyTx{
		To:       to,
		Nonce:    nonce,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
		Value:    value,
	}
	fees := sdk.NewCoins(sdk.NewCoin(cfg.BaseDenom, sdkmath.NewIntFromBigInt(getFee(gasPrice, gasLimit))))
	ppriv, err := priv.ToECDSA()
	s.S.Require().NoError(err)
	ethTx, err := ethtypes.SignNewTx(ppriv, ethSigner, tx)
	s.S.Require().NoError(err)
	bz, err := ethTx.MarshalBinary()
	s.S.Require().NoError(err)
	senderstr, err := s.AddressCodec().BytesToString(types.AccAddressFromEvm(from))
	s.S.Require().NoError(err)
	return &types.MsgExecuteEth{Data: bz, Sender: senderstr}, fees, gasLimit
}

func (s *AppContext) SignEthMessage(
	priv *ethsecp256k1.PrivKey,
	msgHash common.Hash,
) []byte {
	ppriv, err := priv.ToECDSA()
	s.S.Require().NoError(err)
	sig, err := cryptoeth.Sign(msgHash.Bytes(), ppriv)
	s.S.Require().NoError(err)
	return sig
}

func (s *AppContext) SignHash191(data []byte) common.Hash {
	msg := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(data)))
	msg = append(msg, data...)
	return cryptoeth.Keccak256Hash([]byte(msg))
}

func getFee(gasPrice *big.Int, gas uint64) *big.Int {
	gasLimit := new(big.Int).SetUint64(gas)
	if gasPrice == nil {
		gasPrice = big.NewInt(1)
	}
	return new(big.Int).Mul(gasPrice, gasLimit)
}

func (s *AppContext) prepareEthTx(
	priv cryptotypes.PrivKey,
	msg sdk.Msg,
	txFee sdk.Coins,
	gasLimit uint64,
) ([]byte, error) {
	txConfig := s.App.TxConfig()
	txBuilder := txConfig.NewTxBuilder()

	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, err
	}

	// Set the extension
	option, err := codectypes.NewAnyWithValue(&types.ExtensionOptionEthereumTx{})
	if err != nil {
		return nil, err
	}

	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "could not set extensions for Ethereum tx")
	}

	builder.SetExtensionOptions(option)
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(txFee)

	// bz are bytes to be broadcasted over the network
	bz, err := txConfig.TxEncoder()(txBuilder.GetTx())
	s.S.Require().NoError(err)
	return bz, nil
}

func (s *AppContext) SetMultiChainExtensionOptions(
	txBuilder client.TxBuilder,
	chainId string,
	index int32,
	txcount int32,
) (client.TxBuilder, error) {
	// Set the extension
	option, err := codectypes.NewAnyWithValue(&networktypes.ExtensionOptionMultiChainTx{ChainId: chainId, Index: index, TxCount: txcount})
	if err != nil {
		return nil, err
	}

	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "could not set extensions for Ethereum tx")
	}

	builder.SetExtensionOptions(option)
	return txBuilder, nil
}

func (s *AppContext) SetMultiChainAtomicExtensionOptions(
	txBuilder client.TxBuilder,
	chainIds []string,
	leaderChainId string,
) (client.TxBuilder, error) {
	// Set the extension
	option, err := codectypes.NewAnyWithValue(&networktypes.ExtensionOptionAtomicMultiChainTx{LeaderChainId: leaderChainId, ChainIds: chainIds})
	if err != nil {
		return nil, err
	}

	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "could not set extensions for atomic tx")
	}

	builder.SetExtensionOptions(option)
	return txBuilder, nil
}

func (s *AppContext) defaultFinalizeBlock(txs [][]byte) (*abci.ResponseFinalizeBlock, error) {
	req := &abci.RequestFinalizeBlock{
		Txs:                txs,
		Height:             s.App.LastBlockHeight() + 1,
		Time:               s.Chain.CurrentHeader.Time,
		Hash:               s.App.LastCommitID().Hash,
		ProposerAddress:    s.Chain.CurrentHeader.ProposerAddress,
		NextValidatorsHash: s.Chain.CurrentHeader.ValidatorsHash,
	}
	res, err := s.App.BaseApp.FinalizeBlock(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *AppContext) FinalizeBlock(txs [][]byte) (*abci.ResponseFinalizeBlock, error) {
	if s.FinalizeBlockHandle != nil {
		return s.FinalizeBlockHandle(txs)
	}
	return s.defaultFinalizeBlock(txs)
}

func (s *AppContext) DeliverEthTx(priv cryptotypes.PrivKey, msg sdk.Msg, txFee sdk.Coins, gasLimit uint64) (*abci.ExecTxResult, error) {
	bz, err := s.prepareEthTx(priv, msg, txFee, gasLimit)
	s.S.Require().NoError(err)
	txs := [][]byte{}
	txs = append(txs, bz)
	res, err := s.FinalizeBlock(txs)
	if err != nil {
		return nil, err
	}
	s.S.Require().Equal(len(res.TxResults), 1)
	return res.TxResults[0], nil
}

func (s *AppContext) SendEthTx(
	priv *ethsecp256k1.PrivKey,
	to *common.Address,
	data []byte,
	value *big.Int,
	gasLimit uint64,
	gasPrice *big.Int,
	accesses *ethtypes.AccessList,
) *abci.ExecTxResult {
	msg, txFee, gasLimit := s.BuildEthTx(priv, to, data, value, gasLimit, gasPrice, accesses)
	bz, err := s.prepareEthTx(priv, msg, txFee, gasLimit)
	s.S.Require().NoError(err)
	txs := [][]byte{}
	txs = append(txs, bz)
	resFin, err := s.FinalizeBlock(txs)
	s.S.Require().NoError(err)
	s.S.Require().Equal(len(resFin.TxResults), 1)
	res := resFin.TxResults[0]
	s.S.Require().True(res.IsOK(), res.GetEvents())
	s.S.Commit()
	return res
}

func (s *AppContext) DeliverTx(account simulation.Account, msgs ...sdk.Msg) (*abci.ExecTxResult, error) {
	bz := s.PrepareCosmosTx(account, msgs, nil, nil, "")
	txs := [][]byte{}
	txs = append(txs, bz)
	res, err := s.FinalizeBlock(txs)
	if err != nil {
		return nil, err
	}

	s.S.Require().Equal(1, len(res.TxResults))
	return res.TxResults[0], nil
}

func (s *AppContext) DeliverTxRaw(txbz []byte) (*abci.ExecTxResult, error) {
	res, err := s.FinalizeBlock([][]byte{txbz})
	if err != nil {
		return nil, err
	}

	s.S.Require().Equal(1, len(res.TxResults))
	return res.TxResults[0], nil
}

func (s *AppContext) DeliverTxWithOpts(account simulation.Account, msg sdk.Msg, memo string, gasLimit uint64, gasPrice *string) (*abci.ExecTxResult, error) {
	_gasLimit := &gasLimit
	if gasLimit == 0 {
		_gasLimit = nil
	}
	bz := s.PrepareCosmosTx(account, []sdk.Msg{msg}, _gasLimit, gasPrice, memo)
	txs := [][]byte{}
	txs = append(txs, bz)
	res, err := s.FinalizeBlock(txs)
	if err != nil {
		return nil, err
	}
	s.S.Require().Equal(len(res.TxResults), 1)
	return res.TxResults[0], nil
}

func (s *AppContext) SimulateTx(account simulation.Account, msgs ...sdk.Msg) (sdk.GasInfo, *sdk.Result, error) {
	bz := s.PrepareCosmosTx(account, msgs, nil, nil, "")
	return s.App.BaseApp.Simulate(bz)
}

func (s *AppContext) BroadcastTxAsync(account simulation.Account, msgs []sdk.Msg, gasLimit *uint64, gasPrice *string, memo string) (*abci.ExecTxResult, error) {
	bz := s.PrepareCosmosTx(account, msgs, gasLimit, gasPrice, memo)

	abciClient := network.NewABCIClient(s.App, s.App.BaseApp, s.App.Logger(), &s.App.NetworkKeeper, nil, nil, s.App.GetActionExecutor().(*network.ActionExecutor))

	_, err := abciClient.BroadcastTxAsync(context.TODO(), bz)
	if err != nil {
		return nil, err
	}
	commitres, err := s.S.CommitBlock()
	if err != nil {
		return nil, err
	}
	s.S.Require().Equal(1, len(commitres.TxResults))
	return commitres.TxResults[0], nil
}

func (s *AppContext) StoreCode(sender simulation.Account, wasmbin []byte, deps []string) uint64 {
	senderstr, err := s.AddressCodec().BytesToString(sender.Address)
	s.S.Require().NoError(err)
	storeCodeMsg := &types.MsgStoreCode{
		Sender:   senderstr,
		ByteCode: wasmbin,
		Deps:     deps,
	}
	res, err := s.DeliverTx(sender, storeCodeMsg)
	s.S.Require().NoError(err)
	s.S.Require().True(res.IsOK(), res.Log, res.GetEvents())
	s.S.Commit()

	codeId := s.GetCodeIdFromEvents(res.GetEvents())

	bytecode, err := s.App.WasmxKeeper.GetByteCode(s.Context(), codeId)
	s.S.Require().NoError(err)
	s.S.Require().Equal(len(wasmbin), len(bytecode), "stored code length mismatch")
	s.S.Require().Equal(wasmbin, bytecode)
	return codeId
}

func (s *AppContext) StoreCodeWithMetadata(sender simulation.Account, wasmbin []byte, deps []string, metadata types.CodeMetadata) uint64 {
	senderstr, err := s.AddressCodec().BytesToString(sender.Address)
	s.S.Require().NoError(err)
	storeCodeMsg := &types.MsgStoreCode{
		Sender:   senderstr,
		ByteCode: wasmbin,
		Deps:     deps,
		Metadata: metadata.ToProto(),
	}

	res, err := s.DeliverTx(sender, storeCodeMsg)
	s.S.Require().NoError(err)
	s.S.Require().True(res.IsOK(), res.GetEvents())
	s.S.Commit()

	codeId := s.GetCodeIdFromEvents(res.GetEvents())

	bytecode, err := s.App.WasmxKeeper.GetByteCode(s.Context(), codeId)
	s.S.Require().NoError(err)
	s.S.Require().Equal(bytecode, wasmbin)
	return codeId
}

func (s *AppContext) Deploy(sender simulation.Account, code []byte, deps []string, instantiateMsg types.WasmxExecutionMessage, funds sdk.Coins, label string, metadata *types.CodeMetadata) (uint64, mcodec.AccAddressPrefixed) {
	msgbz, err := json.Marshal(instantiateMsg)
	s.S.Require().NoError(err)
	if metadata == nil {
		metadata = &types.CodeMetadata{Name: "mycontract"}
	}
	senderstr, err := s.AddressCodec().BytesToString(sender.Address)
	s.S.Require().NoError(err)
	storeCodeMsg := &types.MsgDeployCode{
		Sender:   senderstr,
		ByteCode: code,
		Deps:     deps,
		Metadata: metadata.ToProto(),
		Msg:      msgbz,
		Funds:    funds,
		Label:    label,
	}

	res, err := s.DeliverTx(sender, storeCodeMsg)
	s.S.Require().NoError(err)
	s.S.Require().True(res.IsOK(), res.GetLog(), res.GetEvents())
	s.S.Commit()

	codeId := s.GetCodeIdFromEvents(res.GetEvents())
	contractAddressStr := s.GetContractAddressFromEvents(res.GetEvents())
	contractAddress, err := s.App.WasmxKeeper.AccBech32Codec().StringToAccAddressPrefixed(contractAddressStr)
	s.S.Require().NoError(err)
	return codeId, contractAddress
}

func (s *AppContext) DeployEvm(sender simulation.Account, evmcode []byte, initMsg types.WasmxExecutionMessage, funds sdk.Coins, label string, metadata *types.CodeMetadata) (uint64, mcodec.AccAddressPrefixed) {
	return s.Deploy(sender, evmcode, []string{types.INTERPRETER_EVM_SHANGHAI}, initMsg, funds, label, metadata)
}

func (s *AppContext) InstantiateCode(sender simulation.Account, codeId uint64, instantiateMsg types.WasmxExecutionMessage, label string, funds sdk.Coins) mcodec.AccAddressPrefixed {
	msgbz, err := json.Marshal(instantiateMsg)
	s.S.Require().NoError(err)
	senderstr, err := s.AddressCodec().BytesToString(sender.Address)
	s.S.Require().NoError(err)
	instantiateContractMsg := &types.MsgInstantiateContract{
		Sender: senderstr,
		CodeId: codeId,
		Label:  label,
		Msg:    msgbz,
		Funds:  funds,
	}
	res, err := s.DeliverTxWithOpts(sender, instantiateContractMsg, "", DEFAULT_GAS_LIMIT, nil)
	s.S.Require().NoError(err)
	s.S.Require().True(res.IsOK(), res.GetLog())
	s.S.Commit()
	contractAddressStr := s.GetContractAddressFromEvents(res.GetEvents())
	contractAddress, err := s.AddressStringToAccAddressPrefixed(contractAddressStr)
	s.S.Require().NoError(err)
	return contractAddress
}

func (s *AppContext) DecodeExecuteResponse(res *abci.ExecTxResult, msg interface{}) error {
	sdkmsg := &sdk.TxMsgData{}
	err := proto.Unmarshal(res.Data, sdkmsg)
	if err != nil {
		return err
	}
	anymsg := sdkmsg.MsgResponses[0]
	msgi, err := mcodec.AnyToSdkMsg(s.App.AppCodec(), anymsg)
	if err != nil {
		return err
	}
	msgii := msgi.(*types.MsgExecuteContractResponse)
	err = json.Unmarshal(msgii.Data, msg)
	if err != nil {
		return err
	}
	return nil
}

func (s *AppContext) ExecuteContract(sender simulation.Account, contractAddress mcodec.AccAddressPrefixed, executeMsg types.WasmxExecutionMessage, funds sdk.Coins, dependencies []string) *abci.ExecTxResult {
	return s.ExecuteContractWithGas(sender, contractAddress, executeMsg, funds, dependencies, 20000000, nil)
}

func (s *AppContext) ExecuteContractWithGas(sender simulation.Account, contractAddress mcodec.AccAddressPrefixed, executeMsg types.WasmxExecutionMessage, funds sdk.Coins, dependencies []string, gasLimit uint64, gasPrice *string) *abci.ExecTxResult {
	res, err := s.ExecuteContractNoCheck(sender, contractAddress, executeMsg, funds, dependencies, gasLimit, gasPrice)
	s.S.Require().NoError(err)
	s.S.Require().True(res.IsOK(), res.GetLog(), res.GetEvents())
	s.S.Require().NotContains(res.GetLog(), "failed to execute message", res.GetEvents())
	s.S.Commit()
	return res
}

func (s *AppContext) ExecuteContractNoCheck(sender simulation.Account, contractAddress mcodec.AccAddressPrefixed, executeMsg types.WasmxExecutionMessage, funds sdk.Coins, dependencies []string, gasLimit uint64, gasPrice *string) (*abci.ExecTxResult, error) {
	msgbz, err := json.Marshal(executeMsg)
	s.S.Require().NoError(err)
	senderstr, err := s.AddressCodec().BytesToString(sender.Address)
	s.S.Require().NoError(err)
	executeContractMsg := &types.MsgExecuteContract{
		Sender:       senderstr,
		Contract:     contractAddress.String(),
		Msg:          msgbz,
		Funds:        funds,
		Dependencies: dependencies,
	}
	return s.DeliverTxWithOpts(sender, executeContractMsg, "", gasLimit, gasPrice)
}

func (s *AppContext) ExecuteContractSimulate(sender simulation.Account, contractAddress mcodec.AccAddressPrefixed, executeMsg types.WasmxExecutionMessage, funds sdk.Coins, dependencies []string) (sdk.GasInfo, *sdk.Result, error) {
	msgbz, err := json.Marshal(executeMsg)
	s.S.Require().NoError(err)
	senderstr, err := s.AddressCodec().BytesToString(sender.Address)
	s.S.Require().NoError(err)
	executeContractMsg := &types.MsgExecuteContract{
		Sender:       senderstr,
		Contract:     contractAddress.String(),
		Msg:          msgbz,
		Funds:        funds,
		Dependencies: dependencies,
	}
	return s.SimulateTx(sender, executeContractMsg)
}

func (s *AppContext) QueryContract(account simulation.Account, contract mcodec.AccAddressPrefixed, msg []byte, funds sdk.Coins, dependencies []string) []byte {
	result := s.WasmxQueryRaw(account, contract, types.WasmxExecutionMessage{Data: msg}, funds, dependencies)
	return result
}

func (s *AppContext) WasmxQuery(account simulation.Account, contract mcodec.AccAddressPrefixed, executeMsg types.WasmxExecutionMessage, funds sdk.Coins, dependencies []string) string {
	result := s.WasmxQueryRaw(account, contract, executeMsg, funds, dependencies)
	return hex.EncodeToString(result)
}

func (s *AppContext) WasmxQueryRaw(account simulation.Account, contract mcodec.AccAddressPrefixed, executeMsg types.WasmxExecutionMessage, funds sdk.Coins, dependencies []string) []byte {
	abcires, err := s.WasmxQueryRawNoCheck(account, contract, executeMsg, funds, dependencies)
	s.S.Require().NoError(err)
	var resp types.QuerySmartContractCallResponse
	err = resp.Unmarshal(abcires.Value)
	s.S.Require().NoError(err)

	var data types.WasmxQueryResponse
	err = json.Unmarshal(resp.Data, &data)
	s.S.Require().NoError(err, abcires)
	return data.Data
}

func (s *AppContext) WasmxQueryRawNoCheck(sender simulation.Account, contractAddress mcodec.AccAddressPrefixed, executeMsg types.WasmxExecutionMessage, funds sdk.Coins, dependencies []string) (*abci.ResponseQuery, error) {
	msgbz, err := json.Marshal(executeMsg)
	s.S.Require().NoError(err)
	senderstr, err := s.AddressCodec().BytesToString(sender.Address)
	s.S.Require().NoError(err)
	query := types.QuerySmartContractCallRequest{
		Sender:       senderstr,
		Address:      contractAddress.String(),
		QueryData:    msgbz,
		Funds:        funds,
		Dependencies: dependencies,
	}
	bz, err := query.Marshal()
	s.S.Require().NoError(err)

	req := &abci.RequestQuery{Data: bz, Path: "/mythos.wasmx.v1.Query/SmartContractCall"}
	return s.App.BaseApp.Query(s.Context().Context(), req)
}

func (s *AppContext) WasmxQueryDebug(account simulation.Account, contract mcodec.AccAddressPrefixed, executeMsg types.WasmxExecutionMessage, funds sdk.Coins, dependencies []string) (string, []byte, error) {
	result, memorySnapshot, err := s.WasmxQueryDebugRaw(account, contract, executeMsg, funds, dependencies)
	return hex.EncodeToString(result), memorySnapshot, err
}

func (s *AppContext) WasmxQueryDebugRaw(sender simulation.Account, contract mcodec.AccAddressPrefixed, executeMsg types.WasmxExecutionMessage, funds sdk.Coins, dependencies []string) ([]byte, []byte, error) {
	msgbz, err := json.Marshal(executeMsg)
	s.S.Require().NoError(err)
	senderstr, err := s.AddressCodec().BytesToString(sender.Address)
	s.S.Require().NoError(err)
	query := types.QueryDebugContractCallRequest{
		Sender:       senderstr,
		Address:      contract.String(),
		QueryData:    msgbz,
		Funds:        funds,
		Dependencies: dependencies,
	}
	bz, err := query.Marshal()
	s.S.Require().NoError(err)

	req := &abci.RequestQuery{Data: bz, Path: "/mythos.wasmx.v1.Query/DebugContractCall"}
	abcires, err := s.App.BaseApp.Query(s.Context().Context(), req)
	if err != nil {
		fmt.Println("abcires", abcires)
		return nil, nil, err
	}
	var resp types.QueryDebugContractCallResponse
	err = resp.Unmarshal(abcires.Value)
	if err != nil {
		fmt.Println("abcires", abcires)
		return resp.Data, resp.MemorySnapshot, err
	}
	var data types.WasmxQueryResponse
	err = json.Unmarshal(resp.Data, &data)
	return data.Data, resp.MemorySnapshot, err
}

func (s *AppContext) SubmitGovProposal(
	sender simulation.Account,
	msgs []sdk.Msg,
	deposit sdk.Coins,
	metadata, title, summary string,
	expedited bool,
) *abci.ExecTxResult {
	senderstr, err := s.AddressCodec().BytesToString(sender.Address)
	s.S.Require().NoError(err)
	proposalMsg, err := govtypes1.NewMsgSubmitProposal(msgs, deposit, senderstr, metadata, title, summary, expedited)
	s.S.Require().NoError(err)
	resp, err := s.DeliverTx(sender, proposalMsg)
	s.S.Require().NoError(err)
	s.S.Require().True(resp.IsOK(), resp.GetLog(), resp.GetEvents())

	proposalId, err := s.GetProposalIdFromEvents(resp.GetEvents())
	s.S.Require().NoError(err)
	proposal, err := s.App.GovKeeper.Proposal(s.Context(), &govtypes1.QueryProposalRequest{ProposalId: proposalId})
	s.S.Require().NoError(err)
	s.S.Require().Equal(govtypes1.StatusVotingPeriod, proposal.Proposal.Status)

	s.S.Commit()
	return resp
}

func (s *AppContext) PassGovProposal(
	valAccount,
	sender simulation.Account,
	msgs []sdk.Msg,
	metadata, title, summary string,
	expedited bool,
) {
	deposit := sdk.NewCoins(sdk.NewCoin(s.Chain.Config.BaseDenom, sdkmath.NewInt(1_000_000_000_000)))
	resp := s.SubmitGovProposal(sender, msgs, deposit, metadata, title, summary, expedited)

	proposalId, err := s.GetProposalIdFromEvents(resp.GetEvents())
	s.S.Require().NoError(err)
	proposal, err := s.App.GovKeeper.Proposal(s.Context(), &govtypes1.QueryProposalRequest{ProposalId: proposalId})
	s.S.Require().NoError(err)
	s.S.Require().Equal(govtypes1.StatusVotingPeriod, proposal.Proposal.Status)
	var voteMsg sdk.Msg
	if !s.Chain.GovernanceContinuous {
		valstr, err := s.AddressCodec().BytesToString(valAccount.Address)
		s.S.Require().NoError(err)
		voteMsg = &govtypes1.MsgVote{
			ProposalId: proposalId,
			Voter:      valstr,
			Option:     govtypes1.OptionYes,
			Metadata:   "votemetadata",
		}
	} else {
		govAddr, err := s.App.WasmxKeeper.GetAddressOrRole(s.Context(), types.ROLE_GOVERNANCE)
		s.S.Require().NoError(err)
		valstr, err := s.AddressCodec().BytesToString(valAccount.Address)
		s.S.Require().NoError(err)

		msg1 := []byte(fmt.Sprintf(`{"DepositVote":{"proposal_id":%d,"option_id":0,"voter":"%s","amount":"0x10000","arbitrationAmount":"0x00","metadata":"votemetadata"}}`, proposalId, valstr))
		msg11 := types.WasmxExecutionMessage{Data: msg1}
		msgbz, err := json.Marshal(msg11)
		s.S.Require().NoError(err)
		voteMsg = &types.MsgExecuteContract{
			Sender:   valstr,
			Contract: govAddr.String(),
			Msg:      msgbz,
		}

		// vote two times, so we pass the threshold
		resp, err = s.DeliverTx(valAccount, voteMsg)
		s.S.Require().NoError(err)
		s.S.Require().True(resp.IsOK(), resp.GetLog(), resp.GetEvents())
		s.S.Commit()
	}

	resp, err = s.DeliverTx(valAccount, voteMsg)
	s.S.Require().NoError(err)
	s.S.Require().True(resp.IsOK(), resp.GetLog(), resp.GetEvents())
	s.S.Commit()

	params, err := s.App.GovKeeper.Params(s.Context(), &govtypes1.QueryParamsRequest{})
	s.S.Require().NoError(err)
	voteEnd := *params.Params.VotingPeriod //  + time.Hour
	s.S.CommitNBlocks_(s.Chain, uint64(voteEnd.Milliseconds()/500))
	s.S.Commit()

	// check proposal passed
	proposal, err = s.App.GovKeeper.Proposal(s.Context(), &govtypes1.QueryProposalRequest{ProposalId: proposalId})
	s.S.Require().NoError(err)
	// s.S.Require().Equal(govtypes1.StatusPassed, proposal.Proposal.Status, "gov proposal does not have status passed")
}

func (s *AppContext) ParseProposal(proposal govtypes1.Proposal) ([]sdk.Msg, error) {
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

func (s *AppContext) Hex2bz(hexd string) []byte {
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

func (s *AppContext) GetFromLog(logstr string, logtype string) *[]Attribute {
	var logs []Log
	err := json.Unmarshal([]byte(logstr), &logs)
	s.S.Require().NoError(err, "could not unmarshal logstr")
	for _, log := range logs {
		for _, ev := range log.Events {
			if ev.Type == logtype {
				return ev.Attributes
			}
		}
	}
	return nil
}

func (s *AppContext) GetCodeIdFromLog(logstr string) uint64 {
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

func (s *AppContext) GetContractAddressFromLog(logstr string) string {
	attrs := s.GetFromLog(logstr, types.EventTypeInstantiate)
	s.S.Require().NotNil(attrs)
	for _, attr := range *attrs {
		if attr.Key == types.AttributeKeyContractAddrCreated {
			return attr.Value
		}
	}
	s.S.Require().True(false, "no contract address found in log")
	return ""
}

func (s *AppContext) GetProposalIdFromLog(logstr string) (uint64, error) {
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

func (s *AppContext) PrintEvents(events []abci.Event) {
	for i, ev := range events {
		fmt.Println("-", i, "-", ev.String())
	}
}

func (s *AppContext) GetSdkEventsByType(events []abci.Event, evtype string) []abci.Event {
	newevs := make([]abci.Event, 0)
	for _, ev := range events {
		if ev.GetType() != evtype {
			continue
		}
		newevs = append(newevs, ev)
	}
	return newevs
}

func (s *AppContext) GetWasmxEvents(events []abci.Event) []abci.Event {
	wasmxlog := types.CustomContractEventPrefix + types.EventTypeWasmxLog
	return s.GetSdkEventsByType(events, wasmxlog)
}

func (s *AppContext) GetAttributeValueFromEvent(event abci.Event, attrkey string) string {
	for _, ev := range event.Attributes {
		if ev.Key == attrkey {
			return ev.Value
		}
	}
	return ""
}

func (s *AppContext) GetEventsByAttribute(events []abci.Event, attrkey string, attrvalue string) []abci.Event {
	newevs := make([]abci.Event, 0)
	evs := s.GetWasmxEvents(events)
	for _, ev := range evs {
		for _, attr := range ev.Attributes {
			if attr.Key == attrkey && attr.Value == attrvalue {
				newevs = append(newevs, ev)
			}
		}
	}
	return newevs
}

func (s *AppContext) GetEwasmEvents(events []abci.Event) []abci.Event {
	ewasmtype := "ewasm" // TODO LOG_TYPE_EWASM
	return s.GetEventsByAttribute(events, "type", ewasmtype)
}

func (s *AppContext) GetEwasmLogs(addressCodec address.Codec, events []abci.Event) ([]*ethtypes.Log, error) {
	evs := s.GetEwasmEvents(events)
	return wasmxutils.TxLogsFromEvents(addressCodec, evs)
}

func (s *AppContext) GetCodeIdFromEvents(events []abci.Event) uint64 {
	evs := s.GetSdkEventsByType(events, "store_code")
	s.S.Require().Equal(1, len(evs), "multiple store_code events")
	for _, attr := range evs[0].Attributes {
		if attr.Key == "code_id" {
			ui64, err := strconv.ParseUint(attr.Value, 10, 64)
			s.S.Require().NoError(err)
			return ui64
		}
	}
	return 0
}

func (s *AppContext) GetContractAddressFromEvents(events []abci.Event) string {
	evs := s.GetSdkEventsByType(events, types.EventTypeInstantiate)
	s.S.Require().Equal(1, len(evs), "multiple instantiate events")
	for _, attr := range evs[0].Attributes {
		if attr.Key == types.AttributeKeyContractAddrCreated {
			return attr.Value
		}
	}
	s.S.Require().True(false, "no contract address found in log")
	return ""
}

func (s *AppContext) GetProposalIdFromEvents(events []abci.Event) (uint64, error) {
	evs := s.GetSdkEventsByType(events, "submit_proposal")
	for _, ev := range evs {
		for _, attr := range ev.Attributes {
			if attr.Key == "proposal_id" {
				val, err := strconv.ParseInt(attr.Value, 10, 64)
				if err != nil {
					return 0, err
				}
				return uint64(val), nil
			}
		}
	}
	return 0, errors.New("proposal id not found")
}

func (s *AppContext) QueryDecode(respbz []byte) []byte {
	var qresp types.WasmxQueryResponse
	err := json.Unmarshal(respbz, &qresp)
	s.S.Require().NoError(err)
	return qresp.Data
}

// func signEthTx() {
// 	privkey, _ := ethsecp256k1.GenerateKey()
// 	ethPriv, err := privkey.ToECDSA()

// 	tx := ethtypes.NewTx(&ethtypes.AccessListTx{
// 		Nonce:    0,
// 		Data:     nil,
// 		To:       &suite.to,
// 		Value:    big.NewInt(10),
// 		GasPrice: big.NewInt(1),
// 		Gas:      21000,
// 	})
// 	tx, err := ethtypes.SignTx(tx, ethtypes.NewEIP2930Signer(suite.chainID), ethPriv)
// }

// // "github.com/cosmos/cosmos-sdk/crypto/keyring"
// func (msg *MsgEthereumTx) Sign(ethSigner ethtypes.Signer, keyringSigner keyring.Signer) error {
// 	from := msg.GetFrom()
// 	if from.Empty() {
// 		return fmt.Errorf("sender address not defined for message")
// 	}

// 	tx := msg.AsTransaction()
// 	txHash := ethSigner.Hash(tx)

// 	sig, _, err := keyringSigner.SignByAddress(from, txHash.Bytes())
// 	if err != nil {
// 		return err
// 	}

// 	tx, err = tx.WithSignature(ethSigner, sig)
// 	if err != nil {
// 		return err
// 	}

// 	return msg.FromEthereumTx(tx)
// }
