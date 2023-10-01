package keeper_test

import (
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"

	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm"
)

func (suite *KeeperTestSuite) TestEwasmPrecompileIdentityDirect() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	contractAddress := types.AccAddressFromHex("0x0000000000000000000000000000000000000004")

	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("aa0000000000000000000000000000000000000000000000000000000077")}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "aa0000000000000000000000000000000000000000000000000000000077")

	queryMsg := "aa0000000000000000000000000000000000000000000000000000000077"
	qres := appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(queryMsg)}, nil, nil)
	s.Require().Equal("aa0000000000000000000000000000000000000000000000000000000077", qres)
}

func (suite *KeeperTestSuite) TestEwasmPrecompileEcrecoverEthDirect() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	contractAddress := types.AccAddressFromHex("0x000000000000000000000000000000000000001f")

	inputhex := "38d18acb67d25c8bb9942764b62f18e17054f66a817bd4295423adf9ed98873e000000000000000000000000000000000000000000000000000000000000001b38d18acb67d25c8bb9942764b62f18e17054f66a817bd4295423adf9ed98873e789d1dd423d25f0772d2748d60f7e4b81bb14d086eba8e8e8efb6dcff8a4ae02"

	qres := appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(inputhex)}, nil, nil)
	s.Require().Equal("000000000000000000000000ceaccac640adf55b2028469bd36ba501f28b699d", qres)
}

// type SignedMesssage struct {
// 	ChainId string `json:"chain_id"`
// 	AccountNumber string `json:"account_number"`
// 	Sequence `json:"sequence"`
// 	Fee
// }

func (suite *KeeperTestSuite) TestVerification() {
	mnemonic := "carpet toddler provide announce damage uniform excite planet gap apology spatial nerve increase clump dial enforce basket thunder grain flee tent sleep heavy bonus"

	// algo secp256k1
	// accountPubKey := []byte{2, 88, 67, 184, 75, 57, 134, 207, 115, 243, 57, 191, 96, 117, 251, 24, 202, 193, 170, 41, 152, 7, 47, 100, 241, 36, 47, 133, 234, 24, 103, 115, 97}
	accountBech32 := "mythos1sap78sy28k29xq48vfk37w5hwvd54ltq54jt69"
	// accountBech32 := "cosmos1sap78sy28k29xq48vfk37w5hwvd54ltqt6pg84"

	s.Require().True(bip39.IsMnemonicValid(mnemonic))

	hdPath := hd.CreateHDPath(118, 0, 0).String()

	derivedPriv, err := hd.Secp256k1.Derive()(mnemonic, "", hdPath)
	s.Require().NoError(err)
	privKey := hd.Secp256k1.Generate()(derivedPriv)

	// privKey := secp256k1.GenPrivKeyFromSecret(pk.GetKey().Seed())
	pubKey := privKey.PubKey()
	address := sdk.AccAddress(pubKey.Address())

	s.Require().Equal(accountBech32, address.String())

	// msg := `{"scope":"oauth","client_id":"1234","nonce":"TGKUgDB77OL4ScJNLaRhTNj3CBJDr1eltGiiivjyYco=","chain_id":"mythos_7000-7"}`
	// signature := "2ExEqYF7XWE/q9D7vgOOwbDm5ZUXeyU0szXQicvIwJVeNj3tq4CjO8a6I1kiQVTW2MQfwhv/2NI6um2JEEvpRQ=="
	// signaturePubKey := "AlhDuEs5hs9z8zm/YHX7GMrBqimYBy9k8SQvheoYZ3Nh"
	// signaturePubKeyType := "tendermint/PubKeySecp256k1"

	// msg := `{"chain_id":"","account_number":"0","sequence":"0","fee":{"gas":"0","amount":[]},"msgs":[{"type":"sign/MsgSignData","value":{"signer":"mythos1sap78sy28k29xq48vfk37w5hwvd54ltq54jt69","data":"a"}}],"memo":""}`
	msg := `{"account_number":"0","chain_id":"","fee":{"amount":[],"gas":"0"},"memo":"","msgs":[{"type":"sign/MsgSignData","value":{"data":"YQ==","signer":"mythos1sap78sy28k29xq48vfk37w5hwvd54ltq54jt69"}}],"sequence":"0"}`
	signature := "LqSI0Lx1Xy7fLUn07Xx0aVV2jpuCA7mHkj/CvwTVbf8CWrVgElmvUuHwikrA50PPwlr8mdspHSt+eVc+br4Qcw=="

	signedMsg := []byte(msg)
	// Sign applies sha256 on the message
	signature2, err := privKey.Sign(signedMsg)
	s.Require().NoError(err)

	signature2str := base64.StdEncoding.EncodeToString(signature2)
	s.Require().Equal(signature, signature2str)

	signatureBz, err := base64.StdEncoding.DecodeString(signature)
	s.Require().NoError(err)
	// VerifySignature applies sha256 on the message
	verified := pubKey.VerifySignature(signedMsg, signatureBz)
	s.Require().True(verified)

	msg = `{"account_number":"0","chain_id":"","fee":{"amount":[],"gas":"0"},"memo":"","msgs":[{"type":"sign/MsgSignData","value":{"data":"eyJzY29wZSI6Im9hdXRoIiwiY2xpZW50X2lkIjoiMTIzNCIsIm5vbmNlIjoiRlo2U25MUmRZL2M0V3dxZ01GSitQSzJjdDd2Q0g5YkVkdkhrNFhWdEZaTT0ifQ==","signer":"mythos1sap78sy28k29xq48vfk37w5hwvd54ltq54jt69"}}],"sequence":"0"}`
	signature = "2KPHrWqQdfOCkaWQQ40XdG4LsJo3v9iMgw00ZIH5fpdO2n4FDg1hqApOS6u1Q7XEk/ylPSm4kYMT2aQkGfeJKw=="
	signature2, err = privKey.Sign([]byte(msg))
	s.Require().NoError(err)
	signature2str = base64.StdEncoding.EncodeToString(signature2)
	s.Require().Equal(signature, signature2str)
	signatureBz, err = base64.StdEncoding.DecodeString(signature)
	s.Require().NoError(err)
	// VerifySignature applies sha256 on the message
	verified = pubKey.VerifySignature([]byte(msg), signatureBz)
	s.Require().True(verified)
}

