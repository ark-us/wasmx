package vmsmtp

import (
	"fmt"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InstantiateSmtpVM(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxSmtpVM(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateSmtpVMMock(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	context.Ctx.Logger().Info(fmt.Sprintf("instantiate SMTP mock APIs: %s", context.ContractInfo.Address.String()))
	wasmx, err := BuildWasmxSmtpVMMock(context, rnh)
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
	vmtypes.DependenciesMap[HOST_WASMX_ENV_SMTP_EXPORT] = true
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_SMTP_VER1, InstantiateSmtpVM)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_SMTP_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_SMTP_VER1, InstantiateSmtpVMMock)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_SMTP_VER1] = true
}
