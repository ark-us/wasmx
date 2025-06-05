package vmhttpserver

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InstantiateHttpServerVm_i32(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxHttpServer_i32(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateHttpVmServerMock_i32(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxHttpServerMock_i32(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateHttpServerVm_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxHttpServer_i64(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateHttpVmServerMock_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxHttpServerMock_i64(context, rnh)
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
	vmtypes.DependenciesMap[HOST_WASMX_ENV_HTTP_EXPORT] = true
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_HTTP_i32_VER1, InstantiateHttpServerVm_i32)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_HTTP_i32_VER1] = true

	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_HTTP_i64_VER1, InstantiateHttpServerVm_i64)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_HTTP_i64_VER1] = true

	// mock apis

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_HTTP_i32_VER1, InstantiateHttpVmServerMock_i32)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_HTTP_i32_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_HTTP_i64_VER1, InstantiateHttpVmServerMock_i64)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_HTTP_i64_VER1] = true
}
