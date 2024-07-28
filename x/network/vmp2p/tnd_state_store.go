package vmp2p

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	sm "github.com/cometbft/cometbft/state"
	cmttypes "github.com/cometbft/cometbft/types"
)

type StateStore struct{}

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
func (s StateStore) Bootstrap(sm.State) error {
	fmt.Println("---StateStore.Bootstrap--")
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
