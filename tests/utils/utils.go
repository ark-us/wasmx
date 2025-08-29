package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"cosmossdk.io/math"
	ibcgotesting "github.com/cosmos/ibc-go/v8/testing"

	cosmosmodtypes "github.com/loredanacirstea/wasmx/x/cosmosmod/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func GenesisModify(genesisState map[string]json.RawMessage, app ibcgotesting.TestingApp) map[string]json.RawMessage {

	// make it easier to test jailing validators in TestStakingJailValidator
	var cosmosmodGenState cosmosmodtypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesisState[cosmosmodtypes.ModuleName], &cosmosmodGenState)
	p, _ := math.LegacyNewDecFromStr("0.6")
	cosmosmodGenState.Slashing.Params.MinSignedPerWindow = p
	cosmosmodGenState.Slashing.Params.SignedBlocksWindow = 4
	cosmosmodGenState.Slashing.Params.DowntimeJailDuration = time.Second * 1

	genesisState[cosmosmodtypes.ModuleName] = app.AppCodec().MustMarshalJSON(&cosmosmodGenState)

	return genesisState
}

func SystemContractsModify(wasmRuntime string) func([]wasmxtypes.SystemContract) []wasmxtypes.SystemContract {
	return func(contracts []wasmxtypes.SystemContract) []wasmxtypes.SystemContract {
		var compiledMap map[string]bool = nil
		if wasmRuntime == "wasmedge" {
			compiledMap = wasmedgeCompiled
		} else {
			compiledMap = wazeroCompiled
		}
		if compiledMap != nil {
			for i := range contracts {
				pinned, ok := compiledMap[contracts[i].Label]
				if ok && pinned {
					contracts[i].Pinned = true
				} else {
					contracts[i].Pinned = false
				}

			}
		}
		return contracts
	}
}

var wazeroCompiled map[string]bool = nil

var wasmedgeCompiled = map[string]bool{
	// wasmxtypes.AUTH_v001:                true,
	// wasmxtypes.ROLES_v001:               true,
	"ecrecovereth":                      true,
	"sha2-256":                          true,
	"ripmd160":                          true,
	"modexp":                            true,
	"ecadd":                             true,
	"ecmul":                             true,
	"ecpairings":                        true,
	"blake2f":                           true,
	wasmxtypes.INTERPRETER_EVM_SHANGHAI: true,
	// // wasmxtypes.INTERPRETER_PYTHON: true,
	// // wasmxtypes.INTERPRETER_JS: true,
	// // wasmxtypes.INTERPRETER_FSM: true,
	// // wasmxtypes.INTERPRETER_TAY: true,
	// "secp384r1":                               true,
	// "secp384r1_registry":                      true,
	// wasmxtypes.STAKING_v001:                   true,
	// wasmxtypes.BANK_v001:                      true,
	// wasmxtypes.ERC20_v001:                     true,
	// wasmxtypes.DERC20_v001:                    true,
	// wasmxtypes.SLASHING_v001:                  true,
	// wasmxtypes.DISTRIBUTION_v001:              true,
	// wasmxtypes.GOV_v001:                       true,
	// wasmxtypes.GOV_CONT_v001:                  true,
	// "raft_library":                            true,
	// "raftp2p_library":                         true,
	// "tendermint_library":                      true,
	// "tendermintp2p_library":                   true,
	// "ava_snowman_library":                     true,
	// wasmxtypes.TIME_v001:                      true,
	// "level0_library":                          true,
	// wasmxtypes.MULTICHAIN_REGISTRY_LOCAL_v001: true,
	// "lobby_library":                           true,
	// wasmxtypes.METAREGISTRY_v001:              true,
	// "level0_ondemand_library":                 true,
	// wasmxtypes.MULTICHAIN_REGISTRY_v001:       true,
	// wasmxtypes.CHAT_v001:                      true,
	// wasmxtypes.CHAT_VERIFIER_v001:             true,
	// wasmxtypes.HOOKS_v001:                     true,
	// wasmxtypes.HOOKS_v001:                     true,
}

func GetCompiledCacheDir(rootdir string, wasmRuntime string) string {
	path := path.Join(rootdir, "../", "codes_compiled", wasmRuntime, runtime.GOOS+"-"+runtime.GOARCH)
	if dirExists(path) {
		fmt.Println("tests use cached precompiles:", path)
		return path
	}
	return ""
}

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}
