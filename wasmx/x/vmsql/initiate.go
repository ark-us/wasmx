package vmsql

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InstantiateSqlVM_i32(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxSqlVM_i32(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateSqlVMMock_i32(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxSqlVMMock_i32(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateSqlVM_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxSqlVM_i64(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateSqlVMMock_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	wasmx, err := BuildWasmxSqlVMMock_i64(context, rnh)
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
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_SQL_i32_VER1, InstantiateSqlVM_i32)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_SQL_i32_VER1] = true

	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_SQL_i64_VER1, InstantiateSqlVM_i64)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_SQL_i64_VER1] = true

	// mock apis

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_SQL_i32_VER1, InstantiateSqlVMMock_i32)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_SQL_i32_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_SQL_i64_VER1, InstantiateSqlVMMock_i64)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_SQL_i64_VER1] = true
}
