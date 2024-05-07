package vmmc

import (
	vmtypes "mythos/v1/x/wasmx/vm"
)

const HOST_WASMX_ENV_MULTICHAIN_VER1 = "wasmx_multichain_1"

var HOST_WASMX_ENV_EXPORT = "wasmx_multichain_"

var HOST_WASMX_ENV_MULTICHAIN = "multichain"

type Context struct {
	*vmtypes.Context
}
