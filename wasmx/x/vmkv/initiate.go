package vmkv

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InstantiateKvVM_i32(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxKvVM_i32(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateKvVMMock_i32(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxKvVMMock_i32(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateKvVM_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxKvVM_i64(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateKvVMMock_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxKvVMMock_i64(context, rnh)
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
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_KVDB_i32_VER1, InstantiateKvVM_i32)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_KVDB_i32_VER1] = true

	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_KVDB_i64_VER1, InstantiateKvVM_i64)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_KVDB_i64_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_KVDB_i32_VER1, InstantiateKvVMMock_i32)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_KVDB_i32_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_KVDB_i64_VER1, InstantiateKvVMMock_i64)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_KVDB_i64_VER1] = true
}
