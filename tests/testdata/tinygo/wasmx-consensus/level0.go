package consensus

const (
	Level0ChainIdFull    = "level0_1000-1"
	Bech32PrefixAccAddr  = "level0"
	Bech32PrefixAccPub   = "level0pub"
	Bech32PrefixValAddr  = "level0"
	Bech32PrefixValPub   = "level0pub"
	Bech32PrefixConsAddr = "level0"
	Bech32PrefixConsPub  = "level0pub"
	Name                 = "level0"
	HumanCoinUnit        = "lvl"
	BaseDenom            = "alvl"
	DenomUnit            = "lvl"
	BaseDenomUnit        = 18
	BondBaseDenom        = "aslvl"
	BondDenom            = "slvl"
)

var Level0Config = ChainConfig{
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
