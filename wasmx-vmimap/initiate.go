package vmimap

import (
	"fmt"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InstantiateImapVM_i32(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxImapVM_i32(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateImapVMMock_i32(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	context.Ctx.Logger().Info(fmt.Sprintf("instantiate IMAP mock i32 APIs: %s", context.ContractInfo.Address.String()))
	wasmx, err := BuildWasmxImapVMMock_i32(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateImapVM_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxImapVM_i64(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateImapVMMock_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	context.Ctx.Logger().Info(fmt.Sprintf("instantiate IMAP mock i64 APIs: %s", context.ContractInfo.Address.String()))
	wasmx, err := BuildWasmxImapVMMock_i64(context, rnh)
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
	// types.AdditionalEntryPointMap[ENTRY_POINT_IMAP] = true
	// types.AdditionalEntryPointMap[ENTRY_POINT_IMAP_SERVER] = true

	vmtypes.DependenciesMap[HOST_WASMX_ENV_IMAP_EXPORT] = true

	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_IMAP_i32_VER1, InstantiateImapVM_i32)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_IMAP_i32_VER1] = true

	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_IMAP_i64_VER1, InstantiateImapVM_i64)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_IMAP_i64_VER1] = true

	// mocked APIs
	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_IMAP_i32_VER1, InstantiateImapVMMock_i32)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_IMAP_i32_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_IMAP_i64_VER1, InstantiateImapVMMock_i64)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_IMAP_i64_VER1] = true

	types.SetEntryPoint(ENTRY_POINT_IMAP)
}
