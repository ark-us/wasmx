package consensus

import (
	"strconv"

	consensus "github.com/loredanacirstea/wasmx-env-consensus/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func GetLeaderChain(chainIds []string) string {
	if len(chainIds) == 0 {
		return ""
	}
	if len(chainIds) == 1 {
		return chainIds[0]
	}
	higherChain := chainIds[0]
	var higherLevel uint32
	for _, idstr := range chainIds {
		id, err := ChainIdFromString(idstr)
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

func BuildChainConfig(denomUnit string, baseDenomUnit uint32, chainBaseName string) ChainConfig {
	bech32PrefixAccAddr := chainBaseName
	bech32PrefixAccPub := chainBaseName + "pub"
	bech32PrefixValAddr := chainBaseName
	bech32PrefixValPub := chainBaseName + "pub"
	bech32PrefixConsAddr := chainBaseName
	bech32PrefixConsPub := chainBaseName + "pub"
	name := chainBaseName
	humanCoinUnit := denomUnit
	baseDenom := "a" + denomUnit
	denomUnitStr := denomUnit
	bondBaseDenom := "as" + denomUnit
	bondDenom := "s" + denomUnit

	return ChainConfig{
		Bech32PrefixAccAddr:  bech32PrefixAccAddr,
		Bech32PrefixAccPub:   bech32PrefixAccPub,
		Bech32PrefixValAddr:  bech32PrefixValAddr,
		Bech32PrefixValPub:   bech32PrefixValPub,
		Bech32PrefixConsAddr: bech32PrefixConsAddr,
		Bech32PrefixConsPub:  bech32PrefixConsPub,
		Name:                 name,
		HumanCoinUnit:        humanCoinUnit,
		BaseDenom:            baseDenom,
		DenomUnit:            denomUnitStr,
		BaseDenomUnit:        baseDenomUnit,
		BondBaseDenom:        bondBaseDenom,
		BondDenom:            bondDenom,
	}
}

func GetDefaultConsensusParams() consensus.ConsensusParams {
	block := consensus.BlockParams{
		MaxBytes: 22020096,
		MaxGas:   30000000,
	}
	evidence := consensus.EvidenceParams{
		MaxAgeNumBlocks: 100000,
		MaxAgeDuration:  172800000000000,
		MaxBytes:        1048576,
	}
	validator := consensus.ValidatorParams{PubKeyTypes: []string{"ed25519"}}
	version := consensus.VersionParams{App: 0}
	abci := consensus.ABCIParams{VoteExtensionsEnableHeight: 0}
	return consensus.ConsensusParams{Block: block, Evidence: evidence, Validator: validator, Version: version, ABCI: abci}
}

// Helpers to create events
func NewEventAttribute(key, value string, index bool) wasmx.EventAttribute {
	return wasmx.EventAttribute{Key: key, Value: value, Index: index}
}
func NewEvent(typ string, attrs ...wasmx.EventAttribute) wasmx.Event {
	return wasmx.Event{Type: typ, Attributes: attrs}
}

// String helpers
func Itoa(i int) string      { return strconv.FormatInt(int64(i), 10) }
func U64toa(i uint64) string { return strconv.FormatUint(i, 10) }
