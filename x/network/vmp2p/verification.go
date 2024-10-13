package vmp2p

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	cmttypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"

	mcodec "mythos/v1/codec"
	mcfg "mythos/v1/config"
	networkserver "mythos/v1/x/network/server"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

type ClientVerification struct {
	Address   *mcodec.AccAddressPrefixed
	logger    log.Logger
	mythosapp mcfg.MythosApp
}

func NewClientVerification(mythosapp mcfg.MythosApp, logger log.Logger, addr *mcodec.AccAddressPrefixed) ClientVerification {
	return ClientVerification{
		mythosapp: mythosapp,
		logger:    logger,
		Address:   addr,
	}
}

func (v ClientVerification) VerifyCommitLight(chainID string, blockID cmttypes.BlockID, height int64, commit *cmttypes.Commit, valset *cmttypes.ValidatorSet) error {
	fmt.Println("-ChainVerification.VerifyCommit-", v.Address.String(), chainID, height, blockID.Hash, commit.Round, len(commit.Signatures))

	if v.Address == nil {
		return valset.VerifyCommitLight(chainID, blockID, height, commit)
	}

	return v.VerifyCommitLightByContract(chainID, blockID, height, commit, valset)
}

func (v ClientVerification) VerifyCommitLightByContract(chainID string, blockID cmttypes.BlockID, height int64, commit *cmttypes.Commit, valset *cmttypes.ValidatorSet) error {
	data := VerifyCommitLightRequest{
		ChainId: chainID,
		BlockID: blockID,
		Commit:  *commit,
		ValSet:  *valset,
	}
	databz, err := json.Marshal(data)
	if err != nil {
		return err
	}

	msg := []byte(fmt.Sprintf(`{"execute":{"action":{"type":"VerifyCommitLight","params": [{"key":"data","value":"%s"}],"event":null}}}`, base64.StdEncoding.EncodeToString(databz)))

	res, err := networkserver.ConsensusQuery(v.mythosapp, v.logger, v.mythosapp.GetNetworkKeeper(), msg)
	if err != nil {
		return err
	}

	var cresp wasmxtypes.ContractResponse
	err = json.Unmarshal(res.Data, &cresp)
	if err != nil {
		return err
	}

	var resp VerifyCommitLightResponse
	err = json.Unmarshal(cresp.Data, &resp)
	if err != nil {
		return err
	}

	if resp.Valid && len(resp.Error) == 0 {
		return nil
	}
	if len(resp.Error) > 0 {
		return fmt.Errorf(resp.Error)
	}
	return fmt.Errorf("commit verification failed without error")
}
