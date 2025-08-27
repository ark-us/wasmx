package vmmc

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
)

func Setup() {
	vmtypes.DependenciesMap[HOST_WASMX_ENV_EXPORT] = true
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_MULTICHAIN_VER1, InstantiateWasmxMultiChainJson_i32)
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_MULTICHAIN_VER1_i32, InstantiateWasmxMultiChainJson_i32)
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_MULTICHAIN_VER1_i64, InstantiateWasmxMultiChainJson_i64)

	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_MULTICHAIN_VER1] = true
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_MULTICHAIN_VER1_i32] = true
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_MULTICHAIN_VER1_i64] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_MULTICHAIN_VER1, InstantiateWasmxMultiChainJsonMock_i32)
	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_MULTICHAIN_VER1_i32, InstantiateWasmxMultiChainJsonMock_i32)
	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_MULTICHAIN_VER1_i64, InstantiateWasmxMultiChainJsonMock_i64)

	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_MULTICHAIN_VER1] = true
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_MULTICHAIN_VER1_i32] = true
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_MULTICHAIN_VER1_i64] = true
}
