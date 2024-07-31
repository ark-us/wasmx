package vmp2p

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cometbftenc "github.com/cometbft/cometbft/crypto/encoding"
	cmtstate "github.com/cometbft/cometbft/proto/tendermint/state"
	sm "github.com/cometbft/cometbft/state"
	cmttypes "github.com/cometbft/cometbft/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"

	networkserver "mythos/v1/x/network/server"
	networktypes "mythos/v1/x/network/types"

	mcfg "mythos/v1/config"
	mctx "mythos/v1/context"
	vmmc "mythos/v1/x/network/vmmc"
)

type StateStore struct {
	ctx *Context
}

type State struct {
	Version cmtstate.Version

	// immutable
	ChainID       string
	InitialHeight int64 // should be 1, not 0, when starting from height 1

	// LastBlockHeight=0 at genesis (ie. block(H=0) does not exist)
	LastBlockHeight int64
	LastBlockID     cmttypes.BlockID
	LastBlockTime   time.Time

	// LastValidators is used to validate block.LastCommit.
	// Validators are persisted to the database separately every time they change,
	// so we can query for historical validator sets.
	// Note that if s.LastBlockHeight causes a valset change,
	// we set s.LastHeightValidatorsChanged = s.LastBlockHeight + 1 + 1
	// Extra +1 due to nextValSet delay.
	NextValidators              []byte
	Validators                  []byte
	LastValidators              []byte
	LastHeightValidatorsChanged int64

	// Consensus parameters used for validating blocks.
	// Changes returned by FinalizeBlock and updated after Commit.
	ConsensusParams                  cmttypes.ConsensusParams
	LastHeightConsensusParamsChanged int64

	// Merkle root of the results from executing prev block
	LastResultsHash []byte

	// the latest AppHash we've received from calling abci.Commit()
	AppHash []byte
}

// LoadFromDBOrGenesisFile loads the most recent state.
// If the chain is new it will use the genesis file from the provided genesis file path as the current state.
func (s StateStore) LoadFromDBOrGenesisFile(string) (sm.State, error) {
	fmt.Println("---StateStore.LoadFromDBOrGenesisFile--")
	return sm.State{}, nil
}

// LoadFromDBOrGenesisDoc loads the most recent state.
// If the chain is new it will use the genesis doc as the current state.
func (s StateStore) LoadFromDBOrGenesisDoc(*cmttypes.GenesisDoc) (sm.State, error) {
	fmt.Println("---StateStore.LoadFromDBOrGenesisDoc--")
	return sm.State{}, nil
}

// Load loads the current state of the blockchain
func (s StateStore) Load() (sm.State, error) {
	fmt.Println("---StateStore.Load--")
	return sm.State{}, nil
}

// LoadValidators loads the validator set at a given height
func (s StateStore) LoadValidators(int64) (*cmttypes.ValidatorSet, error) {
	fmt.Println("---StateStore.LoadValidators--")
	return nil, nil
}

// LoadFinalizeBlockResponse loads the abciResponse for a given height
func (s StateStore) LoadFinalizeBlockResponse(int64) (*abci.ResponseFinalizeBlock, error) {
	fmt.Println("---StateStore.LoadFinalizeBlockResponse--")
	return nil, nil
}

// LoadLastABCIResponse loads the last abciResponse for a given height
func (s StateStore) LoadLastFinalizeBlockResponse(int64) (*abci.ResponseFinalizeBlock, error) {
	fmt.Println("---StateStore.LoadLastFinalizeBlockResponse--")
	return nil, nil
}

// LoadConsensusParams loads the consensus params for a given height
func (s StateStore) LoadConsensusParams(int64) (cmttypes.ConsensusParams, error) {
	fmt.Println("---StateStore.LoadConsensusParams--")
	return cmttypes.ConsensusParams{}, nil
}

// Save overwrites the previous state with the updated one
func (s StateStore) Save(sm.State) error {
	fmt.Println("---StateStore.Save--")
	return nil
}

// SaveFinalizeBlockResponse saves ABCIResponses for a given height
func (s StateStore) SaveFinalizeBlockResponse(int64, *abci.ResponseFinalizeBlock) error {
	fmt.Println("---StateStore.SaveFinalizeBlockResponse--")
	return nil
}

