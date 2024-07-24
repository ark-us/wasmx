package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simulation "github.com/cosmos/cosmos-sdk/types/simulation"

	cmttypes "github.com/cometbft/cometbft/types"

	mcfg "mythos/v1/config"

	// networkserver "mythos/v1/x/network/server"
	testdata "mythos/v1/x/network/keeper/testdata/wasmx"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestBlockHeader() {
	wasmbinTo := testdata.WasmxSimpleStorage
	chainId := mcfg.MYTHOS_CHAIN_ID_TEST
	suite.SetCurrentChain(chainId)
	chain := suite.GetChain(chainId)

	initBalance := sdkmath.NewInt(10_000_000_000)
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
