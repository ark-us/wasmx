package keeper_test

// import (
// 	"encoding/hex"
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"strconv"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/require"
// 	"github.com/stretchr/testify/suite"

// 	"github.com/golang/protobuf/proto" //nolint

// 	"github.com/cosmos/cosmos-sdk/client/tx"
// 	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
// 	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/cosmos/cosmos-sdk/types/simulation"
// 	"github.com/cosmos/cosmos-sdk/types/tx/signing"
// 	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
// 	abci "github.com/tendermint/tendermint/abci/types"
// 	"github.com/tendermint/tendermint/libs/log"
// 	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
// 	db "github.com/tendermint/tm-db"

// 	"wasmx/app"
// 	"wasmx/x/wasmx/keeper"
// 	"wasmx/x/wasmx/types"
// )

// type KeeperTestSuite struct {
// 	suite.Suite

// 	ctx sdk.Context

// 	app *app.App
// 	h   *app.TestSupport
// 	// queryClient types.QueryClient
// 	// signer      keyring.Signer
// 	consAddress sdk.ConsAddress
// 	// validator   stakingtypes.Validator
// 	denom           string
// 	faucet          *keeper.TestFaucet
// 	ProposerAddress []byte
// 	deliverTx       func(tx abci.RequestDeliverTx) abci.ResponseDeliverTx
// }

// var s *KeeperTestSuite

// func TestKeeperTestSuite(t *testing.T) {
// 	s = new(KeeperTestSuite)
// 	suite.Run(t, s)
// }

// func (suite *KeeperTestSuite) SetupTest() {
// 	// suite.app = app.Setup(false)
// 	suite.SetupApp()
// }

// func (suite *KeeperTestSuite) SetupApp() {
// 	t := suite.T()

// 	suite.denom = "amyt"
// 	chainId := "mythos_7000-1"
// 	db := db.NewMemDB()
// 	gapp := app.New(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, app.DefaultNodeHome, 0, app.MakeEncodingConfig(), app.EmptyBaseAppOptions{})
// 	genesisState := app.NewDefaultGenesisState()

// 	app.SetupWithSingleValidatorGenTX(t, genesisState, chainId)

// 	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
// 	require.NoError(t, err)
// 	// Initialize the chain
// 	now := time.Now().UTC()
// 	consensusParams := app.DefaultConsensusParams
// 	consensusParams.Block.MaxGas = 500_000_000_000
// 	gapp.InitChain(
// 		abci.RequestInitChain{
// 			ChainId:         chainId,
// 			ConsensusParams: consensusParams,
// 			Time:            now,
// 			Validators:      []abci.ValidatorUpdate{},
// 			AppStateBytes:   stateBytes,
// 		},
// 	)
// 	gapp.Commit()
// 	suite.app = gapp
// 	suite.ProposerAddress = []byte{182, 108, 128, 84, 106, 186, 182, 110, 93, 95, 17, 148, 50, 158, 25, 187, 140, 206, 92, 21}

// 	now = time.Now().UTC()
// 	header := tmproto.Header{
// 		ChainID:         chainId,
// 		Height:          2,
// 		Time:            now,
// 		AppHash:         []byte("myAppHash"),
// 		ProposerAddress: suite.ProposerAddress,
// 	}
// 	gapp.BaseApp.BeginBlock(abci.RequestBeginBlock{Header: header})

// 	suite.ctx = gapp.BaseApp.NewContext(false, header)

// 	params := suite.app.EwasmKeeper.GetParams(suite.ctx)
// 	params.EnableEwasm = true
// 	suite.app.EwasmKeeper.SetParams(suite.ctx, params)

// 	suite.faucet = keeper.NewTestFaucet(t, suite.ctx, suite.app.EwasmKeeper.BankKeeper, types.ModuleName, sdk.NewCoin(suite.denom, sdk.NewInt(100_000_000_000)))
// }

// // Commit commits and starts a new block with an updated context.
// func (suite *KeeperTestSuite) Commit() {
// 	suite.CommitAfter(time.Second * 0)
// }

// // Commit commits a block at a given time.
// func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
// 	header := suite.ctx.BlockHeader()
// 	suite.app.EndBlock(abci.RequestEndBlock{Height: header.Height})
// 	_ = suite.app.Commit()

