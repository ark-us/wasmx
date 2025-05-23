package vmoauth2client

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InstantiateOAuth2Vm(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxOAuth2Client(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateOAuth2VmMock(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxOAuth2ClientMock(context, rnh)
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
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_OAUTH2CLIENT_VER1, InstantiateOAuth2Vm)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_OAUTH2CLIENT_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_OAUTH2CLIENT_VER1, InstantiateOAuth2VmMock)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_OAUTH2CLIENT_VER1] = true
}
