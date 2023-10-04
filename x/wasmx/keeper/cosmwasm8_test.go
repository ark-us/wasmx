package keeper_test

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"mythos/v1/crypto/ethsecp256k1"

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

	// taken from cosmwasm/contracts/crypto_verify
	//go:embed testdata/cw8/crypto_verify-aarch64.wasm
	crypto_verify []byte

	//go:embed testdata/cw8/cw20_base-aarch64.wasm
	cw20_base []byte
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

type ChainQueryInner struct {
	Request cw8types.QueryRequest `json:"request"`
}
type ChainQuery struct {
	Chain ChainQueryInner `json:"chain"`
}

type RawQueryInner struct {
	Contract string `json:"contract"`
	Key      []byte `json:"key"`
}
type RawQuery struct {
	Raw RawQueryInner `json:"raw"`
}

type SpecialQueryCapitalized struct {
	Text string `json:"text"`
}
type SpecialQuery struct {
	Capitalized SpecialQueryCapitalized `json:"capitalized"`
}

type VerifyCosmosSignature struct {
	/// Message to verify.
	Message []byte `json:"message"`
	/// Serialized signature. Cosmos format (64 bytes).
	Signature []byte `json:"signature"`
	/// Serialized compressed (33 bytes) or uncompressed (65 bytes) public key.
	PublicKey []byte `json:"public_key"`
}

type VerifyEthereumText struct {
	/// Message to verify. This will be wrapped in the standard container
	/// `"\x19Ethereum Signed Message:\n" + len(message) + message` before verification.
	Message string `json:"message"`
	/// Serialized signature. Fixed length format (64 bytes `r` and `s` plus the one byte `v`).
	Signature []byte `json:"signature"`
	/// Signer address.
	/// This is matched case insensitive, so you can provide checksummed and non-checksummed addresses. Checksums are not validated.
	SignerAddress string `json:"signer_address"`
}

type VerifyEthereumTransaction struct {
	/// Ethereum address in hex format (42 characters, starting with 0x)
	From string `json:"from"`
	/// Ethereum address in hex format (42 characters, starting with 0x)
	To    string `json:"to"`
	Nonce uint64 `json:"nonce"`
	// GasLimit Uint128 `json:"gas_limit"`
	// GasPrice Uint128 `json:"gas_price"`
	// Value Uint128 `json:"value"`
	Data    []byte `json:"data"`
	ChainId uint64 `json:"chain_id"`
	R       []byte `json:"r"`
	S       []byte `json:"s"`
	V       uint64 `json:"v"`
}

type VerifyTendermintSignature struct {
	/// Message to verify.
	Message []byte `json:"message"`
	/// Serialized signature. Tendermint format (64 bytes).
	Signature []byte `json:"signature"`
	/// Serialized public key. Tendermint format (32 bytes).
	PublicKey []byte `json:"public_key"`
}

type VerifyTendermintBatch struct {
	/// Messages to verify.
	Messages [][]byte `json:"messages"`
	/// Serialized signatures. Tendermint format (64 bytes).
	Signatures [][]byte `json:"signatures"`
	/// Serialized public keys. Tendermint format (32 bytes).
	PublicKeys [][]byte `json:"public_keys"`
}

type ListVerificationSchemes struct{}

type VerifyCosmosSignatureWrap struct {
	VerifyCosmosSignature VerifyCosmosSignature `json:"verify_cosmos_signature"`
}

type VerifyEthereumTextWrap struct {
	VerifyEthereumText VerifyEthereumText `json:"verify_ethereum_text"`
}

type VerifyTendermintSignatureWrap struct {
	VerifyTendermintSignature VerifyTendermintSignature `json:"verify_tendermint_signature"`
}

type VerifyTendermintBatchWrap struct {
	VerifyTendermintBatch VerifyTendermintBatch `json:"verify_tendermint_batch"`
}

type ListVerificationSchemesWrap struct {
	ListVerificationSchemes ListVerificationSchemes `json:"list_verification_schemes"`
}

