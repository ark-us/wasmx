package vmsql

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InstantiateSqlVM(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxSqlVM(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateSqlVMMock(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxSqlVMMock(context, rnh)
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
	vmtypes.DependenciesMap[HOST_WASMX_ENV_SQL_EXPORT] = true
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_SQL_VER1, InstantiateSqlVM)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_SQL_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_SQL_VER1, InstantiateSqlVMMock)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_SQL_VER1] = true
}
