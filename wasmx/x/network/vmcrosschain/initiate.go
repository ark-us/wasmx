package vmcrosschain

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
)

func Setup() {
	vmtypes.DependenciesMap[HOST_WASMX_ENV_CROSSCHAIN_EXPORT] = true
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_CROSSCHAIN_VER1, InstantiateWasmxCrossChainJson_i32)
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_CROSSCHAIN_VER1_i32, InstantiateWasmxCrossChainJson_i32)
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_CROSSCHAIN_VER1_i64, InstantiateWasmxCrossChainJson_i64)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_CROSSCHAIN_VER1] = true
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_CROSSCHAIN_VER1_i32] = true
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_CROSSCHAIN_VER1_i64] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_CROSSCHAIN_VER1, InstantiateWasmxCrossChainJsonMock_i32)
	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_CROSSCHAIN_VER1_i32, InstantiateWasmxCrossChainJsonMock_i32)
	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_CROSSCHAIN_VER1_i64, InstantiateWasmxCrossChainJsonMock_i64)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_CROSSCHAIN_VER1] = true
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_CROSSCHAIN_VER1_i32] = true
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_CROSSCHAIN_VER1_i64] = true
}
