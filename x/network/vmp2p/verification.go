package vmp2p

import (
	"fmt"

	cmttypes "github.com/cometbft/cometbft/types"

	mcodec "mythos/v1/codec"
)

type ClientVerification struct {
	Address *mcodec.AccAddressPrefixed
}

func NewClientVerification(addr *mcodec.AccAddressPrefixed) ClientVerification {
	return ClientVerification{Address: addr}
}

func (v ClientVerification) VerifyCommitLight(chainID string, blockID cmttypes.BlockID, height int64, commit *cmttypes.Commit, valset *cmttypes.ValidatorSet) error {
	fmt.Println("-ChainVerification.VerifyCommit-", v.Address.String(), chainID, height, blockID.Hash, commit.Round, len(commit.Signatures))

	if v.Address == nil {
		return valset.VerifyCommitLight(chainID, blockID, height, commit)
	}

	// TODO call the verification contract
	return nil
}