const SECP256K1_MESSAGE_HEX = "5c868fedb8026979ebd26f1ba07c27eedf4ff6d10443505a96ecaf21ba8c4f0937b3cd23ffdc3dd429d4cd1905fb8dbcceeff1350020e18b58d2ba70887baa3a9b783ad30d3fbf210331cdd7df8d77defa398cdacdfc2e359c7ba4cae46bb74401deb417f8b912a1aa966aeeba9c39c7dd22479ae2b30719dca2f2206c5eb4b7"
const SECP256K1_SIGNATURE_HEX = "207082eb2c3dfa0b454e0906051270ba4074ac93760ba9e7110cd9471475111151eb0dbbc9920e72146fb564f99d039802bf6ef2561446eb126ef364d21ee9c4"
const SECP256K1_PUBLIC_KEY_HEX = "04051c1ee2190ecfb174bfe4f90763f2b4ff7517b70a2aec1876ebcfd644c4633fb03f3cfbd94b1f376e34592d9d41ccaf640bb751b00a1fadeb0c01157769eb73"

// TEST 3 test vector from https://tools.ietf.org/html/rfc8032#section-7.1
const ED25519_MESSAGE_HEX = "af82"
const ED25519_SIGNATURE_HEX = "6291d657deec24024827e69c3abe01a30ce548a284743a445e3680d7db5ac3ac18ff9b538d16f290ae67f760984dc6594a7c15e9716ed28dc027beceea1ec40a"
const ED25519_PUBLIC_KEY_HEX = "fc51cd8e6218a1a38da47ed00230f0580816ed13ba3303ac5deb911548908025"

// Signed text "connect all the things" using MyEtherWallet with private key b5b1870957d373ef0eeffecc6e4812c0fd08f554b37b233526acc331bf1544f7
const ETHEREUM_MESSAGE = "connect all the things"
const ETHEREUM_SIGNATURE_HEX = "dada130255a447ecf434a2df9193e6fbba663e4546c35c075cd6eea21d8c7cb1714b9b65a4f7f604ff6aad55fba73f8c36514a512bbbba03709b37069194f8a41b"
const ETHEREUM_SIGNER_ADDRESS = "0x12890D2cce102216644c59daE5baed380d84830c"

// TEST 2 test vector from https://tools.ietf.org/html/rfc8032#section-7.1
const ED25519_MESSAGE2_HEX = "72"
const ED25519_SIGNATURE2_HEX = "92a009a9f0d4cab8720e820b5f642540a2b27b5416503f8fb3762223ebdb69da085ac1e43e15996e458f3613d0f11d8c387b2eaeb4302aeeb00d291612bb0c00"
const ED25519_PUBLIC_KEY2_HEX = "3d4017c3e843895a92b70aa74d1b7ebc9c982ccf2ec4968cc0cd55f12af4660c"

func (suite *KeeperTestSuite) TestWasmxCWSimpleContract() {
	wasmbin := cwSimpleContract
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)

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

	data = []byte(`{"increase":{}}`)
	abcires, err := appA.WasmxQueryRawNoCheck(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().NoError(err)
	s.Require().True(abcires.IsErr())
	s.Require().Contains(abcires.Log, cw8types.ERROR_FLAG_QUERY)

	data = []byte(`{"increase":{}}`)
	_, _, err = appA.ExecuteContractSimulate(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().NoError(err)
}

type Cw20Coin struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

type MinterResponse struct {
}

type InstantiateMarketingInfo struct {
}

type CW20InstantiateMsg struct {
	Name            string                    `json:"name"`
	Symbol          string                    `json:"symbol"`
	Decimals        uint8                     `json:"decimals"`
	InitialBalances []Cw20Coin                `json:"initial_balances"`
	Mint            *MinterResponse           `json:"mint"`
	Marketing       *InstantiateMarketingInfo `json:"marketing"`
}

func (suite *KeeperTestSuite) TestWasmxCW20() {
	wasmbin := cw20_base
	sender := suite.GetRandomAccount()
	recipient := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	expectedDeps := []string{types.CW_ENV_8}

	codeId := appA.StoreCode(sender, wasmbin, nil)
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
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: calld}, "cwSimpleContract", nil)

	calld = []byte(fmt.Sprintf(`{"balance":{"address":"%s"}}`, sender.Address.String()))
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: calld}, nil, nil)
	suite.Require().Equal(`{"balance":"10000000000000000"}`, string(qres))

	data := []byte(fmt.Sprintf(`{"transfer":{"recipient":"%s","amount":"%s"}}`, recipient.Address.String(), "2000000000000000"))
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	calld = []byte(fmt.Sprintf(`{"balance":{"address":"%s"}}`, recipient.Address.String()))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: calld}, nil, nil)
	suite.Require().Equal(`{"balance":"2000000000000000"}`, string(qres))

	calld = []byte(fmt.Sprintf(`{"balance":{"address":"%s"}}`, sender.Address.String()))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: calld}, nil, nil)
	suite.Require().Equal(`{"balance":"8000000000000000"}`, string(qres))
}

