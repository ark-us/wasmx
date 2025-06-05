package vmoauth2client

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InstantiateOAuth2Vm_i32(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxOAuth2Client_i32(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateOAuth2VmMock_i32(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxOAuth2ClientMock_i32(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateOAuth2Vm_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxOAuth2Client_i64(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateOAuth2VmMock_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxOAuth2ClientMock_i64(context, rnh)
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
	vmtypes.DependenciesMap[HOST_WASMX_ENV_OAUTH2CLIENT_EXPORT] = true
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_OAUTH2CLIENT_i32_VER1, InstantiateOAuth2Vm_i32)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_OAUTH2CLIENT_i32_VER1] = true

	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_OAUTH2CLIENT_i64_VER1, InstantiateOAuth2Vm_i64)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_OAUTH2CLIENT_i64_VER1] = true

	// mock apis

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_OAUTH2CLIENT_i32_VER1, InstantiateOAuth2VmMock_i32)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_OAUTH2CLIENT_i32_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_OAUTH2CLIENT_i64_VER1, InstantiateOAuth2VmMock_i64)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_OAUTH2CLIENT_i64_VER1] = true
}
