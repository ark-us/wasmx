package vm

import (
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func BuildWasiEnv(_context *Context, rnh memc.RuntimeHandler, enableNonDeterminism bool) (interface{}, error) {
	wasi_randomGet_fn := wasi_randomGet
	wasi_clockTimeGet_fn := wasi_clockTimeGet
	if enableNonDeterminism {
		wasi_randomGet_fn = wasi_randomGet_NonD
		wasi_clockTimeGet_fn = wasi_clockTimeGet_NonD
	}

	context := &WasiContext{
		c:           _context,
		openFiles:   map[int32]*openFile{},
		fileMapping: map[string][]byte{},
		dirMapping:  map[string]bool{},
	}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		// 1) args_get(argv: Pointer<Pointer<u8>>, argv_buf: Pointer<u8>) -> errno
		vm.BuildFn(
			"args_get",
			wasi_argsGet,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32()}, // argv, argv_buf
			[]interface{}{vm.ValType_I32()},                   // errno
			0,
		),

		// 2) args_sizes_get() -> (errno, size, size)
		vm.BuildFn(
			"args_sizes_get",
			wasi_argsSizesGet,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32()}, // no inputs
			[]interface{}{vm.ValType_I32()},                   // errno, size, size
			0,
		),

		// 3) environ_get(environ: Pointer<Pointer<u8>>, environ_buf: Pointer<u8>) -> errno
		vm.BuildFn(
			"environ_get",
			wasi_environGet,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32()}, // environ, environ_buf
			[]interface{}{vm.ValType_I32()},                   // errno
			0,
		),

		// 4) environ_sizes_get() -> (errno, size, size)
		vm.BuildFn(
			"environ_sizes_get",
			wasi_environSizesGet,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 5) clock_res_get(id: clockid) -> (errno, timestamp)
		vm.BuildFn(
			"clock_res_get",
			wasi_clockResGet,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32()}, // clockid
			[]interface{}{vm.ValType_I32()},                   // errno, timestamp
			0,
		),

		// 6) clock_time_get(id: clockid, precision: timestamp) -> (errno, timestamp)
		vm.BuildFn(
			"clock_time_get",
			wasi_clockTimeGet_fn,
			[]interface{}{vm.ValType_I32(), vm.ValType_I64(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 7) fd_advise(fd: fd, offset: filesize, len: filesize, advice: advice) -> errno
		vm.BuildFn(
			"fd_advise",
			wasi_fdAdvise,
			[]interface{}{vm.ValType_I32(), vm.ValType_I64(), vm.ValType_I64(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 8) fd_allocate(fd: fd, offset: filesize, len: filesize) -> errno
		vm.BuildFn(
			"fd_allocate",
			wasi_fdAllocate,
			[]interface{}{vm.ValType_I32(), vm.ValType_I64(), vm.ValType_I64()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 9) fd_close(fd: fd) -> errno
		vm.BuildFn(
			"fd_close",
			wasi_fdClose,
			[]interface{}{vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 10) fd_datasync(fd: fd) -> errno
		vm.BuildFn(
			"fd_datasync",
			wasi_fdDatasync,
			[]interface{}{vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 11) fd_fdstat_get(fd: fd) -> (errno, fdstat)
		// fdstat is typically written back to memory rather than returned.
		// But if you’re returning it directly, you need to decide how to represent fdstat as ValTypes.
		vm.BuildFn(
			"fd_fdstat_get",
			wasi_fdFdstatGet,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()}, // For example: (errno, fdstatPlaceholder)
			0,
		),

		// 12) fd_fdstat_set_flags(fd: fd, flags: fdflags) -> errno
		vm.BuildFn(
			"fd_fdstat_set_flags",
			wasi_fdFdstatSetFlags,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 13) fd_fdstat_set_rights(fd: fd, fs_rights_base: rights, fs_rights_inheriting: rights) -> errno
		vm.BuildFn(
			"fd_fdstat_set_rights",
			wasi_fdFdstatSetRights,
			[]interface{}{vm.ValType_I32(), vm.ValType_I64(), vm.ValType_I64()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 14) fd_filestat_get(fd: fd) -> (errno, filestat)
		// Same note as fdstat: filestat often is written to memory, but here we show a direct return placeholder.
		vm.BuildFn(
			"fd_filestat_get",
			wasi_fdFilestatGet,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()}, // placeholder
			0,
		),

		// 15) fd_filestat_set_size(fd: fd, size: filesize) -> errno
		vm.BuildFn(
			"fd_filestat_set_size",
			wasi_fdFilestatSetSize,
			[]interface{}{vm.ValType_I32(), vm.ValType_I64()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 16) fd_filestat_set_times(fd: fd, atim: timestamp, mtim: timestamp, fst_flags: fstflags) -> errno
		vm.BuildFn(
			"fd_filestat_set_times",
			wasi_fdFilestatSetTimes,
			[]interface{}{vm.ValType_I32(), vm.ValType_I64(), vm.ValType_I64(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 17) fd_pread(fd: fd, iovs: iovec_array, offset: filesize) -> (errno, size)
		vm.BuildFn(
			"fd_pread",
			wasi_fdPread,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I64(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 18) fd_prestat_get(fd: fd) -> (errno, prestat)
		// Typically returns a structure in memory, but you can placeholder it similarly as above.
		vm.BuildFn(
			"fd_prestat_get",
			wasi_fdPrestatGet,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()}, // placeholder
			0,
		),

		// 19) fd_prestat_dir_name(fd: fd, path: Pointer<u8>, path_len: size) -> errno
		vm.BuildFn(
			"fd_prestat_dir_name",
			wasi_fdPrestatDirName,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 20) fd_pwrite(fd: fd, iovs: ciovec_array, offset: filesize) -> (errno, size)
		vm.BuildFn(
			"fd_pwrite",
			wasi_fdPwrite,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I64(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 21) fd_read(fd: fd, iovs: iovec_array) -> (errno, size)
		vm.BuildFn(
			"fd_read",
			wasi_fdRead,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 22) fd_readdir(fd: fd, buf: Pointer<u8>, buf_len: size, cookie: dircookie) -> (errno, size)
		vm.BuildFn(
			"fd_readdir",
			wasi_fdReaddir,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I64(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 23) fd_renumber(fd: fd, to: fd) -> errno
		vm.BuildFn(
			"fd_renumber",
			wasi_fdRenumber,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 24) fd_seek(fd: fd, offset: filedelta, whence: whence) -> (errno, filesize)
		vm.BuildFn(
			"fd_seek",
			wasi_fdSeek,
			[]interface{}{vm.ValType_I32(), vm.ValType_I64(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 25) fd_sync(fd: fd) -> errno
		vm.BuildFn(
			"fd_sync",
			wasi_fdSync,
			[]interface{}{vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 26) fd_tell(fd: fd) -> (errno, filesize)
		vm.BuildFn(
			"fd_tell",
			wasi_fdTell,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 27) fd_write(fd: fd, iovs: ciovec_array) -> (errno, size)
		vm.BuildFn(
			"fd_write",
			wasi_fdWrite,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 28) path_create_directory(fd: fd, path: string) -> errno
		// Typically "path" is a pointer+length, but we simplify to one pointer here.
		vm.BuildFn(
			"path_create_directory",
			wasi_pathCreateDirectory,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 29) path_filestat_get(fd: fd, flags: lookupflags, path: string) -> (errno, filestat)
		vm.BuildFn(
			"path_filestat_get",
			wasi_pathFilestatGet,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()}, // placeholder
			0,
		),

		// 30) path_filestat_set_times(fd: fd, flags: lookupflags, path: string,
		//     atim: timestamp, mtim: timestamp, fst_flags: fstflags) -> errno
		vm.BuildFn(
			"path_filestat_set_times",
			wasi_pathFilestatSetTimes,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I64(), vm.ValType_I64(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 31) path_link(old_fd: fd, old_flags: lookupflags, old_path: string,
		//     new_fd: fd, new_path: string) -> errno
		vm.BuildFn(
			"path_link",
			wasi_pathLink,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 32) path_open(fd: fd, dirflags: lookupflags, path: string,
		//     oflags: oflags, fs_rights_base: rights, fs_rights_inheriting: rights,
		//     fdflags: fdflags) -> (errno, fd)
		vm.BuildFn(
			"path_open",
			wasi_pathOpen,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I64(), vm.ValType_I64(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()}, // errno, fd
			0,
		),

		// 33) path_readlink(fd: fd, path: string, buf: Pointer<u8>, buf_len: size) -> (errno, size)
		vm.BuildFn(
			"path_readlink",
			wasi_pathReadlink,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 34) path_remove_directory(fd: fd, path: string) -> errno
		vm.BuildFn(
			"path_remove_directory",
			wasi_pathRemoveDirectory,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 35) path_rename(fd: fd, old_path: string, new_fd: fd, new_path: string) -> errno
		vm.BuildFn(
			"path_rename",
			wasi_pathRename,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 36) path_symlink(old_path: string, fd: fd, new_path: string) -> errno
		vm.BuildFn(
			"path_symlink",
			wasi_pathSymlink,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 37) path_unlink_file(fd: fd, path: string) -> errno
		vm.BuildFn(
			"path_unlink_file",
			wasi_pathUnlinkFile,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 38) poll_oneoff(in: ConstPointer<subscription>, out: Pointer<event>, nsubscriptions: size) -> (errno, size)
		vm.BuildFn(
			"poll_oneoff",
			wasi_pollOneoff,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 39) proc_exit(rval: exitcode)
		// Typically does not return; might raise a trap or shut down the instance.
		// But as a placeholder, you can define a return type of none (empty).
		vm.BuildFn(
			"proc_exit",
			wasi_procExit,
			[]interface{}{vm.ValType_I32()},
			[]interface{}{},
			0,
		),

		// 40) proc_raise(sig: signal) -> errno
		vm.BuildFn(
			"proc_raise",
			wasi_procRaise,
			[]interface{}{vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 41) sched_yield() -> errno
		vm.BuildFn(
			"sched_yield",
			wasi_schedYield,
			[]interface{}{},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 42) random_get(buf: Pointer<u8>, buf_len: size) -> errno
		vm.BuildFn(
			"random_get",
			wasi_randomGet_fn,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		vm.BuildFn(
			"sock_accept",
			wasi_sockAccept,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 43) sock_recv(fd: fd, ri_data: iovec_array, ri_flags: riflags) -> (errno, size, roflags)
		vm.BuildFn(
			"sock_recv",
			wasi_sockRecv,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 44) sock_send(fd: fd, si_data: ciovec_array, si_flags: siflags) -> (errno, size)
		vm.BuildFn(
			"sock_send",
			wasi_sockSend,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),

		// 45) sock_shutdown(fd: fd, how: sdflags) -> errno
		vm.BuildFn(
			"sock_shutdown",
			wasi_sockShutdown,
			[]interface{}{vm.ValType_I32(), vm.ValType_I32()},
			[]interface{}{vm.ValType_I32()},
			0,
		),
	}
	// wasi_unstable
	return vm.BuildModule(rnh, "wasi_snapshot_preview1", context, fndefs)
}