func (suite *KeeperTestSuite) TestWasmxCW20ByEthereumTx() {
	wasmbin := cw20_base
	deployer := suite.GetRandomAccount()
	recipient := suite.GetRandomAccount()
	priv, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err)
	senderAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())
	sender := simulation.Account{
		Address: senderAddress,
	}
	initBalance := sdkmath.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), deployer.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	expectedDeps := []string{types.CW_ENV_8}

	codeId := appA.StoreCode(deployer, wasmbin, nil)
	codeInfo := appA.App.WasmxKeeper.GetCodeInfo(appA.Context(), codeId)
	s.Require().ElementsMatch(expectedDeps, codeInfo.Deps, "wrong deps")

	instantiateMsg := CW20InstantiateMsg{
		Name:     "cw20",
		Symbol:   "TKN",
		Decimals: 18,
		InitialBalances: []Cw20Coin{
			{Address: deployer.Address.String(), Amount: "10000000000000000"},
			{Address: sender.Address.String(), Amount: "10000000000000000"},
		},
	}
	calld, err := json.Marshal(instantiateMsg)
	s.Require().NoError(err)
	contractAddress := appA.InstantiateCode(deployer, codeId, types.WasmxExecutionMessage{Data: calld}, "cwSimpleContract", nil)
	contractEvmAddress := types.EvmAddressFromAcc(contractAddress)

	databz := []byte(fmt.Sprintf(`{"transfer":{"recipient":"%s","amount":"%s"}}`, recipient.Address.String(), "2000000000000000"))

	appA.SendEthTx(priv, &contractEvmAddress, databz, nil, uint64(1000000), big.NewInt(10000), nil)

	databz = []byte(fmt.Sprintf(`{"balance":{"address":"%s"}}`, sender.Address.String()))
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: databz}, nil, nil)
	suite.Require().Equal(`{"balance":"8000000000000000"}`, string(qres))
}

func (suite *KeeperTestSuite) TestWasmxCwAtomicSwap() {
	wasmbin := cw20_atomic_swap
	sender := suite.GetRandomAccount()
	recipient := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000_000)

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
	coins := sdk.NewCoins(sdk.NewCoin(appA.Denom, sdkmath.NewInt(10000000)))

	data := fmt.Sprintf(`{"create":{"id":"swap1","hash":"%s","recipient":"%s","expires":{"at_height":10000}}}`, hashHex, recipient.Address.String())
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(data)}, coins, nil)

	balanceContract, err := appA.App.BankKeeper.Balance(appA.Context(), &banktypes.QueryBalanceRequest{Address: contractAddress.String(), Denom: appA.Denom})
	s.Require().NoError(err)

	balanceReceiver, err := appA.App.BankKeeper.Balance(appA.Context(), &banktypes.QueryBalanceRequest{Address: recipient.Address.String(), Denom: appA.Denom})
	s.Require().NoError(err)

	s.Require().Equal(coins[0].Amount, balanceContract.Balance.Amount)
	s.Require().Equal(sdkmath.NewInt(0), balanceReceiver.Balance.Amount)

	data = fmt.Sprintf(`{"release":{"id":"swap1","preimage":"%s"}}`, preimage)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(data)}, nil, nil)

	balanceContract, err = appA.App.BankKeeper.Balance(appA.Context(), &banktypes.QueryBalanceRequest{Address: contractAddress.String(), Denom: appA.Denom})
	s.Require().NoError(err)

	balanceReceiver, err = appA.App.BankKeeper.Balance(appA.Context(), &banktypes.QueryBalanceRequest{Address: recipient.Address.String(), Denom: appA.Denom})
	s.Require().NoError(err)

	s.Require().Equal(coins[0].Amount, balanceReceiver.Balance.Amount)
	s.Require().Equal(sdkmath.NewInt(0), balanceContract.Balance.Amount)

	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(`{"list":{}}`)}, nil, nil)
	suite.Require().Equal(`{"swaps":[]}`, string(qres))

	data = fmt.Sprintf(`{"create":{"id":"swap2","hash":"%s","recipient":"%s","expires":{"at_height":10000}}}`, hashHex, recipient.Address.String())
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(data)}, coins, nil)

	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(`{"list":{}}`)}, nil, nil)
	suite.Require().Equal(`{"swaps":["swap2"]}`, string(qres))

	data = fmt.Sprintf(`{"create":{"id":"swap3","hash":"%s","recipient":"%s","expires":{"at_height":10000}}}`, hashHex, recipient.Address.String())
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(data)}, coins, nil)

	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(`{"list":{"start_after":"swap1","limit":2}}`)}, nil, nil)
	suite.Require().Equal(`{"swaps":["swap2","swap3"]}`, string(qres))
}

