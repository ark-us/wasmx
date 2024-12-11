package vmp2p

import (
	"encoding/json"
	"fmt"

	sm "github.com/cometbft/cometbft/state"
	cmttypes "github.com/cometbft/cometbft/types"
)

type BlockStore struct{}

func (s BlockStore) Base() int64 {
	return 0
}

func (s BlockStore) Height() int64 {
	return 0
}

func (s BlockStore) Size() int64 {
	return 0
}

func (s BlockStore) LoadBaseMeta() *cmttypes.BlockMeta {
	return nil
}

func (s BlockStore) LoadBlockMeta(height int64) *cmttypes.BlockMeta {
	return nil
}

func (s BlockStore) LoadBlock(height int64) *cmttypes.Block {
	return nil
}

func (s BlockStore) SaveBlock(block *cmttypes.Block, blockParts *cmttypes.PartSet, seenCommit *cmttypes.Commit) {

}

func (s BlockStore) SaveBlockWithExtendedCommit(block *cmttypes.Block, blockParts *cmttypes.PartSet, seenCommit *cmttypes.ExtendedCommit) {

}

func (s BlockStore) PruneBlocks(height int64, state sm.State) (uint64, int64, error) {
	return 0, 0, nil
}

func (s BlockStore) LoadBlockByHash(hash []byte) *cmttypes.Block {
	return nil
}

func (s BlockStore) LoadBlockMetaByHash(hash []byte) *cmttypes.BlockMeta {
	return nil
}

func (s BlockStore) LoadBlockPart(height int64, index int) *cmttypes.Part {
	return nil
}

func (s BlockStore) LoadBlockCommit(height int64) *cmttypes.Commit {
	return nil
}

func (s BlockStore) LoadSeenCommit(height int64) *cmttypes.Commit {
	return nil
}

func (s BlockStore) LoadBlockExtendedCommit(height int64) *cmttypes.ExtendedCommit {
	return nil
}

func (s BlockStore) DeleteLatestBlock() error {
	return nil
}

func (s BlockStore) Close() error {
	return nil
}

func (s BlockStore) SaveSeenCommit(height int64, seenCommit *cmttypes.Commit) error {
	seenCommitbz, err := json.Marshal(seenCommit)
	fmt.Println("---BlockStore.SaveSeenCommit--", err, string(seenCommitbz))
	if err != nil {
		return err
	}
	// TODO commitAfterStateSync
	return nil
}
