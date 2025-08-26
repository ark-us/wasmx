package keeper_test

import (
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simulation "github.com/cosmos/cosmos-sdk/types/simulation"

	cmttypes "github.com/cometbft/cometbft/types"

	mcfg "github.com/loredanacirstea/wasmx/config"

	testdata "github.com/loredanacirstea/mythos-tests/network/testdata/wasmx"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	networkserver "github.com/loredanacirstea/wasmx/x/network/server"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestBlockHeader() {
	wasmbinTo := testdata.WasmxSimpleStorage
	chainId := mcfg.MYTHOS_CHAIN_ID_TEST
	suite.SetCurrentChain(chainId)
	chain := suite.GetChain(chainId)

	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	sender := simulation.Account{
		PrivKey: chain.SenderPrivKey,
		PubKey:  chain.SenderAccount.GetPubKey(),
		Address: chain.SenderAccount.GetAddress(),
	}

	appA := s.AppContext()
	denom := appA.Chain.Config.BaseDenom
	senderPrefixedLevel0 := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixedLevel0, sdk.NewCoin(denom, initBalance))
	suite.Commit()

	codeIdTo := appA.StoreCode(sender, wasmbinTo, nil)
	appA.InstantiateCode(sender, codeIdTo, wasmxtypes.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"crosschain_contract":"%s"}`, wasmxtypes.ROLE_MULTICHAIN_REGISTRY))}, "wasmbinTo", nil)

	abcicli := appA.ABCIClient()
	lastHeight, err := abcicli.LatestBlockHeight(appA.Context())
	suite.Require().NoError(err)
	lastHeight = lastHeight - 1
	lastBlock, err := abcicli.Block(appA.Context(), &lastHeight)
	suite.Require().NoError(err)

	lastBlockResult, err := abcicli.Commit(appA.Context(), &lastHeight)
	suite.Require().NoError(err)

	suite.Require().Equal(lastBlock.BlockID.Hash, lastBlock.Block.Header.Hash())

	err = lastBlockResult.Header.ValidateBasic()
	suite.Require().NoError(err)

	validators, err := abcicli.Validators(appA.Context(), nil, nil, nil)
	suite.Require().NoError(err)
	valSet, err := cmttypes.ValidatorSetFromExistingValidators(validators.Validators)
	suite.Require().NoError(err)

	lb := &cmttypes.LightBlock{
		SignedHeader: &lastBlockResult.SignedHeader,
		ValidatorSet: valSet,
	}
	err = lb.ValidateBasic(chainId)
	suite.Require().NoError(err)

	suite.Require().Equal(hex.EncodeToString(lastBlockResult.SignedHeader.ValidatorsHash), hex.EncodeToString(valSet.Hash()))

	err = valSet.VerifyCommitLight(chainId, lastBlockResult.Commit.BlockID, lastBlockResult.Height, lastBlockResult.Commit)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestStateSyncBootstrap() {
	SkipFixmeTests(suite.T(), "TestStateSyncBootstrap")
	statestr := `{"Version":{"consensus":{"block":11},"software":"0.38.6"},"ChainID":"mythos_7000-14","InitialHeight":1,"LastBlockHeight":10,"LastBlockID":{"hash":"F3EDBE8D0B34156827D2693A73E809C7A9C783DB06409060B173139F6EBDAA59","parts":{"total":1,"hash":"F3EDBE8D0B34156827D2693A73E809C7A9C783DB06409060B173139F6EBDAA59"}},"LastBlockTime":"2024-07-29T14:37:43.302Z","NextValidators":"eyJ2YWxpZGF0b3JzIjpbeyJhZGRyZXNzIjoiNmFkMDlmYTZjOThiZDNmZTYyMzM2MTg2NjU2Yjg2MWZmMzkzNzE0NiIsInB1Yl9rZXkiOnsidHlwZV91cmwiOiIvY29zbW9zLmNyeXB0by5lZDI1NTE5LlB1YktleSIsInZhbHVlIjoiZXlKclpYa2lPaUo2Y0cxUk5VTm9WeXRFYWtKSWRrMXlhVGRKY2pKdllrbE5aMjlZVTBsVldsRnliRFo0VFdaMWFWSnZQU0o5In0sInZvdGluZ19wb3dlciI6IjEwMDAwMDAwMDAwMDAwMCIsInByb3Bvc2VyX3ByaW9yaXR5IjoiMCJ9XSwicHJvcG9zZXIiOnsiYWRkcmVzcyI6IjZhZDA5ZmE2Yzk4YmQzZmU2MjMzNjE4NjY1NmI4NjFmZjM5MzcxNDYiLCJwdWJfa2V5Ijp7InR5cGVfdXJsIjoiL2Nvc21vcy5jcnlwdG8uZWQyNTUxOS5QdWJLZXkiLCJ2YWx1ZSI6ImV5SnJaWGtpT2lKNmNHMVJOVU5vVnl0RWFrSklkazF5YVRkSmNqSnZZa2xOWjI5WVUwbFZXbEZ5YkRaNFRXWjFhVkp2UFNKOSJ9LCJ2b3RpbmdfcG93ZXIiOiIxMDAwMDAwMDAwMDAwMDAiLCJwcm9wb3Nlcl9wcmlvcml0eSI6IjAifX0=","Validators":"eyJ2YWxpZGF0b3JzIjpbeyJhZGRyZXNzIjoiNmFkMDlmYTZjOThiZDNmZTYyMzM2MTg2NjU2Yjg2MWZmMzkzNzE0NiIsInB1Yl9rZXkiOnsidHlwZV91cmwiOiIvY29zbW9zLmNyeXB0by5lZDI1NTE5LlB1YktleSIsInZhbHVlIjoiZXlKclpYa2lPaUo2Y0cxUk5VTm9WeXRFYWtKSWRrMXlhVGRKY2pKdllrbE5aMjlZVTBsVldsRnliRFo0VFdaMWFWSnZQU0o5In0sInZvdGluZ19wb3dlciI6IjEwMDAwMDAwMDAwMDAwMCIsInByb3Bvc2VyX3ByaW9yaXR5IjoiMCJ9XSwicHJvcG9zZXIiOnsiYWRkcmVzcyI6IjZhZDA5ZmE2Yzk4YmQzZmU2MjMzNjE4NjY1NmI4NjFmZjM5MzcxNDYiLCJwdWJfa2V5Ijp7InR5cGVfdXJsIjoiL2Nvc21vcy5jcnlwdG8uZWQyNTUxOS5QdWJLZXkiLCJ2YWx1ZSI6ImV5SnJaWGtpT2lKNmNHMVJOVU5vVnl0RWFrSklkazF5YVRkSmNqSnZZa2xOWjI5WVUwbFZXbEZ5YkRaNFRXWjFhVkp2UFNKOSJ9LCJ2b3RpbmdfcG93ZXIiOiIxMDAwMDAwMDAwMDAwMDAiLCJwcm9wb3Nlcl9wcmlvcml0eSI6IjAifX0=","LastValidators":"eyJ2YWxpZGF0b3JzIjpbeyJhZGRyZXNzIjoiNmFkMDlmYTZjOThiZDNmZTYyMzM2MTg2NjU2Yjg2MWZmMzkzNzE0NiIsInB1Yl9rZXkiOnsidHlwZV91cmwiOiIvY29zbW9zLmNyeXB0by5lZDI1NTE5LlB1YktleSIsInZhbHVlIjoiZXlKclpYa2lPaUo2Y0cxUk5VTm9WeXRFYWtKSWRrMXlhVGRKY2pKdllrbE5aMjlZVTBsVldsRnliRFo0VFdaMWFWSnZQU0o5In0sInZvdGluZ19wb3dlciI6IjEwMDAwMDAwMDAwMDAwMCIsInByb3Bvc2VyX3ByaW9yaXR5IjoiMCJ9XSwicHJvcG9zZXIiOnsiYWRkcmVzcyI6IjZhZDA5ZmE2Yzk4YmQzZmU2MjMzNjE4NjY1NmI4NjFmZjM5MzcxNDYiLCJwdWJfa2V5Ijp7InR5cGVfdXJsIjoiL2Nvc21vcy5jcnlwdG8uZWQyNTUxOS5QdWJLZXkiLCJ2YWx1ZSI6ImV5SnJaWGtpT2lKNmNHMVJOVU5vVnl0RWFrSklkazF5YVRkSmNqSnZZa2xOWjI5WVUwbFZXbEZ5YkRaNFRXWjFhVkp2UFNKOSJ9LCJ2b3RpbmdfcG93ZXIiOiIxMDAwMDAwMDAwMDAwMDAiLCJwcm9wb3Nlcl9wcmlvcml0eSI6IjAifX0=","LastHeightValidatorsChanged":12,"ConsensusParams":{"block":{"max_bytes":22020096,"max_gas":-1},"evidence":{"max_age_num_blocks":100000,"max_age_duration":172800000000000,"max_bytes":1048576},"validator":{"pub_key_types":["ed25519"]},"version":{"app":0},"abci":{"vote_extensions_enable_height":0}},"LastHeightConsensusParamsChanged":11,"LastResultsHash":"47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=","AppHash":"1i83m5cD/4e8D5zzYHKRsi47lIwd1G9msitiP8gV03I="}`

	appA := s.AppContext()

	statebase64 := base64.StdEncoding.EncodeToString([]byte(statestr))
	msg := []byte(fmt.Sprintf(`{"execute":{"action":{"type":"bootstrapAfterStateSync","params": [{"key":"state","value":"%s"}],"event":null}}}`, statebase64))

	// msg := []byte(fmt.Sprintf(`{"bootstrapAfterStateSync":{"state":"%s"}}`, base64.StdEncoding.EncodeToString([]byte(statestr))))

	err := networkserver.ConsensusTx(appA.App, s.App().Logger(), appA.App.GetNetworkKeeper(), msg)
	s.Require().NoError(err)
}