// 	header.Height += 1
// 	header.Time = header.Time.Add(t)
// 	suite.app.BeginBlock(abci.RequestBeginBlock{
// 		Header: header,
// 	})

// 	// update ctx
// 	meter := sdk.NewGasMeter(300_000_000_000)
// 	suite.ctx = suite.app.BaseApp.NewContext(false, header).WithBlockGasMeter(meter)
// }

// var DEFAULT_GAS_PRICE = "0.05amyt"
// var DEFAULT_GAS_LIMIT = uint64(15_000_000)

// func (s *KeeperTestSuite) prepareCosmosTx(account simulation.Account, msgs []sdk.Msg, gasLimit *uint64, gasPrice *string) []byte {
// 	encodingConfig := app.MakeEncodingConfig()
// 	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
// 	var parsedGasPrices sdk.DecCoins
// 	var err error

// 	if gasLimit != nil {
// 		txBuilder.SetGasLimit(*gasLimit)
// 	} else {
// 		txBuilder.SetGasLimit(DEFAULT_GAS_LIMIT)
// 	}

// 	if gasPrice != nil {
// 		parsedGasPrices, err = sdk.ParseDecCoins(*gasPrice)
// 	} else {
// 		parsedGasPrices, err = sdk.ParseDecCoins(DEFAULT_GAS_PRICE)
// 	}
// 	s.Require().NoError(err)
// 	feeAmount := parsedGasPrices.AmountOf("amyt").MulInt64(int64(DEFAULT_GAS_LIMIT)).RoundInt()

// 	fees := &sdk.Coins{{Denom: s.denom, Amount: feeAmount}}
// 	txBuilder.SetFeeAmount(*fees)
// 	err = txBuilder.SetMsgs(msgs...)
// 	s.Require().NoError(err)

// 	seq, err := s.app.EwasmKeeper.AccountKeeper.GetSequence(s.ctx, account.Address)
// 	s.Require().NoError(err)

// 	// First round: we gather all the signer infos. We use the "set empty
// 	// signature" hack to do that.
// 	sigV2 := signing.SignatureV2{
// 		PubKey: account.PubKey,
// 		Data: &signing.SingleSignatureData{
// 			SignMode:  encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
// 			Signature: nil,
// 		},
// 		Sequence: seq,
// 	}

// 	err = txBuilder.SetSignatures(sigV2)
// 	s.Require().NoError(err)

// 	// Second round: all signer infos are set, so each signer can sign.
// 	accNumber := s.app.EwasmKeeper.AccountKeeper.GetAccount(s.ctx, account.Address).GetAccountNumber()
// 	signerData := authsigning.SignerData{
// 		ChainID:       s.ctx.ChainID(),
// 		AccountNumber: accNumber,
// 		Sequence:      seq,
// 	}
// 	sigV2, err = tx.SignWithPrivKey(
// 		encodingConfig.TxConfig.SignModeHandler().DefaultMode(), signerData,
// 		txBuilder, account.PrivKey, encodingConfig.TxConfig,
// 		seq,
// 	)
// 	s.Require().NoError(err)

// 	err = txBuilder.SetSignatures(sigV2)
// 	s.Require().NoError(err)

// 	// bz are bytes to be broadcasted over the network
// 	bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
// 	s.Require().NoError(err)
// 	return bz
// }

// func (s *KeeperTestSuite) DeliverTx(account simulation.Account, msgs ...sdk.Msg) abci.ResponseDeliverTx {
// 	bz := s.prepareCosmosTx(account, msgs, nil, nil)
// 	req := abci.RequestDeliverTx{Tx: bz}
// 	res := s.app.BaseApp.DeliverTx(req)
// 	return res
// }

// func (s *KeeperTestSuite) SimulateTx(account simulation.Account, msgs ...sdk.Msg) (sdk.GasInfo, *sdk.Result) {
// 	bz := s.prepareCosmosTx(account, msgs, nil, nil)
// 	gasInfo, res, err := s.app.Simulate(bz)
// 	s.Require().NoError(err)
// 	return gasInfo, res
// }