// Bootstrap is used for bootstrapping state when not starting from a initial height.
func (s StateStore) Bootstrap(state sm.State) error {
	fmt.Println("---StateStore.Bootstrap--", state)
	// bootstrap consensus contract
	// this will update storage contract too

	nextValidators, err := cmtValidatorSetToWasmxValidatorSet(s.ctx.Context.CosmosHandler.Codec().InterfaceRegistry(), state.NextValidators)
	if err != nil {
		return err
	}
	nextValidatorsBz, err := s.ctx.Context.CosmosHandler.JSONCodec().MarshalJSON(nextValidators)
	if err != nil {
		return err
	}
	validators, err := cmtValidatorSetToWasmxValidatorSet(s.ctx.Context.CosmosHandler.Codec().InterfaceRegistry(), state.Validators)
	if err != nil {
		return err
	}
	validatorsBz, err := s.ctx.Context.CosmosHandler.JSONCodec().MarshalJSON(validators)
	if err != nil {
		return err
	}
	lastValidators, err := cmtValidatorSetToWasmxValidatorSet(s.ctx.Context.CosmosHandler.Codec().InterfaceRegistry(), state.LastValidators)
	if err != nil {
		return err
	}
	lastValidatorsBz, err := s.ctx.Context.CosmosHandler.JSONCodec().MarshalJSON(lastValidators)
	if err != nil {
		return err
	}

	cstate := &State{
		Version:                          state.Version,
		ChainID:                          state.ChainID,
		InitialHeight:                    state.InitialHeight,
		LastBlockHeight:                  state.LastBlockHeight,
		LastBlockID:                      state.LastBlockID,
		LastBlockTime:                    state.LastBlockTime,
		NextValidators:                   nextValidatorsBz,
		Validators:                       validatorsBz,
		LastValidators:                   lastValidatorsBz,
		LastHeightValidatorsChanged:      state.LastHeightValidatorsChanged,
		ConsensusParams:                  state.ConsensusParams,
		LastHeightConsensusParamsChanged: state.LastHeightConsensusParamsChanged,
		LastResultsHash:                  state.LastResultsHash,
		AppHash:                          state.AppHash,
	}

	statebz, err := json.Marshal(cstate)
	fmt.Println("---StateStore.Bootstrap--", err, string(statebz))
	if err != nil {
		return err
	}

	chainCfg, err := mcfg.GetChainConfig(s.ctx.Context.Ctx.ChainID())
	if err != nil {
		return err
	}
	// mapp, ok := s.ctx.Context.GetApplication().(mcfg.MythosApp)

	// for i, chainId := range msrvconfig.Network.InitialChains {
	// }

	// portOffset := int32(nodeOffset * uint32(len(initialChains)))
	portOffset := int32(1 * uint32(2))
	nodePorts := mctx.GetInitialChainNodePorts(1, portOffset)

	ctx := &vmmc.Context{Context: s.ctx.Context}
	req := &vmmc.StartSubChainMsg{
		ChainId:     s.ctx.Context.Ctx.ChainID(),
		ChainConfig: *chainCfg,
		NodePorts:   nodePorts,
	}
	// vmmc.StartApp(ctx, req)

	multichainapp, err := mcfg.GetMultiChainApp(ctx.GoContextParent)
	if err != nil {
		return err
	}
	var app mcfg.MythosApp
	found := false
	iapp, err := multichainapp.GetApp(req.ChainId)
	if err == nil {
		app_, ok := iapp.(mcfg.MythosApp)
		if ok {
			app = app_
			found = true
		}
	}
	if !found {
		app = multichainapp.NewApp(req.ChainId, &req.ChainConfig)
	}

	msg := []byte(fmt.Sprintf(`{"execute":{"action":{"type":"bootstrapAfterStateSync","params": [{"key":"state","value":"%s"}],"event":null}}}`, base64.StdEncoding.EncodeToString(statebz)))
	err = networkserver.ConsensusTx(app, s.ctx.Logger, app.GetNetworkKeeper(), msg)
	if err != nil {
		return err
	}

	err = networkserver.StartNode(app, s.ctx.Logger, app.GetNetworkKeeper())
	if err != nil {
		return err
	}

	return nil
}

// PruneStates takes the height from which to start pruning and which height stop at
func (s StateStore) PruneStates(int64, int64, int64) error {
	fmt.Println("---StateStore.PruneStates--")
	return nil
}

// Saves the height at which the store is bootstrapped after out of band statesync
func (s StateStore) SetOfflineStateSyncHeight(height int64) error {
	fmt.Println("---StateStore.SetOfflineStateSyncHeight--")
	return nil
}

// Gets the height at which the store is bootstrapped after out of band statesync
func (s StateStore) GetOfflineStateSyncHeight() (int64, error) {
	fmt.Println("---StateStore.GetOfflineStateSyncHeight--")
	return 0, nil
}

// Close closes the connection with the database
func (s StateStore) Close() error {
	fmt.Println("---StateStore.Close--")
	return nil
}

func cmtValidatorSetToWasmxValidatorSet(interfaceRegistry cdctypes.InterfaceRegistry, valset *cmttypes.ValidatorSet) (*networktypes.ValidatorSet, error) {
	validators, err := cmtValidatorsToWasmxValidators(interfaceRegistry, valset.Validators)
	if err != nil {
		return nil, err
	}
	proposer, err := cmtValidatorToWasmxValidator(interfaceRegistry, valset.Proposer)
	if err != nil {
		return nil, err
	}
	return &networktypes.ValidatorSet{
		Validators: validators,
		Proposer:   *proposer,
	}, nil
}

func cmtValidatorsToWasmxValidators(interfaceRegistry cdctypes.InterfaceRegistry, vals []*cmttypes.Validator) ([]networktypes.TendermintValidator, error) {
	cmtvals := make([]networktypes.TendermintValidator, len(vals))
	for i, val := range vals {
		newval, err := cmtValidatorToWasmxValidator(interfaceRegistry, val)
		if err != nil {
			return nil, err
		}
		cmtvals[i] = *newval
	}
	return cmtvals, nil
}

func cmtValidatorToWasmxValidator(interfaceRegistry cdctypes.InterfaceRegistry, val *cmttypes.Validator) (*networktypes.TendermintValidator, error) {
	keyProto, err := cometbftenc.PubKeyToProto(val.PubKey)
	if err != nil {
		return nil, err
	}
	pubkey, err := cryptocodec.FromCmtProtoPublicKey(keyProto)
	if err != nil {
		return nil, err
	}
	keyAny, err := cdctypes.NewAnyWithValue(pubkey)
	if err != nil {
		return nil, err
	}

	v := &networktypes.TendermintValidator{
		Address:          hex.EncodeToString(val.Address),
		PubKey:           keyAny,
		VotingPower:      val.VotingPower,
		ProposerPriority: val.ProposerPriority,
	}
	return v, nil
}
