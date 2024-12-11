package types

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	sdkerr "cosmossdk.io/errors"
)

var (
	regexChainID         = `[a-z0-9]{1,}`
	regexLevelSeparator  = `_{0,1}`        // optional level separator
	regexLevel           = `([0-9]*){0,1}` // optional level group
	regexEIP155Separator = `_{1}`
	regexEIP155          = `[1-9][0-9]*`
	regexEpochSeparator  = `-{1}`
	regexEpoch           = `[1-9][0-9]*`
	wasmxChainID         = regexp.MustCompile(fmt.Sprintf(`^(%s)%s%s%s(%s)%s(%s)$`, regexChainID, regexLevelSeparator, regexLevel, regexEIP155Separator, regexEIP155, regexEpochSeparator, regexEpoch))
	// ^([a-z0-9]{1,})_{0,1}([1-9][0-9]*){0,1}_{1}([1-9][0-9]*)-{1}([1-9][0-9]*)$
)

// both should work, with or without level
// mythos_8000-1
// chain0_1_10001-1

// IsValidChainID returns false if the given chain identifier is incorrectly formatted.
func IsValidChainID(chainID string) bool {
	return wasmxChainID.MatchString(chainID)
}

type ChainId struct {
	Full      string `json:"full"`
	BaseName  string `json:"base_name"`
	Level     uint32 `json:"level"`
	EvmId     uint64 `json:"evmid"`
	ForkIndex uint32 `json:"fork_index"`
}

func ParseChainID(chainId string) (*ChainId, error) {
	parts := strings.Split(chainId, "_")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid chain id: %s", chainId)
	}

	baseName := parts[0]
	level := uint32(0)
	lastPart := ""

	if len(parts) == 2 {
		lastPart = parts[1]
	} else {
		level64, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid level in chain id: %s", chainId)
		}
		level = uint32(level64)
		lastPart = parts[2]
	}

	parts2 := strings.Split(lastPart, "-")
	if len(parts2) != 2 {
		return nil, fmt.Errorf("invalid last part in chain id: %s", chainId)
	}

	evmId, err := strconv.ParseUint(parts2[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid EVM ID in chain id: %s", chainId)
	}

	forkIndex64, err := strconv.ParseUint(parts2[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid fork index in chain id: %s", chainId)
	}
	forkIndex := uint32(forkIndex64)

	return &ChainId{
		Full:      chainId,
		BaseName:  baseName,
		Level:     level,
		EvmId:     evmId,
		ForkIndex: forkIndex,
	}, nil
}

// ParseEvmChainID parses a string chain identifier's epoch to an Ethereum-compatible
// chain-id in *big.Int format. The function returns an error if the chain-id has an invalid format
func ParseEvmChainID(chainID string) (*big.Int, error) {
	chainID = strings.TrimSpace(chainID)
	if len(chainID) > 48 {
		return nil, sdkerr.Wrapf(ErrInvalidChainID, "chain-id '%s' cannot exceed 48 chars", chainID)
	}

	matches := wasmxChainID.FindStringSubmatch(chainID)
	if matches == nil || len(matches) != 5 || matches[1] == "" {
		return nil, sdkerr.Wrapf(ErrInvalidChainID, "matches for %s: %v", chainID, matches)
	}

	// verify that the chain-id entered is a base 10 integer
	chainIDInt, ok := new(big.Int).SetString(matches[3], 10)
	if !ok {
		return nil, sdkerr.Wrapf(ErrInvalidChainID, "epoch %s must be base-10 integer format", matches[3])
	}

	return chainIDInt, nil
}
