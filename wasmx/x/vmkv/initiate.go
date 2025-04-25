package vmkv

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InstantiateKvVM(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxKvVM(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateKvVMMock(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxKvVMMock(context, rnh)
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
	vmtypes.DependenciesMap[HOST_WASMX_ENV_KVDB_EXPORT] = true
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_KVDB_VER1, InstantiateKvVM)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_KVDB_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_KVDB_VER1, InstantiateKvVMMock)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_KVDB_VER1] = true
}
