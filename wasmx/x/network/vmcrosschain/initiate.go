package vmcrosschain

import (
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
)

// !!!!This is an internal API only to be used by trusted system contracts
func InstantiateWasmxCrossChainJson(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	var err error
	wasmx, err := BuildWasmxCrosschainJson1(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateWasmxCrossChainJsonMock(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	var err error
	wasmx, err := BuildWasmxCrosschainJson1Mock(context, rnh)
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
	vmtypes.DependenciesMap[HOST_WASMX_ENV_CROSSCHAIN_EXPORT] = true
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_CROSSCHAIN_VER1, InstantiateWasmxCrossChainJson)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_CROSSCHAIN_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_CROSSCHAIN_VER1, InstantiateWasmxCrossChainJsonMock)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_CROSSCHAIN_VER1] = true
}
