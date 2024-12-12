package config

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	menc "github.com/loredanacirstea/wasmx/encoding"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

const FEE_COLLECTOR = "fee_collector"

const (
	Bech32Prefix  = "mythos"
	Name          = "mythos"
	HumanCoinUnit = "myt"
	BaseDenom     = "amyt"
	DenomUnit     = "myt"
	BaseDenomUnit = 18
	BondBaseDenom = "asmyt"
	BondDenom     = "smyt"
)

// PowerReduction defines the default power reduction value for staking
var PowerReduction = sdkmath.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(BaseDenomUnit), nil))

var (
	// Bech32PrefixAccAddr defines the Bech32 prefix of an account's address
	Bech32PrefixAccAddr = Bech32Prefix
	// Bech32PrefixAccPub defines the Bech32 prefix of an account's public key
	Bech32PrefixAccPub = Bech32Prefix + sdk.PrefixPublic
	// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
	Bech32PrefixValAddr = Bech32Prefix
	// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
	Bech32PrefixValPub = Bech32Prefix + sdk.PrefixPublic
	// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
	Bech32PrefixConsAddr = Bech32Prefix
	// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
	Bech32PrefixConsPub = Bech32Prefix + sdk.PrefixPublic
)

// SetBech32Prefixes sets the global prefixes to be used when serializing addresses and public keys to Bech32 strings.
func SetBech32Prefixes(config *sdk.Config, newcfg menc.ChainConfig) {
	config.SetBech32PrefixForAccount(newcfg.Bech32PrefixAccAddr, newcfg.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(newcfg.Bech32PrefixValAddr, newcfg.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(newcfg.Bech32PrefixConsAddr, newcfg.Bech32PrefixConsPub)
}

func SetGlobalChainConfigById(chainId string) error {
	cfg, ok := PrefixesMap[chainId]
	if !ok {
		return fmt.Errorf("chain_id configuration not found: %s", chainId)
	}
	return SetGlobalChainConfig(chainId, cfg)
}

func SetGlobalChainConfig(chainId string, cfg menc.ChainConfig) error {
	config := sdk.GetConfig()
	// TODO rewrite cosmos
	sdk.SetAddrCacheEnabled(false)
	SetBech32Prefixes(config, cfg)
	return nil
}

var PrefixesMap = map[string]menc.ChainConfig{}
var ChainIdsInit = []string{}

func GetChainConfig(chainId string) (*menc.ChainConfig, error) {
	conf, ok := PrefixesMap[chainId]
	if !ok {
		chainIdParsed, err := wasmxtypes.ParseChainID(chainId)
		if err != nil {
			return nil, err
		}
		conf, ok = PrefixesMap[chainIdParsed.BaseName]
		if ok {
			CacheChainConfig(chainId, conf)
		}
	}
	if !ok {
		return nil, fmt.Errorf("chain_id configuration not found: %s", chainId)
	}
	return &conf, nil
}

func CacheChainConfig(chainId string, conf menc.ChainConfig) {
	PrefixesMap[chainId] = conf
}

var LEVEL0_BASE_NAME = "level0"
var LEVEL0_CHAIN_ID = "level0_1000-1"
var MYTHOS_BASE_NAME = "mythos"
var MYTHOS_CHAIN_ID_TEST = "mythos_7001-1"
var MYTHOS_CHAIN_ID_TESTNET = "mythos_7000-14"
var DEFAULT_MYTHOS_CONFIG = menc.ChainConfig{
	Bech32PrefixAccAddr:  Bech32PrefixAccAddr,
	Bech32PrefixAccPub:   Bech32PrefixAccPub,
	Bech32PrefixValAddr:  Bech32PrefixValAddr,
	Bech32PrefixValPub:   Bech32PrefixValPub,
	Bech32PrefixConsAddr: Bech32PrefixConsAddr,
	Bech32PrefixConsPub:  Bech32PrefixConsPub,
	Name:                 Name,
	HumanCoinUnit:        HumanCoinUnit,
	BaseDenom:            BaseDenom,
	DenomUnit:            DenomUnit,
	BaseDenomUnit:        BaseDenomUnit,
	BondBaseDenom:        BondBaseDenom,
	BondDenom:            BondDenom,
}
var DEFAULT_LEVEL0_CONFIG = menc.ChainConfig{
	Bech32PrefixAccAddr:  "level0",
	Bech32PrefixAccPub:   "level0pub",
	Bech32PrefixValAddr:  "level0",
	Bech32PrefixValPub:   "level0pub",
	Bech32PrefixConsAddr: "level0",
	Bech32PrefixConsPub:  "level0pub",
	Name:                 "level0",
	HumanCoinUnit:        "lvl",
	BaseDenom:            "alvl",
	DenomUnit:            "lvl",
	BaseDenomUnit:        18,
	BondBaseDenom:        "aslvl",
	BondDenom:            "slvl",
}

// TODO this needs to be in a contract
func init() {
	// these chains are initialized by the testnet
	ChainIdsInit = []string{
		MYTHOS_CHAIN_ID_TESTNET,
		LEVEL0_CHAIN_ID,
	}
	PrefixesMap[MYTHOS_BASE_NAME] = DEFAULT_MYTHOS_CONFIG
	PrefixesMap[MYTHOS_CHAIN_ID_TEST] = DEFAULT_MYTHOS_CONFIG
	PrefixesMap[MYTHOS_CHAIN_ID_TESTNET] = DEFAULT_MYTHOS_CONFIG
	PrefixesMap[LEVEL0_BASE_NAME] = DEFAULT_LEVEL0_CONFIG
	PrefixesMap[LEVEL0_CHAIN_ID] = DEFAULT_LEVEL0_CONFIG
}

func GetMultiChainStoreKey(chainId string, storeKey string) string {
	return chainId + "_" + storeKey
}
