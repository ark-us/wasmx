package vmp2p

import (
	"mythos/v1/x/wasmx/types"
	vmtypes "mythos/v1/x/wasmx/vm"
	memc "mythos/v1/x/wasmx/vm/memory/common"
)

func InstantiateWasmxP2PJson(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxP2P1(context, rnh)
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
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_P2P_VER1, InstantiateWasmxP2PJson)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_P2P_VER1] = true
}
