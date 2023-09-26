package ibctesting

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cometbft/cometbft/version"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/cosmos/ibc-go/v8/testing/mock"

	wasmxapp "mythos/v1/app"
)

// ChainIDPrefix defines the default chain ID prefix for Mythos test chains
var ChainIDPrefix = "mythos_7000-"

func init() {
	ibcgotesting.ChainIDPrefix = ChainIDPrefix
}

// NewTestChain initializes a new TestChain instance with a single validator set using a
// generated private key. It also creates a sender account to be used for delivering transactions.
//
// The first block height is committed to state in order to allow for client creations on
// counterparty chains. The TestChain will return with a block height starting at 2.
//
// Time management is handled by the Coordinator in order to ensure synchrony between chains.
// Each update of any chain increments the block header time for all chains by 5 seconds.
func NewTestChain(t *testing.T, coord *ibcgotesting.Coordinator, chainID string) *ibcgotesting.TestChain {
	// generate validator private/public key
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	signersByAddress := make(map[string]tmtypes.PrivValidator, 1)
	signersByAddress[pubKey.Address().String()] = privVal

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)

	senderAddress := common.BytesToAddress(senderPrivKey.PubKey().Address().Bytes())

	amount := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)

	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(wasmxapp.BondDenom, amount)),
	}

	app := SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{acc}, chainID, balance)

	consAddress := sdk.ConsAddress(senderPrivKey.PubKey().Address())

	// create current header and call begin block
	header := tmproto.Header{
		ChainID:         chainID,
		Height:          1,
		Time:            coord.CurrentTime.UTC(),
		ProposerAddress: consAddress.Bytes(),
		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	}

	txConfig := app.GetTxConfig()

	// create an account to send transactions from
	chain := &ibcgotesting.TestChain{
		T:             t,
		Coordinator:   coord,
		ChainID:       chainID,
		App:           app,
		CurrentHeader: header,
		QueryServer:   app.GetIBCKeeper(),
		TxConfig:      txConfig,
		Codec:         app.AppCodec(),
		Vals:          valSet,
		NextVals:      valSet,
		Signers:       signersByAddress,
		SenderPrivKey: senderPrivKey,
		SenderAccount: acc,
	}

	mapp, ok := app.(*wasmxapp.App)
	require.True(t, ok, "not app")
	ctx := chain.GetContext()
	mapp.AccountKeeper.SetAccount(ctx, acc)

	valAddr := sdk.ValAddress(senderAddress.Bytes())
	_validator, err := stakingtypes.NewValidator(valAddr, senderPrivKey.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	_validator = stakingkeeper.TestingUpdateValidator(mapp.StakingKeeper, ctx, _validator, true)
	mapp.StakingKeeper.AfterValidatorCreated(ctx, _validator.GetOperator())

	err = mapp.StakingKeeper.SetValidatorByConsAddr(ctx, _validator)
	require.NoError(t, err)
	mapp.StakingKeeper.SetValidator(ctx, _validator)

	coord.CommitBlock(chain)

	return chain
}

func NewTransferPath(chainA, chainB *ibcgotesting.TestChain) *ibcgotesting.Path {
	path := ibcgotesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibcgotesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibcgotesting.TransferPort

	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = "ics20-1"
	path.EndpointB.ChannelConfig.Version = "ics20-1"

	return path
}