func (suite *KeeperTestSuite) TestEwasmPrecompileEcrecoverDirect() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	contractAddress := types.AccAddressFromHex("0x0000000000000000000000000000000000000001")

	message := []byte("This is a test message")
	msgHash_ := sha256.Sum256(message)
	msgHash := msgHash_[:]

	// Signature must be compatible with Ethereum
	privKeyBtcec := (sender.PrivKey).(*secp256k1.PrivKey)
	btcecPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBtcec.Key)
	signature, err := btcec.SignCompact(btcec.S256(), btcecPrivKey, msgHash, false)
	s.Require().NoError(err)
	v := signature[0] - 27
	copy(signature, signature[1:])
	signature[64] = v

	verified := sender.PubKey.VerifySignature(message, signature[:64])
	s.Require().True(verified)

	inputbz := append(msgHash, signature...)
	senderhex := strings.ToLower(types.EvmAddressFromAcc(sender.Address).Hex())

	qres := appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: inputbz}, nil, nil)
	s.Require().Equal(senderhex[2:], qres)
}

func (suite *KeeperTestSuite) TestEwasmPrecompileModexpDirect() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	contractAddress := types.AccAddressFromHex("0x0000000000000000000000000000000000000005")

	// <length_of_BASE> <length_of_EXPONENT> <length_of_MODULUS> <BASE> <EXPONENT> <MODULUS>
	// https://eips.ethereum.org/EIPS/eip-198
	// https://github.com/ethereumproject/evm-rs/blob/master/precompiled/modexp/src/lib.rs#L133
	// https://github.com/ethereum/tests/blob/develop/GeneralStateTests/stPreCompiledContracts/modexpTests.json

	calldata := "00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000002003fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2efffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f"
	qres := appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calldata)}, nil, nil)
	expected := "0000000000000000000000000000000000000000000000000000000000000001"
	s.Require().Equal(expected, qres)

	calldata = "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000020fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2efffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calldata)}, nil, nil)
	expected = "0000000000000000000000000000000000000000000000000000000000000000"
	s.Require().Equal(expected, qres)

	calldata = "00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002003ffff800000000000000000000000000000000000000000000000000000000000000007"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calldata)}, nil, nil)
	expected = "3b01b01ac41f2d6e917c6d6a221ce793802469026d9ab7578fa2e79e4da6aaab"
	s.Require().Equal(expected, qres)

	calldata = "00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002003ffff80"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calldata)}, nil, nil)
	expected = "3b01b01ac41f2d6e917c6d6a221ce793802469026d9ab7578fa2e79e4da6aaab"
	s.Require().Equal(expected, qres)

	calldata = "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd"
	res := appA.ExecuteContractNoCheck(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calldata)}, nil, nil, 1500000, nil)
	s.Require().True(res.IsErr(), res.GetLog())

	calldata = "00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000004003fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2efffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calldata)}, nil, nil)
	expected = "fd24265072b6b01f9bc93300ae72996f1eb0ef2cc4a943b140c6bf2215143e51765316da9900a45dc6b1c0f71df37fbf1a15f274353de964b74822bf76b98b19"
	s.Require().Equal(expected, qres)

	// modexp less than 32 bytes
	calldata = "00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001c03fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2efffffffffffffffffffffffffffffffffffffffffffffffffffffffe"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calldata)}, nil, nil)
	expected = "bfcd5188ac621ad4d2690eaee537a3b7114509341402010dcb8f1c31"
	s.Require().Equal(expected, qres)

	// modexp with modulus 48 bytes
	calldata = "0000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000300a366e771bd3bb98a41a1e5c0748561b41e8dc9c22ebc6a9d39dd8797631915ddf7e7dcc5700e7727d336f7c61b3e28a2654cd1c523fb5f67bb11684dad2da807ca0b2a17c660a0f2639fa5026ab9d2ac8d8116848e534bdf3c0955850d7175601fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffeffffffff0000000000000000ffffffff"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calldata)}, nil, nil)
	expected = "5de8b2b22ecdf6790f0c7de8ea01bdd6fb8446353273f6053dd29c5ef32974403861d4b388cefccf2e01f63f53b6ffe0"
	s.Require().Equal(expected, qres)

	// extra
	calldata = "00000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000003003e501df64c8d7065d58eac499351e2afcdc74fda6bd4980919ca5dcf51075e51e36e9442aba748d8d9931e0f1332bd6fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffeffffffff0000000000000000fffffffdfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffeffffffff0000000000000000ffffffff"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calldata)}, nil, nil)
	expected = "ba2909a8e60a55d7a0caf129a18c6c6aa41434c431646bb4a928e76ad732152f35eb59e6df429de7323e5813809f03dc"
	s.Require().Equal(expected, qres)
}

