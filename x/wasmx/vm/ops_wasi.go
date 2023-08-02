package vm

import (
	"mythos/v1/x/wasmx/types"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

func wasiFdWrite(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	returns[0] = 0
	return returns, wasmedge.Result_Success
}

func wasiFdClose(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	returns[0] = 0
	return returns, wasmedge.Result_Success
}

func wasiFdFdstatGet(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	returns[0] = 0
	return returns, wasmedge.Result_Success
}

func wasiFdSeek(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	returns[0] = 0
	return returns, wasmedge.Result_Success
}

func wasiProcExit(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	returns[0] = 0
	return returns, wasmedge.Result_Success
	// process.exit(code);
	// or
	// throw 'trying to exit with code ' + code;
}

// bufPtr, bufLen
func wasiRandomGet(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	returns[0] = 0
	return returns, wasmedge.Result_Success
	// crypto.getRandomValues(loadSlice(bufPtr, bufLen));
	// return 0;
}

func wasiClockTimeGet(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	returns[0] = 0
	return returns, wasmedge.Result_Success
}

func wasiDefaultFn(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	returns[0] = 0
	return returns, wasmedge.Result_Success
}

func BuildWasiSnapshotPreview1(context *Context) *wasmedge.Module {
	env := wasmedge.NewModule("wasi_snapshot_preview1")
	functype__i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i64_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I64},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i64i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I64, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i64i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I64, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i64i64i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I64, wasmedge.ValType_I64, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32i64i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I64, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32i32i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32i32i32i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32i32i32i64i64i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I64, wasmedge.ValType_I64, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32i32i64i64i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I64, wasmedge.ValType_I64, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32i32i32i32i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)

	env.AddFunction("proc_exit", wasmedge.NewFunction(functype_i32_, wasiProcExit, context, 0))
	env.AddFunction("random_get", wasmedge.NewFunction(functype_i32i32_i32, wasiRandomGet, context, 0))
	env.AddFunction("clock_time_get", wasmedge.NewFunction(functype_i32i64i32_i32, wasiClockTimeGet, context, 0))
	env.AddFunction("sched_yield", wasmedge.NewFunction(functype__i32, wasiDefaultFn, context, 0))
	env.AddFunction("fd_read", wasmedge.NewFunction(functype_i32i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("fd_filestat_set_size", wasmedge.NewFunction(functype_i32i64_i32, wasiDefaultFn, context, 0))
	env.AddFunction("poll_oneoff", wasmedge.NewFunction(functype_i32i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("args_sizes_get", wasmedge.NewFunction(functype_i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("args_get", wasmedge.NewFunction(functype_i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("fd_filestat_get", wasmedge.NewFunction(functype_i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("fd_readdir", wasmedge.NewFunction(functype_i32i32i32i64i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("fd_write", wasmedge.NewFunction(functype_i32i32i32i32_i32, wasiFdWrite, context, 0))
	env.AddFunction("fd_close", wasmedge.NewFunction(functype_i32_i32, wasiFdClose, context, 0))
	env.AddFunction("fd_fdstat_get", wasmedge.NewFunction(functype_i32i32_i32, wasiFdFdstatGet, context, 0))
	env.AddFunction("fd_seek", wasmedge.NewFunction(functype_i32i64i32i32_i32, wasiFdSeek, context, 0))
	env.AddFunction("fd_fdstat_set_flags", wasmedge.NewFunction(functype_i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("fd_prestat_get", wasmedge.NewFunction(functype_i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("fd_prestat_dir_name", wasmedge.NewFunction(functype_i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("fd_sync", wasmedge.NewFunction(functype_i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("path_create_directory", wasmedge.NewFunction(functype_i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("path_unlink_file", wasmedge.NewFunction(functype_i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("path_rename", wasmedge.NewFunction(functype_i32i32i32i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("path_remove_directory", wasmedge.NewFunction(functype_i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("path_filestat_get", wasmedge.NewFunction(functype_i32i32i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("path_readlink", wasmedge.NewFunction(functype_i32i32i32i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("path_link", wasmedge.NewFunction(functype_i32i32i32i32i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("path_open", wasmedge.NewFunction(functype_i32i32i32i32i32i64i64i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("path_symlink", wasmedge.NewFunction(functype_i32i32i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("environ_get", wasmedge.NewFunction(functype_i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("environ_sizes_get", wasmedge.NewFunction(functype_i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("path_filestat_set_times", wasmedge.NewFunction(functype_i32i32i32i32i64i64i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("fd_filestat_set_times", wasmedge.NewFunction(functype_i32i64i64i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("fd_datasync", wasmedge.NewFunction(functype_i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("fd_pread", wasmedge.NewFunction(functype_i32i32i32i64i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("sock_shutdown", wasmedge.NewFunction(functype_i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("sock_open", wasmedge.NewFunction(functype_i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("sock_send", wasmedge.NewFunction(functype_i32i32i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("sock_recv", wasmedge.NewFunction(functype_i32i32i32i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("sock_connect", wasmedge.NewFunction(functype_i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("sock_setsockopt", wasmedge.NewFunction(functype_i32i32i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("sock_bind", wasmedge.NewFunction(functype_i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("sock_listen", wasmedge.NewFunction(functype_i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("sock_accept", wasmedge.NewFunction(functype_i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("sock_getlocaladdr", wasmedge.NewFunction(functype_i32i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("sock_getpeeraddr", wasmedge.NewFunction(functype_i32i32i32i32_i32, wasiDefaultFn, context, 0))
	env.AddFunction("sock_getaddrinfo", wasmedge.NewFunction(functype_i32i32i32i32i32i32i32i32_i32, wasiDefaultFn, context, 0))

	return env
}

func BuildWasiUnstable(context *Context) *wasmedge.Module {
	env := wasmedge.NewModule("wasi_unstable")
	functype__i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i64i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I64, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)

	env.AddFunction("fd_write", wasmedge.NewFunction(functype_i32i32i32i32_i32, wasiFdWrite, context, 0))
	env.AddFunction("fd_close", wasmedge.NewFunction(functype__i32, wasiFdClose, context, 0))
	env.AddFunction("fd_fdstat_get", wasmedge.NewFunction(functype__i32, wasiFdFdstatGet, context, 0))
	env.AddFunction("fd_seek", wasmedge.NewFunction(functype__i32, wasiFdSeek, context, 0))
	env.AddFunction("proc_exit", wasmedge.NewFunction(functype_i32_, wasiProcExit, context, 0))
	env.AddFunction("random_get", wasmedge.NewFunction(functype_i32i32_i32, wasiRandomGet, context, 0))
	env.AddFunction("clock_time_get", wasmedge.NewFunction(functype_i32i64i32_i32, wasiClockTimeGet, context, 0))

	return env
}

func ExecuteWasi(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	if funcName != types.ENTRY_POINT_INSTANTIATE {
		funcName = "_start"
	}
	return nil, nil
}
