package vmp2p

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
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

func InstantiateWasmxP2PJsonMock(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxP2P1Mock(context, rnh)
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
	vmtypes.DependenciesMap[HOST_WASMX_ENV_P2P_EXPORT] = true
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_P2P_VER1, InstantiateWasmxP2PJson)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_P2P_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_P2P_VER1, InstantiateWasmxP2PJsonMock)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_P2P_VER1] = true
}