func (suite *KeeperTestSuite) TestEwasmPrecompileSecretSharingDirect() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	contractAddress := types.AccAddressFromHex("0x0000000000000000000000000000000000000022")

	args1 := vm.InputShamirSplit{
		Secret:    "this is a secret",
		Count:     4,
		Threshold: 2,
	}

	fabi := vm.SecretSharingAbi.Methods["shamirSplit"]
	input, err := fabi.Inputs.Pack(args1.Secret, args1.Count, args1.Threshold)
	s.Require().NoError(err)
	input = append(fabi.ID, input...)

	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: input}, nil, nil)

	unpacked, err := fabi.Outputs.Unpack(qres)
	s.Require().NoError(err)

	var tuple vm.ResultShares
	err = fabi.Outputs.Copy(&tuple, unpacked)
	s.Require().NoError(err)

	sampleShares := []string{
		tuple.Shares[0],
		tuple.Shares[2],
	}

	fabi = vm.SecretSharingAbi.Methods["shamirRecover"]
	input, err = fabi.Inputs.Pack(sampleShares)
	s.Require().NoError(err)
	input = append(fabi.ID, input...)

	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: input}, nil, nil)

	unpacked, err = fabi.Outputs.Unpack(qres)
	s.Require().NoError(err)

	var result vm.ResultSecret
	err = fabi.Outputs.Copy(&result, unpacked)
	s.Require().NoError(err)
	s.Require().Equal(args1.Secret, result.Secret)
}