// func (s *KeeperTestSuite) Query(account simulation.Account, contract sdk.AccAddress, queryData []byte, queryPath string) abci.ResponseQuery {
// 	req := abci.RequestQuery{Data: queryData, Path: queryPath}
// 	return s.app.BaseApp.Query(req)
// }

// // SmartQuery This will serialize the query message and submit it to the contract.
// // The response is parsed into the provided interface.
// // Usage: SmartQuery(addr, QueryMsg{Foo: 1}, &response)
// func (s *KeeperTestSuite) SmartQuery(contractAddr string, queryMsg interface{}, response interface{}) error {
// 	msg, err := json.Marshal(queryMsg)
// 	if err != nil {
// 		return err
// 	}

// 	req := types.QuerySmartContractStateRequest{
// 		Address:   contractAddr,
// 		QueryData: msg,
// 	}
// 	reqBin, err := proto.Marshal(&req)
// 	if err != nil {
// 		return err
// 	}

// 	// TODO: what is the query?
// 	res := s.app.Query(abci.RequestQuery{
// 		Path: "/cosmwasm.wasm.v1.Query/SmartContractState",
// 		Data: reqBin,
// 	})

// 	if res.Code != 0 {
// 		return fmt.Errorf("query failed: (%d) %s", res.Code, res.Log)
// 	}

// 	// unpack protobuf
// 	var resp types.QuerySmartContractStateResponse
// 	err = proto.Unmarshal(res.Value, &resp)
// 	if err != nil {
// 		return err
// 	}
// 	// unpack json content
// 	return json.Unmarshal(resp.Data, response)
// }

// func (s *KeeperTestSuite) EwasmQuery(account simulation.Account, contract sdk.AccAddress, queryData []byte, funds sdk.Coins) string {
// 	query := types.QuerySmartContractCallRequest{
// 		Sender:    account.Address.String(),
// 		Address:   contract.String(),
// 		QueryData: queryData,
// 		Funds:     funds,
// 	}
// 	bz, err := query.Marshal()
// 	s.Require().NoError(err)

// 	req := abci.RequestQuery{Data: bz, Path: "/cosmwasm.wasm.v1.Query/SmartContractCall"}
// 	abcires := s.app.BaseApp.Query(req)
// 	var resp types.QuerySmartContractCallResponse
// 	err = resp.Unmarshal(abcires.Value)
// 	s.Require().NoError(err)

// 	var data keeper.EwasmQueryResponse
// 	err = json.Unmarshal(resp.Data, &data)
// 	s.Require().NoError(err)
// 	return hex.EncodeToString(data.Data)
// }

// func (s *KeeperTestSuite) DeliverTxWithOpts(account simulation.Account, msg sdk.Msg, gasLimit uint64, gasPrice *string) abci.ResponseDeliverTx {
// 	bz := s.prepareCosmosTx(account, []sdk.Msg{msg}, &gasLimit, gasPrice)
// 	req := abci.RequestDeliverTx{Tx: bz}
// 	res := s.app.BaseApp.DeliverTx(req)
// 	return res
// }

// func (s *KeeperTestSuite) GetRandomAccount() simulation.Account {
// 	pk := ed25519.GenPrivKey()
// 	privKey := secp256k1.GenPrivKeyFromSecret(pk.GetKey().Seed())
// 	pubKey := privKey.PubKey()
// 	address := sdk.AccAddress(pubKey.Address())
// 	account := simulation.Account{
// 		PrivKey: privKey,
// 		PubKey:  pubKey,
// 		Address: address,
// 	}
// 	return account
// }

// func (s *KeeperTestSuite) StoreCode(sender simulation.Account, wasmbin []byte) uint64 {
// 	storeCodeMsg := &types.MsgStoreCode{
// 		Sender:       sender.Address.String(),
// 		WASMByteCode: wasmbin,
// 	}

// 	res := s.DeliverTx(sender, storeCodeMsg)
// 	s.Require().True(res.IsOK(), res.GetLog())
// 	s.Commit()

// 	codeId := s.GetCodeIdFromLog(res.GetLog())

// 	bytecode, err := s.app.EwasmKeeper.TwasmKeeper.GetByteCode(s.ctx, codeId)
// 	s.Require().NoError(err)
// 	s.Require().Equal(bytecode, wasmbin)
// 	return codeId
// }

