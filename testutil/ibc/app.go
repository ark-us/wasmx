package ibctesting

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/rand"
	tmtypes "github.com/cometbft/cometbft/types"

	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	ibcgotesting "github.com/cosmos/ibc-go/v8/testing"

	wasmxapp "mythos/v1/app"
	mcodec "mythos/v1/codec"
	mcfg "mythos/v1/config"
	menc "mythos/v1/encoding"
	cosmosmodtypes "mythos/v1/x/cosmosmod/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

var DefaultTestingAppInit func(chainId string, chainCfg *menc.ChainConfig, index int32) (ibcgotesting.TestingApp, map[string]json.RawMessage) = wasmxapp.SetupTestingApp

// SetupWithGenesisValSet initializes a new SimApp with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit (10^6) in the default token of the simapp from first genesis
// account. A Nop logger is set in SimApp.
func SetupWithGenesisValSet(t *testing.T, valSet *tmtypes.ValidatorSet, genAccs []cosmosmodtypes.GenesisAccount, chainID string, chaincfg menc.ChainConfig, index int32, balances ...banktypes.Balance) (ibcgotesting.TestingApp, *abci.ResponseInitChain) {
	app, genesisState, err := BuildGenesisData(valSet, genAccs, chainID, chaincfg, index, balances)
	require.NoError(t, err)

	return InitAppChain(t, app, genesisState, chainID)
}

func InitAppChain(t *testing.T, app ibcgotesting.TestingApp, genesisState map[string]json.RawMessage, chainID string) (ibcgotesting.TestingApp, *abci.ResponseInitChain) {
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	req := &abci.RequestInitChain{
		ChainId:         chainID,
		InitialHeight:   1,
		Time:            time.Now().UTC(),
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: wasmxapp.DefaultTestingConsensusParams,
		AppStateBytes:   stateBytes,
	}

	// init chain will set the validator set and initialize the genesis accounts
	resInit, err := app.InitChain(req)
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

func BuildGenesisData(valSet *tmtypes.ValidatorSet, genAccs []cosmosmodtypes.GenesisAccount, chainID string, chaincfg menc.ChainConfig, index int32, balances []banktypes.Balance) (ibcgotesting.TestingApp, map[string]json.RawMessage, error) {
	app, genesisState := DefaultTestingAppInit(chainID, &chaincfg, index)
	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]cosmosmodtypes.Delegation, 0, len(valSet.Validators))
	signingInfos := make([]slashingtypes.SigningInfo, 0, len(valSet.Validators))

	bondAmt := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)

	for i, val := range valSet.Validators {
		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		if err != nil {
			return app, nil, err
		}
		pkAny, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			return app, nil, err
		}
		valAddr := genAccs[i].GetAddressPrefixed().String()
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

		delAddr := genAccs[0].GetAddressPrefixed().String()
		delegation := cosmosmodtypes.Delegation{
			DelegatorAddress: delAddr,
			ValidatorAddress: valAddr,
			Amount:           bondAmt,
		}
		delegations = append(delegations, delegation)
	}

	// set validators and delegations
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = chaincfg.BondBaseDenom
	stakingGenesis := cosmosmodtypes.NewStakingGenesisState(stakingParams, validators, delegations, chaincfg.BaseDenom)

	// update total supply
	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens and delegated tokens to total supply
		totalSupply = totalSupply.Add(b.Coins.Add(sdk.NewCoin(chaincfg.BaseDenom, bondAmt))...)
	}
	bankGenesis := cosmosmodtypes.DefaultBankGenesisState(chaincfg.DenomUnit, chaincfg.BaseDenomUnit)
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
		return app, nil, err
	}
	rewardDenom := bankGenesis.DenomInfo[2].Metadata.Base
	distributionGenesis := cosmosmodtypes.DefaultDistributionGenesisState(chaincfg.BaseDenom, rewardDenom)
	cosmosmodGenesis := cosmosmodtypes.NewGenesisState(*stakingGenesis, *bankGenesis, *govGenesis, *authGenesis, *slashingGenesis, *distributionGenesis)
	genesisState[cosmosmodtypes.ModuleName] = app.AppCodec().MustMarshalJSON(cosmosmodGenesis)

	addrCodec := mcodec.NewAccBech32Codec(chaincfg.Bech32PrefixAccAddr, mcodec.NewAddressPrefixedFromAcc)

	// We are using precompiled contracts to avoid compiling at every chain instantiation

	feeCollector, err := addrCodec.BytesToString(authtypes.NewModuleAddress(mcfg.FEE_COLLECTOR))
	if err != nil {
		return app, nil, err
	}
	mintAddress, err := addrCodec.BytesToString(authtypes.NewModuleAddress("mint"))
	if err != nil {
		return app, nil, err
	}
	bootstrapAccount, err := addrCodec.BytesToString(sdk.AccAddress(rand.Bytes(address.Len)))
	if err != nil {
		return app, nil, err
	}

	wasmxGenesis := wasmxtypes.DefaultGenesisState(bootstrapAccount, feeCollector, mintAddress, 1, false, "{}")
	// mydir, err := os.Getwd()
	// if err != nil {
	// 	panic(err)
	// }
	// wasmxGenesis.CompiledFolderPath = path.Join(mydir, "../../../", "testutil", "codes_compiled")
	genesisState[wasmxtypes.ModuleName] = app.AppCodec().MustMarshalJSON(wasmxGenesis)

	return app, genesisState, nil
}
