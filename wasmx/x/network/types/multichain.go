package types

import (
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// GetLeaderChain determines the chain with the highest level from a list of chain IDs
func GetLeaderChain(chainIds []string) string {
	if len(chainIds) == 0 {
		return ""
	}
	if len(chainIds) == 1 {
		return chainIds[0]
	}

	higherChain := chainIds[0]
	var higherLevel uint32 = 0

	for _, chainId := range chainIds {
		id, err := wasmxtypes.ParseChainID(chainId)
		if err != nil {
			continue
		}
		if higherLevel < id.Level {
			higherLevel = id.Level
			higherChain = id.Full
		}
	}

	return higherChain
}
