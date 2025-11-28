package vmoauth2server

import (
	"context"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InitiateOAuth2ServerContext(ctx context.Context, vmContext *vmtypes.Context) (context.Context, error) {
	oauth2Ctx := &OAuth2ServerContext{
		Instances: make(map[string]*OAuth2ServerInstance),
	}
	return context.WithValue(ctx, OAuth2ServerContextKey, oauth2Ctx), nil
}

func InstantiateOAuth2ServerVm_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep, allDeps []types.SystemDep, hasCoreRole bool) error {
	wasmx, err := BuildWasmxOAuth2Server_i64(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateOAuth2ServerVmMock_i64(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep, allDeps []types.SystemDep, hasCoreRole bool) error {
	wasmx, err := BuildWasmxOAuth2ServerMock_i64(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateOAuth2ServerVm_i32(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep, allDeps []types.SystemDep, hasCoreRole bool) error {
	wasmx, err := BuildWasmxOAuth2Server_i32(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func InstantiateOAuth2ServerVmMock_i32(context *vmtypes.Context, rnh memc.RuntimeHandler, dep *types.SystemDep, allDeps []types.SystemDep, hasCoreRole bool) error {
	wasmx, err := BuildWasmxOAuth2ServerMock_i32(context, rnh)
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
	vmtypes.DependenciesMap[HOST_WASMX_ENV_OAUTH2SERVER_EXPORT] = true

	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_OAUTH2SERVER_i32_VER1, InstantiateOAuth2ServerVm_i32)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_OAUTH2SERVER_i32_VER1] = true

	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_OAUTH2SERVER_i64_VER1, InstantiateOAuth2ServerVm_i64)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_OAUTH2SERVER_i64_VER1] = true

	// mock apis
	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_OAUTH2SERVER_i32_VER1, InstantiateOAuth2ServerVmMock_i32)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_OAUTH2SERVER_i32_VER1] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_OAUTH2SERVER_i64_VER1, InstantiateOAuth2ServerVmMock_i64)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_OAUTH2SERVER_i64_VER1] = true
}
