package wasmx

import "encoding/json"

type MintMinter struct {
	Inflation        string `json:"inflation"`
	AnnualProvisions string `json:"annual_provisions"`
}

type MintParams struct {
	MintDenom           string `json:"mint_denom"`
	InflationRateChange string `json:"inflation_rate_change"`
	InflationMax        string `json:"inflation_max"`
	InflationMin        string `json:"inflation_min"`
	GoalBonded          string `json:"goal_bonded"`
	BlocksPerYear       string `json:"blocks_per_year"`
}

type MintGenesisState struct {
	Minter MintMinter `json:"minter"`
	Params MintParams `json:"params"`
}

const (
	mintInflation           = "0.130000000000000000"
	mintAnnualProvisions    = "0.000000000000000000"
	mintInflationRateChange = "0.130000000000000000"
	mintInflationMax        = "0.200000000000000000"
	mintInflationMin        = "0.070000000000000000"
	mintGoalBonded          = "0.670000000000000000"
	mintBlocksPerYear       = "6311520"
)

func GetDefaultMintGenesis(mintBaseDenom string) []byte {
	gs := MintGenesisState{
		Minter: MintMinter{Inflation: mintInflation, AnnualProvisions: mintAnnualProvisions},
		Params: MintParams{
			MintDenom:           mintBaseDenom,
			InflationRateChange: mintInflationRateChange,
			InflationMax:        mintInflationMax,
			InflationMin:        mintInflationMin,
			GoalBonded:          mintGoalBonded,
			BlocksPerYear:       mintBlocksPerYear,
		},
	}
	bz, _ := json.Marshal(&gs)
	return bz
}
