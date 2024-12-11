package vmmc

import (
	"wasmx/v1/x/wasmx/types"
	vmtypes "wasmx/v1/x/wasmx/vm"
	memc "wasmx/v1/x/wasmx/vm/memory/common"
)

func InstantiateWasmxMultiChainJson(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxMultichainJson1(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func Setup() {
	vmtypes.DependenciesMap[HOST_WASMX_ENV_EXPORT] = true
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_MULTICHAIN_VER1, InstantiateWasmxMultiChainJson)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_MULTICHAIN_VER1] = true
}