func (suite *KeeperTestSuite) TestWasmxCwReflect() {
	wasmbin := wasm_reflect
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000_000)

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
	suite.Require().Equal(string(expectedReplyBz), string(qres))

	// test chain query
	query := ChainQuery{
		Chain: ChainQueryInner{
			Request: cw8types.QueryRequest{
				Bank: &cw8types.BankQuery{
					Balance: &cw8types.BalanceQuery{
						Address: contractAddress.String(),
						Denom:   appA.Denom,
					},
				},
			},
		},
	}
	queryBz, err := json.Marshal(query)
	s.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: queryBz}, nil, nil)
	var qresData cw8types.RawResponse
	err = json.Unmarshal(qres, &qresData)
	s.Require().NoError(err)
	suite.Require().Equal(`{"amount":{"denom":"amyt","amount":"0"}}`, string(qresData.Data))

	// test chain query WasmQuery
	queryWasmx := types.WasmxExecutionMessage{
		Data: []byte(`{"value":{}}`),
	}
	queryWasmxBz, err := json.Marshal(queryWasmx)
	s.Require().NoError(err)
	query = ChainQuery{
		Chain: ChainQueryInner{
			Request: cw8types.QueryRequest{
				Wasm: &cw8types.WasmQuery{
					Smart: &cw8types.SmartQuery{
						ContractAddr: contractAddressCounter.String(),
						Msg:          queryWasmxBz,
					},
				},
			},
		},
	}
	queryBz, err = json.Marshal(query)
	s.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: queryBz}, nil, nil)
	var qresData2 cw8types.RawResponse
	err = json.Unmarshal(qres, &qresData2)
	s.Require().NoError(err)
	suite.Require().Equal(`{"value":4}`, string(qresData2.Data))

	// test querying another contract
	query2 := RawQuery{
		Raw: RawQueryInner{
			Contract: contractAddressCounter.String(),
			Key:      []byte("counter"),
		},
	}
	queryBz, err = json.Marshal(query2)
	s.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: queryBz}, nil, nil)
	var qresData3 cw8types.RawResponse
	err = json.Unmarshal(qres, &qresData3)
	s.Require().NoError(err)
	suite.Require().Equal(`4`, string(qresData3.Data))
}

