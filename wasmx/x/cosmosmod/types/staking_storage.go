package types

import (
	"math/big"
	"strconv"
	"strings"
)

// storage keys for derc20
const STAKING_SPLIT = "."
const STAKING_DELEGATOR_TO_VALIDATORS_KEY = "delegator_to_validators."
const STAKING_DELEGATOR_TO_DELEGATION_KEY = "delegator_to_delegation."
const STAKING_VALIDATOR_TO_DELEGATORS_KEY = "validator_to_delegators."
const STAKING_VALIDATOR_DELEGATION_KEY = "validator_delegation."

type MsgJail struct {
	ConsensusAddress string `json:"consaddr"`
}

type MsgUnjail struct {
	ConsensusAddress string `json:"consaddr"`
}

// key_delegator_validator => amount
func ParseStoredDelegation(key []byte, value []byte) (delegator string, validator string, amount *big.Int, err error) {
	parts := strings.Split(string(key), STAKING_SPLIT)
	delegator = parts[1]
	validator = parts[2]
	_amount, err := strconv.ParseInt(string(value), 10, 64)
	amount = big.NewInt(_amount)
	return delegator, validator, amount, err
}
