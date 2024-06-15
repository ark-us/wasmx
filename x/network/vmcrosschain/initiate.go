package vmcrosschain

import (
	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	vmtypes "mythos/v1/x/wasmx/vm"
)

// !!!!This is an internal API only to be used by trusted system contracts
func InstantiateWasmxCrossChainJson(context *vmtypes.Context, contractVm *wasmedge.VM, dep *types.SystemDep) ([]func(), error) {
	var cleanups []func()
	var err error
	wasmx := BuildWasmxCrosschainJson1(context)
	err = contractVm.RegisterModule(wasmx)
	if err != nil {
		return cleanups, err
	}
	cleanups = append(cleanups, wasmx.Release)
	return cleanups, nil
}

func Setup() {
	vmtypes.DependenciesMap[HOST_WASMX_ENV_CROSSCHAIN_EXPORT] = true
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_CROSSCHAIN_VER1, InstantiateWasmxCrossChainJson)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_CROSSCHAIN_VER1] = true
}
