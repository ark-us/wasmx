package ibctesting

import (
	"encoding/json"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/types"

	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	ibcgotesting "github.com/cosmos/ibc-go/v8/testing"

	wasmxapp "mythos/v1/app"
	mcfg "mythos/v1/config"
	cosmosmodtypes "mythos/v1/x/cosmosmod/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

var DefaultTestingAppInit func(chainId string, index int32) (ibcgotesting.TestingApp, map[string]json.RawMessage) = wasmxapp.SetupTestingApp

// SetupWithGenesisValSet initializes a new SimApp with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit (10^6) in the default token of the simapp from first genesis
// account. A Nop logger is set in SimApp.
func SetupWithGenesisValSet(t *testing.T, valSet *tmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount, chainID string, chaincfg mcfg.ChainConfig, index int32, balances ...banktypes.Balance) (ibcgotesting.TestingApp, *abci.ResponseInitChain) {
	app, genesisState := DefaultTestingAppInit(chainID, index)
	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]cosmosmodtypes.Delegation, 0, len(valSet.Validators))
	signingInfos := make([]slashingtypes.SigningInfo, 0, len(valSet.Validators))

	bondAmt := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		require.NoError(t, err)
		pkAny, err := codectypes.NewAnyWithValue(pk)
		require.NoError(t, err)
		valAddr := sdk.ValAddress(val.Address).String()
		validator := stakingtypes.Validator{
			OperatorAddress:   valAddr,
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdkmath.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdkmath.LegacyNewDecWithPrec(3, 1), sdkmath.LegacyNewDecWithPrec(5, 1), sdkmath.LegacyNewDecWithPrec(1, 1)),
			MinSelfDelegation: sdkmath.ZeroInt(),
		}
		validators = append(validators, validator)
		delegation := cosmosmodtypes.Delegation{
			DelegatorAddress: genAccs[0].GetAddress().String(),
			ValidatorAddress: valAddr,
			Amount:           bondAmt,
		}
		// (genAccs[0].GetAddress().String(), valAddr, sdkmath.LegacyOneDec())
		delegations = append(delegations, delegation)
	}

	// set validators and delegations
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = chaincfg.BondBaseDenom
	stakingGenesis := cosmosmodtypes.NewStakingGenesisState(stakingParams, validators, delegations)

	// update total supply
	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens and delegated tokens to total supply
		totalSupply = totalSupply.Add(b.Coins.Add(sdk.NewCoin(chaincfg.BaseDenom, bondAmt))...)
	}
	bankGenesis := cosmosmodtypes.DefaultBankGenesisState(chaincfg.DenomUnit, chaincfg.BaseDenomUnit, chaincfg.Name)
	bankGenesis.Supply = totalSupply
	bankGenesis.Balances = balances

	govGenesis := cosmosmodtypes.DefaultGovGenesisState()
	govGenesis.Params.MinDeposit = sdk.NewCoins(sdk.NewCoin(chaincfg.BaseDenom, sdkmath.NewInt(1_000_000_000)))
	votingPeriod := time.Millisecond * 500
	govGenesis.Params.VotingPeriod = votingPeriod.Milliseconds()

	slashingGenesis := cosmosmodtypes.DefaultSlashingGenesisState()
	slashingGenesis.SigningInfos = signingInfos

	// set genesis accounts
	authGenesis, err := cosmosmodtypes.NewAuthGenesisStateFromCosmos(app.AppCodec(), authtypes.DefaultParams(), genAccs)
	if err != nil {
		panic(err)
	}
	cosmosmodGenesis := cosmosmodtypes.NewGenesisState(*stakingGenesis, *bankGenesis, *govGenesis, *authGenesis, *slashingGenesis, *cosmosmodtypes.DefaultDistributionGenesisState(chaincfg.BaseDenom))
	genesisState[cosmosmodtypes.ModuleName] = app.AppCodec().MustMarshalJSON(cosmosmodGenesis)

	// We are using precompiled contracts to avoid compiling at every chain instantiation
	wasmxGenesis := wasmxtypes.DefaultGenesisState()
	mydir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	wasmxGenesis.CompiledFolderPath = path.Join(mydir, "../../../", "testutil", "codes_compiled")
	genesisState[wasmxtypes.ModuleName] = app.AppCodec().MustMarshalJSON(wasmxGenesis)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	// init chain will set the validator set and initialize the genesis accounts
	resInit, err := app.InitChain(
		&abci.RequestInitChain{
			ChainId:         chainID,
			InitialHeight:   1,
			Time:            time.Now().UTC(),
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: wasmxapp.DefaultTestingConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)
	require.NoError(t, err)

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
		Time:   time.Now().UTC(),
		// Hash:               app.LastCommitID().Hash,
		// NextValidatorsHash: valSet.Hash(),
	})
	require.NoError(t, err)

	return app, resInit
}
