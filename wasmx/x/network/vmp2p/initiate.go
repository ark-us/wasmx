package vmp2p

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
)

func Setup() {
	vmtypes.DependenciesMap[HOST_WASMX_ENV_P2P_EXPORT] = true
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_P2P_VER1, InstantiateWasmxP2PJson_i32)
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_P2P_VER1_i32, InstantiateWasmxP2PJson_i32)
	vmtypes.SetSystemDepHandler(HOST_WASMX_ENV_P2P_VER1_i64, InstantiateWasmxP2PJson_i64)
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_P2P_VER1] = true
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_P2P_VER1_i32] = true
	types.SUPPORTED_HOST_INTERFACES[HOST_WASMX_ENV_P2P_VER1_i64] = true

	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_P2P_VER1, InstantiateWasmxP2PJsonMock_i32)
	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_P2P_VER1_i32, InstantiateWasmxP2PJsonMock_i32)
	vmtypes.SetSystemDepHandlerMock(HOST_WASMX_ENV_P2P_VER1_i64, InstantiateWasmxP2PJsonMock_i64)
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_P2P_VER1] = true
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_P2P_VER1_i32] = true
	types.PROTECTED_HOST_APIS[HOST_WASMX_ENV_P2P_VER1_i64] = true
}
