package vmhttpclient

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InstantiateHttpClientVm(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxHttpClient(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateHttpClientVmMock(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxHttpClientMock(context, rnh)
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
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_HTTP_VER1, InstantiateHttpClientVm)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_HTTP_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_HTTP_VER1, InstantiateHttpClientVmMock)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_HTTP_VER1] = true
}