// func (s *KeeperTestSuite) InstantiateCode(sender simulation.Account, codeId uint64, instantiateMsgStr string) sdk.AccAddress {
// 	instantiateMsg := []byte(instantiateMsgStr)
// 	instantiateCodeMsg := &types.MsgInstantiateContract{
// 		Sender: sender.Address.String(),
// 		CodeID: codeId,
// 		Label:  "test",
// 		Msg:    instantiateMsg,
// 	}
// 	res := s.DeliverTxWithOpts(sender, instantiateCodeMsg, 1000000, nil) // 135690
// 	s.Require().True(res.IsOK(), res.GetLog())
// 	s.Commit()
// 	contractAddressStr := s.GetContractAddressFromLog(res.GetLog())
// 	contractAddress := sdk.MustAccAddressFromBech32(contractAddressStr)
// 	return contractAddress
// }

// func (s *KeeperTestSuite) ExecuteContract(sender simulation.Account, contractAddress sdk.AccAddress, executeMsgStr string, funds sdk.Coins) abci.ResponseDeliverTx {
// 	return s.ExecuteContractWithOpts(sender, contractAddress, executeMsgStr, funds, 1500000, nil)
// }

// func (s *KeeperTestSuite) ExecuteContractWithOpts(sender simulation.Account, contractAddress sdk.AccAddress, executeMsgStr string, funds sdk.Coins, gasLimit uint64, gasPrice *string) abci.ResponseDeliverTx {
// 	res := s.ExecuteContractNoCheck(sender, contractAddress, executeMsgStr, funds, gasLimit, gasPrice)
// 	s.Require().True(res.IsOK(), res.GetLog())
// 	s.Require().NotContains(res.GetLog(), "failed to execute message", res.GetLog())
// 	s.Commit()
// 	return res
// }

// func (s *KeeperTestSuite) ExecuteContractNoCheck(sender simulation.Account, contractAddress sdk.AccAddress, executeMsgStr string, funds sdk.Coins, gasLimit uint64, gasPrice *string) abci.ResponseDeliverTx {
// 	executeMsgBz := []byte(executeMsgStr)
// 	executeMsg := &types.MsgExecuteContract{
// 		Sender:   sender.Address.String(),
// 		Contract: contractAddress.String(),
// 		Msg:      executeMsgBz,
// 		Funds:    funds,
// 	}
// 	return s.DeliverTxWithOpts(sender, executeMsg, gasLimit, gasPrice) // 135690
// }

// type Attribute struct {
// 	Key   string
// 	Value string
// }

// type Event struct {
// 	Type       string
// 	Attributes *[]Attribute
// }

// type Log struct {
// 	MsgIndex uint64
// 	Events   []Event
// }

// func (s *KeeperTestSuite) GetFromLog(logstr string, logtype string) *[]Attribute {
// 	var logs []Log
// 	err := json.Unmarshal([]byte(logstr), &logs)
// 	s.Require().NoError(err)
// 	for _, log := range logs {
// 		for _, ev := range log.Events {
// 			if ev.Type == logtype {
// 				return ev.Attributes
// 			}
// 		}
// 	}
// 	return nil
// }

// func (s *KeeperTestSuite) GetCodeIdFromLog(logstr string) uint64 {
// 	attrs := s.GetFromLog(logstr, "store_code")
// 	if attrs == nil {
// 		return 0
// 	}
// 	for _, attr := range *attrs {
// 		if attr.Key == "code_id" {
// 			ui64, err := strconv.ParseUint(attr.Value, 10, 64)
// 			s.Require().NoError(err)
// 			return ui64
// 		}
// 	}
// 	return 0
// }

// func (s *KeeperTestSuite) GetContractAddressFromLog(logstr string) string {
// 	attrs := s.GetFromLog(logstr, "instantiate")
// 	s.Require().NotNil(attrs)
// 	for _, attr := range *attrs {
// 		if attr.Key == "_contract_address" {
// 			return attr.Value
// 		}
// 	}
// 	s.Require().True(false, "no contract address found in log")
// 	return ""
// }
