package keeper_test

import (
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/merkle"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	mcodec "wasmx/v1/codec"
	testdata "wasmx/v1/x/wasmx/keeper/testdata/taylor"
	"wasmx/v1/x/wasmx/types"
	vmtypes "wasmx/v1/x/wasmx/vm/types"
)

func (suite *KeeperTestSuite) TestInterpreterTaySimpleStorage() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	deps := []string{types.INTERPRETER_TAY}
	codeId := appA.StoreCode(sender, []byte(testdata.SimpleStorageTay), deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{}}`)}, "SimpleContractTay", nil)

	key := []byte("hello")
	data := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	value := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("sammy"), value)

	data = []byte(`{"get":{"key":"hello"}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("sammy"), resp)
}

func (suite *KeeperTestSuite) TestInterpreterTayERC20() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	deps := []string{types.INTERPRETER_TAY}
	codeId := appA.StoreCode(sender, []byte(testdata.ERC20Tay), deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{"name":"Token","symbol":"TKN","decimals":8}}`)}, "ERC20Tay", nil)

	data := []byte(`{"getName":{}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte(`{"name":"Token"}`), resp)

	data = []byte(`{"getSymbol":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte(`{"symbol":"TKN"}`), resp)

	data = []byte(`{"getDecimals":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte(`{"decimals":8}`), resp)

	data = []byte(`{"mint":{}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, senderPrefixed.String()))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`{"balance":"0x00000000000000000000000000000000000000000000000000000000000003e8"}`, string(resp))

	data = []byte(`{"totalSupply":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte(`{"supply":"0x00000000000000000000000000000000000000000000000000000000000003e8"}`), resp)

	data = []byte(`{"transfer":{"receiver":"level01w9g4qsddjlfajfl7vehgagvysrzec9flwknee2","value":"0x20"}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(`{"balanceOf":{"owner":"level01w9g4qsddjlfajfl7vehgagvysrzec9flwknee2"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte(`{"balance":"0x0000000000000000000000000000000000000000000000000000000000000020"}`), resp)
}

func (suite *KeeperTestSuite) TestInterpreterTayJson() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	deps := []string{types.INTERPRETER_TAY}
	codeId := appA.StoreCode(sender, []byte(testdata.JsonTestTay), deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{}}`)}, "JsonTestTay", nil)

	data := []byte(`{"add":{"set":{"key":"hello","value":"sammy"}}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`{"additional":{"obj":{"a":8,"b":9},"another":4},"set":{"key":"hello","value":"sammy"}}`, string(resp))

	data = []byte(`{"identity":[1, 2, 3]}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`[1,2,3]`, string(resp))

	data = []byte(`{"identity":{"set":{"key":"hello","value":"sammy", "somearr": [6, "aaa", 5], "somev": true}}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`{"set":{"key":"hello","value":"sammy","somearr":[6,"aaa",5],"somev":true}}`, string(resp))

	data = []byte(`{"identity":"ana"}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`"ana"`, string(resp))

	data = []byte(`{"identity":5}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`5`, string(resp))

	data = []byte(`{"identity":true}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`true`, string(resp))
}

func (suite *KeeperTestSuite) TestInterpreterTayOpcodes() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	senderP := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderP, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	deps := []string{types.INTERPRETER_TAY}
	codeId := appA.StoreCode(sender, []byte(testdata.OpcodesTay), deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{}}`)}, "OpcodesTay", nil)

	data := []byte(`{"getChainId":{}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("level0_1000-1", string(resp))

	data = []byte(`{"base64dec":{"value":"aGVsbG9vYQ=="}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("hellooa", string(resp))

	data = []byte(`{"base64dec":{"value":"aGVsbG8="}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("hello", string(resp))

	data = []byte(`{"base64enc":{"value":"hellooa"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("aGVsbG9vYQ==", string(resp))

	data = []byte(`{"base64enc":{"value":"hello"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("aGVsbG8=", string(resp))

	data = []byte(`{"sha256":{"value":"aGVsbG8="}}`) // hello
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8", hex.EncodeToString(resp))

	data = []byte(`{"bufferstringify":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`{"somebuf":"AQIDBAU="}`, string(resp))

	// data = []byte(`{"uintstringify":{}}`)
	// resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	// s.Require().Equal(`{"someu256":"AQIDBAU="}`, string(resp))

	data = []byte(`{"getCaller":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(senderP.String(), string(resp))

	data = []byte(`{"getAddress":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(contractAddress.String(), string(resp))

	data = []byte(`{"storageStore":{"key":"hello","value":"sammy"}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(`{"storageLoad":{"key":"hello"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("sammy", string(resp))

	data = []byte(`{"storageStore":{"key":"hella","value":"bart"}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(`{"storageLoadRange":{"startkey":"hella","endkey":"hello"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`("YmFydA==")`, string(resp))

	data = []byte(`{"storageLoadRange":{"startkey":"hella","endkey":""}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`("YmFydA==","c2FtbXk=")`, string(resp))

	data = []byte(`{"storageLoadRangePairs":{"startkey":"hella","endkey":"hello"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`[{"key":"aGVsbGE=","value":"YmFydA=="}]`, string(resp))

	data = []byte(`{"storageLoadRangePairs":{"startkey":"hella","endkey":""}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`[{"key":"aGVsbGE=","value":"YmFydA=="},{"key":"aGVsbG8=","value":"c2FtbXk="}]`, string(resp))

	data = []byte(`{"log":{"type":"newevent","data":"YmFydA==","topics":[]}}`)
	eres := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	evs := appA.GetWasmxEvents(eres.Events)
	s.Require().GreaterOrEqual(len(evs), 1)
	ev := appA.GetEventsByAttribute(evs, "type", "newevent")
	s.Require().Equal("0x62617274", appA.GetAttributeValueFromEvent(ev[0], "data"))

	// and getFinishData
	data = []byte(`{"setFinishData":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("hello", string(resp))

	data = []byte(`{"revert":{}}`)
	eres, err := appA.ExecuteContractNoCheck(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 5500000, nil)
	s.Require().NoError(err)
	s.Require().True(eres.IsErr())
	s.Require().Contains(eres.Log, "someerror")

	data = []byte(`{"humanize":{"value":"level01ghkpwkcyrfg87ys05atjmc66wln7xnrx5urcjm"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("level01ghkpwkcyrfg87ys05atjmc66wln7xnrx5urcjm", string(resp))

	addr1, err := mcodec.AccAddressPrefixedFromBech32("level01ghkpwkcyrfg87ys05atjmc66wln7xnrx5urcjm")
	s.Require().NoError(err)
	addrCodec := mcodec.NewBech32Codec("mythos", mcodec.NewAddressPrefixedFromAcc)
	equivaddr, err := addrCodec.BytesToString(addr1.Bytes())
	s.Require().NoError(err)
	data = []byte(fmt.Sprintf(`{"addr_equivalent":{"a":"level01ghkpwkcyrfg87ys05atjmc66wln7xnrx5urcjm","b":"%s"}}`, equivaddr))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("true", string(resp))

	data = []byte(`{"humanize_mc":{"a":"level01ghkpwkcyrfg87ys05atjmc66wln7xnrx5urcjm"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("level01ghkpwkcyrfg87ys05atjmc66wln7xnrx5urcjm", string(resp))
}

func (suite *KeeperTestSuite) TestInterpreterTay2Opcodes() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	senderP := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderP, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	deps := []string{types.INTERPRETER_TAY}
	codeId := appA.StoreCode(sender, []byte(testdata.OpcodesTay2), deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{}}`)}, "OpcodesTay2", nil)

	data := []byte(`{"LoggerInfo":{"a":"info: some message"}}`)
	appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(`{"LoggerError":{"a":"error: some message"}}`)
	appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(`{"LoggerDebug":{"a":"debug: some message"}}`)
	appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(`{"getAddressByRole":{"a":"bank"}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	addrprefix, err := appA.App.WasmxKeeper.GetAddressOrRole(appA.Context(), "bank")
	s.Require().NoError(err)
	s.Require().Equal(addrprefix.String(), string(resp))

	data = []byte(fmt.Sprintf(`{"getRoleByAddress":{"a":"%s"}}`, addrprefix.String()))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	role := appA.App.WasmxKeeper.GetRoleByContractAddress(appA.Context(), addrprefix)
	s.Require().Equal(role, string(resp))

	data = []byte(`{"emitCosmosEvents":[{"type":"newtype","attributes":[{"key":"somekey","value":"someval","index":true}]}]}`)
	eres := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	evs := appA.GetSdkEventsByType(eres.Events, "newtype")
	s.Require().GreaterOrEqual(len(evs), 1)
	s.Require().Equal("somekey", evs[0].Attributes[0].Key)
	s.Require().Equal("someval", evs[0].Attributes[0].Value)

	data = []byte(`{"MerkleHash":{"a":["aGVsbG8="]}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	slicebz, err := base64.StdEncoding.DecodeString("aGVsbG8=")
	s.Require().NoError(err)
	hashbz := merkle.HashFromByteSlices([][]byte{slicebz})
	s.Require().Equal(hex.EncodeToString(hashbz), hex.EncodeToString(resp))

	consprivkey := ed25519.GenPrivKey()
	conspubkey := consprivkey.PubKey()

	data = []byte(fmt.Sprintf(`{"ed25519PubToHex":{"pubkey":"%s"}}`, base64.StdEncoding.EncodeToString(conspubkey.Bytes())))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	hexAddr := ed25519.PubKey(conspubkey.Bytes()).Address()
	s.Require().Equal(strings.ToLower(hexAddr.String()), hex.EncodeToString(resp))

	data = []byte(fmt.Sprintf(`{"ed25519Sign":{"msg":"somemessage","privkey":"%s"}}`, base64.StdEncoding.EncodeToString(consprivkey)))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	data = []byte(fmt.Sprintf(`{"ed25519Verify":{"msg":"somemessage","pubkey":"%s","signature":"%s"}}`, base64.StdEncoding.EncodeToString(conspubkey.Bytes()), base64.StdEncoding.EncodeToString(resp)))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("true", string(resp))

	data = []byte(fmt.Sprintf(`{"getBalance":{"address":"%s","denom":"%s"}}`, senderP.String(), appA.Chain.Config.BaseDenom))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	bal := appA.App.BankKeeper.GetBalance(appA.Context(), sender.Address.Bytes(), appA.Chain.Config.BaseDenom)
	respcoin := &sdk.Coin{}
	err = appA.App.AppCodec().UnmarshalJSON(resp, respcoin)
	s.Require().NoError(err)
	s.Require().Equal(bal.Denom, respcoin.Denom)
	s.Require().Equal(bal.Amount, respcoin.Amount)

	internalmsg := types.WasmxExecutionMessage{Data: appA.Hex2bz("aa0000000000000000000000000000000000000000000000000000000077")}
	msgbz, err := json.Marshal(internalmsg)
	suite.Require().NoError(err)
	msg := &types.MsgExecuteContract{
		Sender:       appA.MustAccAddressToString(sender.Address),
		Contract:     appA.MustAccAddressToString(types.AccAddressFromHex(types.ADDR_IDENTITY)),
		Msg:          msgbz,
		Funds:        nil,
		Dependencies: nil,
	}
	txbz := appA.PrepareCosmosTx(sender, []sdk.Msg{msg}, nil, nil, "")
	data = []byte(fmt.Sprintf(`{"decodeCosmosTxToJson":{"value":"%s"}}`, base64.StdEncoding.EncodeToString(txbz)))
	txsdk, err := appA.App.TxConfig().TxDecoder()(txbz)
	s.Require().NoError(err)
	txjson, err := appA.App.TxConfig().TxJSONEncoder()(txsdk)
	s.Require().NoError(err)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(string(txjson), string(resp))

	data = []byte(fmt.Sprintf(`{"verifyCosmosTx":{"value":"%s"}}`, base64.StdEncoding.EncodeToString(txbz)))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("true", string(resp))

	codeId2 := appA.StoreCode(sender, []byte(testdata.SimpleStorageTay), deps)
	contractAddress2 := appA.InstantiateCode(sender, codeId2, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{}}`)}, "SimpleContractTay", nil)
	data = []byte(`{"set":{"key":"hello","value":"bella"}}`)
	callvalue := sdkmath.NewInt(0)
	callreq := vmtypes.SimpleCallRequestRaw{
		To:       contractAddress2.String(),
		Calldata: data,
		IsQuery:  false,
		Value:    &callvalue,
		GasLimit: big.NewInt(100000),
	}
	callreqstr, err := json.Marshal(callreq)
	s.Require().NoError(err)
	data = []byte(fmt.Sprintf(`{"call":{"value":%s}}`, callreqstr))
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(`{"get":{"key":"hello"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress2, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("bella", string(resp))

	data = []byte(fmt.Sprintf(`{"getAccount":{"a":"%s"}}`, senderP.String()))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	acc, err := appA.App.AccountKeeper.GetAccountPrefixed(appA.Context(), senderP)
	s.Require().NoError(err)
	respacc := &authtypes.BaseAccount{}
	err = appA.App.JSONCodec().UnmarshalJSON(resp, respacc)
	s.Require().NoError(err)
	s.Require().Equal(acc.GetAccountNumber(), respacc.GetAccountNumber())
	s.Require().Equal(acc.GetPubKey(), respacc.GetPubKey())
	s.Require().Equal(acc.GetSequence(), respacc.GetSequence())
	s.Require().Equal(acc.GetAddressPrefixed().String(), respacc.Address)

	data = []byte(fmt.Sprintf(`{"createAccount":{"codeid":%d,"msg":"","label":"newsimplestorage","funds":[]}}`, codeId2))
	eres = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	newaddr := appA.GetContractAddressFromEvents(eres.GetEvents())
	s.Require().Equal(45, len(newaddr))

	data = []byte(fmt.Sprintf(`{"create2Account":{"codeid":%d,"msg":"","label":"newsimplestorage","salt":"0x00000000000000000000000000000000000000000000000000000000000003e8","funds":[]}}`, codeId2))
	eres = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	newaddr = appA.GetContractAddressFromEvents(eres.GetEvents())
	s.Require().Equal(45, len(newaddr))

	// TODO
	// data = []byte(fmt.Sprintf(`{"createAccountInterpreted":{"bytecode":"%s","msg":"","label":"newsimplestorage","funds":[]}}`, base64.StdEncoding.EncodeToString([]byte(testdata.SimpleStorageTay))))
	// eres = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	// newaddr = appA.GetContractAddressFromEvents(eres.GetEvents())
	// s.Require().Equal(45, len(newaddr))

	// data = []byte(fmt.Sprintf(`{"create2AccountInterpreted":{"bytecode":"%s","msg":"","label":"newsimplestorage","funds":[]}}`, base64.StdEncoding.EncodeToString([]byte(testdata.SimpleStorageTay))))
	// eres = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	// newaddr = appA.GetContractAddressFromEvents(eres.GetEvents())
	// s.Require().Equal(45, len(newaddr))

	// lasth, err := appA.ABCIClient().LatestBlockHeight(appA.Context())
	// s.Require().NoError(err)
	// data = []byte(`{"getCurrentBlock":{}}`)
	// resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	// blockdata, err := appA.ABCIClient().Block(appA.Context(), &lasth)
	// s.Require().NoError(err)
	// blockdatastr, err := json.Marshal(blockdata)
	// s.Require().NoError(err)
	// s.Require().Equal(string(blockdatastr), string(resp))

	// blockheight := int64(2)
	// data = []byte(`{"getBlockHash":{"index":"0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	// resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	// blockdata, err = appA.ABCIClient().Block(appA.Context(), &blockheight)
	// s.Require().NoError(err)
	// s.Require().Equal(blockdata.BlockID.Hash, resp)
}

func (suite *KeeperTestSuite) TestInterpreterTayU256() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	senderP := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderP, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	deps := []string{types.INTERPRETER_TAY}
	codeId := appA.StoreCode(sender, []byte(testdata.OpcodesTay), deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{}}`)}, "OpcodesTay", nil)

	data := []byte(`{"add":{"a":"0x00000000000000000000000000000000000000000000000000000000000003e8","b": "0x00000000000000000000000000000000000000000000000000000000000003e8"}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("00000000000000000000000000000000000000000000000000000000000007d0", hex.EncodeToString(resp))

	data = []byte(`{"sub":{"a":"0x0000000000000000000000000000000000000000000000000000000000000003","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", hex.EncodeToString(resp))

	data = []byte(`{"mul":{"a":"0x0000000000000000000000000000000000000000000000000000000000000003","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", hex.EncodeToString(resp))

	data = []byte(`{"mul":{"a":"0x0000000000000000000000000ff0000000000000000000000000000000000003","b": "0x00000000000000000000000000000000000000000000000000000ff000000003"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("00000000000000fe010000002fd00000000000000000000000002fd000000009", hex.EncodeToString(resp))

	data = []byte(`{"div":{"a":"0x0000000000000000000000000000000000000000000000000000000000000009","b": "0x0000000000000000000000000000000000000000000000000000000000000003"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", hex.EncodeToString(resp))

	data = []byte(`{"div":{"a":"0x000000000000000000000000000000000000000000000000fffffffffffffff4","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000007ffffffffffffffa", hex.EncodeToString(resp))

	data = []byte(`{"lt":{"a":"0x0000000000000000000000000000000000000000000000000000000000000003","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("false", string(resp))

	data = []byte(`{"gt":{"a":"0x0000000000000000000000000000000000000000000000000000000000000003","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("true", string(resp))

	data = []byte(`{"eq":{"a":"0x0000000000000000000000000022222222222222222222222222222222222222","b": "0x0000000000000000000000000022222222222222222222222222222222222222"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("true", string(resp))

	data = []byte(`{"iszero":{"a":"0x0000000000000000000000000000000000000000000000000000000000000000"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("true", string(resp))

	data = []byte(`{"iszero":{"a":"0x0000000000000000000000000000000000000000000000000000000000000001"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("false", string(resp))

	data = []byte(`{"mod":{"a":"0x0000000000000000000000000000000000000000000000000000000000000008","b": "0x0000000000000000000000000000000000000000000000000000000000000003"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000002", hex.EncodeToString(resp))

	data = []byte(`{"not":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd", hex.EncodeToString(resp))

	data = []byte(`{"and":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002","b": "0x0000000000000000000000000000000000000000000000000000000000000003"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000002", hex.EncodeToString(resp))

	data = []byte(`{"or":{"a":"0xfffffff000000000000000000fffffffffffffffff0000000000f0f0f0ffffff","b": "0xff000ff00000000ffff00000ffffffff000000fffff0000000000f0f0f0fffff"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("fffffff00000000ffff00000fffffffffffffffffff000000000ffffffffffff", hex.EncodeToString(resp))

	data = []byte(`{"xor":{"a":"0xfffffff000000000000000000fffffffffffffffff0000000000f0f0f0ffffff","b": "0xff000ff00000000ffff00000ffffffff000000fffff0000000000f0f0f0fffff"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("00fff0000000000ffff00000f0000000ffffff0000f000000000fffffff00000", hex.EncodeToString(resp))

	data = []byte(`{"sgt":{"a":"0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000000", hex.EncodeToString(resp))

	data = []byte(`{"sgt":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002","b": "0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", hex.EncodeToString(resp))

	data = []byte(`{"slt":{"a":"0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", hex.EncodeToString(resp))

	data = []byte(`{"slt":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002","b": "0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000000", hex.EncodeToString(resp))

	data = []byte(`{"smod":{"a":"0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4","b": "0x0000000000000000000000000000000000000000000000000000000000000005"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe", hex.EncodeToString(resp))

	data = []byte(`{"shr":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002","b": "0x000000000000000000000000000000000000000000000000000000000000000c"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", hex.EncodeToString(resp))

	data = []byte(`{"shr":{"a":"0x0000000000000000000000000000000000000000000000000000000000000000","b": "0xaa0000000000000000000000000000000000000000000000000000000000000c"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("aa0000000000000000000000000000000000000000000000000000000000000c", hex.EncodeToString(resp))

	data = []byte(`{"shr":{"a":"0x0000000000000000000000000000000000000000000000000000000000000110","b": "0xc84a6e6ec1e7f30f5c812eeba420f76900000000000000000000000000000000"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000000", hex.EncodeToString(resp))

	data = []byte(`{"shr":{"a":"0x0000000000000000000000000000000000000000000000000000000000000080","b": "0xc84a6e6ec1e7f30f5c812eeba420f76900000000000000000000000000000000"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("00000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769", hex.EncodeToString(resp))

	data = []byte(`{"shr":{"a":"0x0000000000000000000000000000000000000000000000000000000000000100","b": "0xc84a6e6ec1e7f30f5c812eeba420f76900000000000000000000000000000000"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000000", hex.EncodeToString(resp))

	data = []byte(`{"shr":{"a":"0x00000000000000000000000000000000000000000000000000000000000000aa","b": "0x983dbea1affedeee253d5921804d11ce119058ba35f397adc02f69162025a0d5"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("000000000000000000000000000000000000000000260f6fa86bffb7bb894f56", hex.EncodeToString(resp))

	data = []byte(`{"shl":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002","b": "0x000000000000000000000000000000000000000000000000000000000000000c"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000030", hex.EncodeToString(resp))

	data = []byte(`{"shl":{"a":"0x00000000000000000000000000000000000000000000000000000000000000aa","b": "0x983dbea1affedeee253d5921804d11ce119058ba35f397adc02f69162025a0d5"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("ce5eb700bda45880968354000000000000000000000000000000000000000000", hex.EncodeToString(resp))

	data = []byte(`{"sar":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002","b": "0x000000000000000000000000000000000000000000000000000000000000000c"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", hex.EncodeToString(resp))

	data = []byte(`{"sar":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002","b": "0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd", hex.EncodeToString(resp))

	data = []byte(`{"addmod":{"a":"0x0000000000000000000000000000000000000000000000000000000000000005","b": "0x0000000000000000000000000000000000000000000000000000000000000003","c":"0x0000000000000000000000000000000000000000000000000000000000000003"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000002", hex.EncodeToString(resp))

	data = []byte(`{"exp":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002","b": "0x0000000000000000000000000000000000000000000000000000000000000003"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000008", hex.EncodeToString(resp))

	data = []byte(`{"exp":{"a":"0x000000000000000000000000000000000bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0089abcdf0123456789abcdf01234567787654320fedcba987654320fedcba99", hex.EncodeToString(resp))

	data = []byte(`{"exp":{"a":"0x0000000000000000000000000000000000000000000000000000000000bbbbbb","b": "0x000000000000000000000000000000000000000000000000000000000000000a"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("00000b83c5b522845b1672474f0defd175b8e0e028b86fe2788ae9dfd2659779", hex.EncodeToString(resp))

	data = []byte(`{"exp":{"a":"0x000000000000000000000000000000000bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","b": "0x0000000000000000000000000000000000000000000000000000000000000003"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("ac97ee6701be9d9ec2076ef8a4726274bbed7dfd6bc915507a939b92784d10c3", hex.EncodeToString(resp))

	data = []byte(`{"exp":{"a":"0x000000000000000000000000000000000bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","b": "0x000000000000000000000000000000000000000000000000000000000000000a"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("900d0865727d1f682df730ef2635b7d6e5b262d98eb00cd8c1c1e1ca8c659779", hex.EncodeToString(resp))

	data = []byte(`{"sdiv":{"a":"0x0ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("07fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffa", hex.EncodeToString(resp))

	data = []byte(`{"sdiv":{"a":"0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffa", hex.EncodeToString(resp))

	data = []byte(`{"byte":{"a":"0xac97ee6701be9d9ec2076ef8a4726274bbed7dfd6bc915507a939b92784d10c3","b":2}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("ee000000", hex.EncodeToString(resp))

	data = []byte(`{"byte":{"a":"0xac97ee6701be9d9ec2076ef8a4726274bbed7dfd6bc915507a939b92784d10c3","b":31}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("c3000000", hex.EncodeToString(resp))

	data = []byte(`{"mulmod":{"a":"0x0000000000000000000000000000000000000000000000000000000000000005","b": "0x0000000000000000000000000000000000000000000000000000000000000003","c":"0x0000000000000000000000000000000000000000000000000000000000000004"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", hex.EncodeToString(resp))

	data = []byte(`{"mulmod":{"a":"0x0000000000000ee000000000000000000bb000000000000000dd000000000005","b": "0x000000000000cc0000000000000000000000aa0000000000000000dd00000003","c":"0x0000000000000000000000000000000000000000000000000000000000000ef4"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000843", hex.EncodeToString(resp))

	data = []byte(`{"signextend":{"a":"0x00000000000000000000000000000000000000000000000000000000000ffff4","b": "0x0000000000000000000000000000000000000000000000000000000000000022"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("00000000000000000000000000000000000000000000000000000000000ffff4", hex.EncodeToString(resp))

	data = []byte(`{"signextend":{"a":"0x00000000000000000000000000000000000000000000000000000000000ffff4","b": "0x0000000000000000000000000000000000000000000000000000000000000020"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("00000000000000000000000000000000000000000000000000000000000ffff4", hex.EncodeToString(resp))

	data = []byte(`{"signextend":{"a":"0x00000000000000000000000000000000000000000000000000000000000ffff4","b": "0x0000000000000000000000000000000000000000000000000000000000000001"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4", hex.EncodeToString(resp))

	data = []byte(`{"signextend":{"a":"0x0000000000000000000000000000000000000000000000000000000000eefff4","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffeefff4", hex.EncodeToString(resp))

	data = []byte(`{"signextend":{"a":"0x000000000000000000000000000000000000000000000000000000000eeffff4","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffeffff4", hex.EncodeToString(resp))

	data = []byte(`{"signextend":{"a":"0x0000000000000000000000000000000000000000000000000000000088eefff4","b": "0x0000000000000000000000000000000000000000000000000000000000000003"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("ffffffffffffffffffffffffffffffffffffffffffffffffffffffff88eefff4", hex.EncodeToString(resp))

	data = []byte(`{"signextend":{"a":"0x0000000000000000000000000000000000000000000000000000000078eefff4","b": "0x0000000000000000000000000000000000000000000000000000000000000003"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000078eefff4", hex.EncodeToString(resp))

	data = []byte(`{"signextend":{"a":"0x0000000000000000000000000000088eeffffffffffffffffffffffffffffff4","b": "0x0000000000000000000000000000000000000000000000000000000000000010"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("ffffffffffffffffffffffffffffff8eeffffffffffffffffffffffffffffff4", hex.EncodeToString(resp))

	data = []byte(`{"signextend":{"a":"0x000000088eeffffffffffffffffffffffffffffffffffffffffffffffffffff4","b": "0x000000000000000000000000000000000000000000000000000000000000001a"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("ffffffffffeffffffffffffffffffffffffffffffffffffffffffffffffffff4", hex.EncodeToString(resp))
}
