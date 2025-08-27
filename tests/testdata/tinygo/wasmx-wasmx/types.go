package wasmxwasmx

import (
	wasmx "github.com/loredanacirstea/wasmx-env"
)

// Params mirrors the empty AssemblyScript Params
type Params struct{}

type Code struct {
	CodeID    uint64         `json:"code_id"`
	CodeInfo  wasmx.CodeInfo `json:"code_info"`
	CodeBytes []byte         `json:"code_bytes"`
}

type Contract struct {
	ContractAddress wasmx.Bech32String      `json:"contract_address"`
	ContractInfo    wasmx.ContractInfo      `json:"contract_info"`
	ContractState   []wasmx.ContractStorage `json:"contract_state"`
}

type Sequence struct {
	IdKey []byte `json:"id_key"`
	Value uint64 `json:"value"`
}

type GenesisState struct {
	Params                  Params                 `json:"params"`
	BootstrapAccountAddress wasmx.Bech32String     `json:"bootstrap_account_address"`
	SystemContracts         []wasmx.SystemContract `json:"system_contracts"`
	Codes                   []Code                 `json:"codes"`
	Contracts               []Contract             `json:"contracts"`
	Sequences               []Sequence             `json:"sequences"`
	CompiledFolderPath      string                 `json:"compiled_folder_path"`
}
