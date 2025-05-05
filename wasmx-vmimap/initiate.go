package vmimap

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InstantiateImapVM(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxImapVM(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateImapVMMock(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	context.Ctx.Logger().Info("* instantiate IMAP mock APIs: %s", context.ContractInfo.Address.String())
	wasmx, err := BuildWasmxImapVMMock(context, rnh)
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
	vmtypes.DependenciesMap[HOST_WASMX_ENV_IMAP_EXPORT] = true
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_IMAP_VER1, InstantiateImapVM)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_IMAP_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_IMAP_VER1, InstantiateImapVMMock)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_IMAP_VER1] = true
}