func (suite *KeeperTestSuite) TestWasmxCwCrypto() {
	wasmbin := crypto_verify
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	expectedDeps := []string{types.CW_ENV_8}

	codeId := appA.StoreCode(sender, wasmbin, nil)
	codeInfo := appA.App.WasmxKeeper.GetCodeInfo(appA.Context(), codeId)
	s.Require().ElementsMatch(expectedDeps, codeInfo.Deps, "wrong deps")

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{}`)}, "crypto_verify", nil)

	// cw_8_secp256k1_verify
	msg, err := hex.DecodeString(SECP256K1_MESSAGE_HEX)
	s.Require().NoError(err)
	signature, err := hex.DecodeString(SECP256K1_SIGNATURE_HEX)
	s.Require().NoError(err)
	pubKey, err := hex.DecodeString(SECP256K1_PUBLIC_KEY_HEX)
	s.Require().NoError(err)
	req := VerifyCosmosSignatureWrap{
		VerifyCosmosSignature: VerifyCosmosSignature{
			Message:   msg,
			Signature: signature,
			PublicKey: pubKey,
		},
	}
	reqBz, err := json.Marshal(req)
	s.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: reqBz}, nil, nil)
	suite.Require().Equal(`{"verifies":true}`, string(qres))

	// cw_8_secp256k1_verify - test 2
	msg = []byte("hello")
	signature, err = sender.PrivKey.Sign(msg)
	s.Require().NoError(err)
	req = VerifyCosmosSignatureWrap{
		VerifyCosmosSignature: VerifyCosmosSignature{
			Message:   msg,
			Signature: signature,
			PublicKey: sender.PubKey.Bytes(),
		},
	}
	reqBz, err = json.Marshal(req)
	s.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: reqBz}, nil, nil)
	suite.Require().Equal(`{"verifies":true}`, string(qres))

	// TODO
	// // ethereum signature test + cw_8_secp256k1_recover_pubkey test
	// signature, err = hex.DecodeString(ETHEREUM_SIGNATURE_HEX)
	// s.Require().NoError(err)
	// reqEth := VerifyEthereumTextWrap{
	// 	VerifyEthereumText: VerifyEthereumText{
	// 		Message:       ETHEREUM_MESSAGE,
	// 		Signature:     signature,
	// 		SignerAddress: ETHEREUM_SIGNER_ADDRESS,
	// 	},
	// }
	// reqBz, err = json.Marshal(reqEth)
	// s.Require().NoError(err)
	// qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: reqBz}, nil, nil)
	// suite.Require().Equal(`{"verifies":true}`, string(qres))

	// cw_8_ed25519_verify
	msg, err = hex.DecodeString(ED25519_MESSAGE_HEX)
	s.Require().NoError(err)
	signature, err = hex.DecodeString(ED25519_SIGNATURE_HEX)
	s.Require().NoError(err)
	pubKey, err = hex.DecodeString(ED25519_PUBLIC_KEY_HEX)
	s.Require().NoError(err)
	reqEd := VerifyTendermintSignatureWrap{
		VerifyTendermintSignature: VerifyTendermintSignature{
			Message:   msg,
			Signature: signature,
			PublicKey: pubKey,
		},
	}
	reqBz, err = json.Marshal(reqEd)
	s.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: reqBz}, nil, nil)
	suite.Require().Equal(`{"verifies":true}`, string(qres))

	// cw_8_ed25519_verify2
	msg2, err := hex.DecodeString(ED25519_MESSAGE2_HEX)
	s.Require().NoError(err)
	signature2, err := hex.DecodeString(ED25519_SIGNATURE2_HEX)
	s.Require().NoError(err)
	pubKey2, err := hex.DecodeString(ED25519_PUBLIC_KEY2_HEX)
	s.Require().NoError(err)
	reqEd = VerifyTendermintSignatureWrap{
		VerifyTendermintSignature: VerifyTendermintSignature{
			Message:   msg2,
			Signature: signature2,
			PublicKey: pubKey2,
		},
	}
	reqBz, err = json.Marshal(reqEd)
	s.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: reqBz}, nil, nil)
	suite.Require().Equal(`{"verifies":true}`, string(qres))

	// TODO
	// // cw_8_ed25519_batch_verify
	// msg, err = hex.DecodeString(SECP256K1_MESSAGE_HEX)
	// s.Require().NoError(err)
	// signature, err = hex.DecodeString(SECP256K1_SIGNATURE_HEX)
	// s.Require().NoError(err)
	// pubKey, err = hex.DecodeString(SECP256K1_PUBLIC_KEY_HEX)
	// s.Require().NoError(err)
	// reqEdBatch := VerifyTendermintBatchWrap{
	// 	VerifyTendermintBatch: VerifyTendermintBatch{
	// 		Messages:   [][]byte{msg, msg2},
	// 		Signatures: [][]byte{signature, signature2},
	// 		PublicKeys: [][]byte{pubKey, pubKey2},
	// 	},
	// }
	// reqBz, err = json.Marshal(reqEdBatch)
	// s.Require().NoError(err)
	// qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: reqBz}, nil, nil)
	// suite.Require().Equal(`{"verifies":true}`, string(qres))

	reqList := ListVerificationSchemesWrap{
		ListVerificationSchemes: ListVerificationSchemes{},
	}
	reqBz, err = json.Marshal(reqList)
	s.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: reqBz}, nil, nil)
	suite.Require().Equal(`{"verification_schemes":["secp256k1","ed25519","ed25519_batch"]}`, string(qres))
}
