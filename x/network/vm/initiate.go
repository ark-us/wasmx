package vm

import (
	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	vmtypes "mythos/v1/x/wasmx/vm"
)

func InstantiateWasmxConsensusJson(context *vmtypes.Context, contractVm *wasmedge.VM, dep *types.SystemDep) ([]func(), error) {
	var cleanups []func()
	var err error
	wasmx := BuildWasmxP2P1(context)
	err = contractVm.RegisterModule(wasmx)
	if err != nil {
		return cleanups, err
	}
	cleanups = append(cleanups, wasmx.Release)
	return cleanups, nil
}

func init() {
	vmtypes.DependenciesMap[HOST_WASMX_ENV_EXPORT] = true
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_P2P_VER1, InstantiateWasmxConsensusJson)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_P2P_VER1] = true
}
