package lib

// Defaults for slashing module (converted from AS defaults.ts)

const (
	SignedBlocksWindow      = int64(10000)
	MinSignedPerWindow      = "0.500000000000000000"
	DowntimeJailDuration    = "600s" // 10 min
	SlashFractionDoubleSign = "0.000000000000000000"
	SlashFractionDowntime   = "0.000000000000000000"
)

func GetDefaultParams() Params {
	return Params{
		SignedBlocksWindow:      SignedBlocksWindow,
		MinSignedPerWindow:      MinSignedPerWindow,
		DowntimeJailDuration:    DowntimeJailDuration,
		SlashFractionDoubleSign: SlashFractionDoubleSign,
		SlashFractionDowntime:   SlashFractionDowntime,
	}
}

func GetDefaultGenesis() GenesisState {
	p := GetDefaultParams()
	return GenesisState{
		Params:       p,
		SigningInfos: []SigningInfo{},
		MissedBlocks: []ValidatorMissedBlocks{},
	}
}
