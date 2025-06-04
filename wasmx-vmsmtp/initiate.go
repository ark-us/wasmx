package vmsmtp

import (
	"fmt"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InstantiateSmtpVM_i32(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxSmtpVM_i32(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateSmtpVMMock_i32(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	context.Ctx.Logger().Info(fmt.Sprintf("instantiate SMTP mock i32 APIs: %s", context.ContractInfo.Address.String()))
	wasmx, err := BuildWasmxSmtpVMMock_i32(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateSmtpVM_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxSmtpVM_i64(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateSmtpVMMock_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	context.Ctx.Logger().Info(fmt.Sprintf("instantiate SMTP mock i64 APIs: %s", context.ContractInfo.Address.String()))
	wasmx, err := BuildWasmxSmtpVMMock_i64(context, rnh)
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
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_SMTP_i32_VER1, InstantiateSmtpVM_i32)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_SMTP_i32_VER1] = true

	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_SMTP_i64_VER1, InstantiateSmtpVM_i64)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_SMTP_i64_VER1] = true

	// mocked apis

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_SMTP_i32_VER1, InstantiateSmtpVMMock_i32)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_SMTP_i32_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_SMTP_i64_VER1, InstantiateSmtpVMMock_i64)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_SMTP_i64_VER1] = true
}
