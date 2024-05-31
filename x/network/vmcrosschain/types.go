package vmcrosschain

import (
	vmtypes "mythos/v1/x/wasmx/vm"
)

const HOST_WASMX_ENV_CROSSCHAIN_VER1 = "wasmx_crosschain_1"

var HOST_WASMX_ENV_CROSSCHAIN_EXPORT = "wasmx_crosschain_"

var HOST_WASMX_ENV_CROSSCHAIN = "crosschain"

type Context struct {
	*vmtypes.Context
}
