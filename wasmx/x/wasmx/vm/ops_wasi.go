package vm

import (
	"sort"
	"strings"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
	wasimem "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/wasi"
)

// wasi_snapshot_preview1
// See https://github.com/WebAssembly/WASI/blob/snapshot-01/phases/snapshot/docs.md#functions
// args_get(argv: Pointer<Pointer<u8>>, argv_buf: Pointer<u8>) -> errno
// args_sizes_get() -> (errno, size, size)
// environ_get(environ: Pointer<Pointer<u8>>, environ_buf: Pointer<u8>) -> errno
// environ_sizes_get() -> (errno, size, size)
// clock_res_get(id: clockid) -> (errno, timestamp)
// clock_time_get(id: clockid, precision: timestamp) -> (errno, timestamp)
// fd_advise(fd: fd, offset: filesize, len: filesize, advice: advice) -> errno
// fd_allocate(fd: fd, offset: filesize, len: filesize) -> errno
// fd_close(fd: fd) -> errno
// fd_datasync(fd: fd) -> errno
// fd_fdstat_get(fd: fd) -> (errno, fdstat)
// fd_fdstat_set_flags(fd: fd, flags: fdflags) -> errno
// fd_fdstat_set_rights(fd: fd, fs_rights_base: rights, fs_rights_inheriting: rights) -> errno
// fd_filestat_get(fd: fd) -> (errno, filestat)
// fd_filestat_set_size(fd: fd, size: filesize) -> errno
// fd_filestat_set_times(fd: fd, atim: timestamp, mtim: timestamp, fst_flags: fstflags) -> errno
// fd_pread(fd: fd, iovs: iovec_array, offset: filesize) -> (errno, size)
// fd_prestat_get(fd: fd) -> (errno, prestat)
// fd_prestat_dir_name(fd: fd, path: Pointer<u8>, path_len: size) -> errno
// fd_pwrite(fd: fd, iovs: ciovec_array, offset: filesize) -> (errno, size)
// fd_read(fd: fd, iovs: iovec_array) -> (errno, size)
// fd_readdir(fd: fd, buf: Pointer<u8>, buf_len: size, cookie: dircookie) -> (errno, size)
// fd_renumber(fd: fd, to: fd) -> errno
// fd_seek(fd: fd, offset: filedelta, whence: whence) -> (errno, filesize)
// fd_sync(fd: fd) -> errno
// fd_tell(fd: fd) -> (errno, filesize)
// fd_write(fd: fd, iovs: ciovec_array) -> (errno, size)
// path_create_directory(fd: fd, path: string) -> errno
// path_filestat_get(fd: fd, flags: lookupflags, path: string) -> (errno, filestat)
// path_filestat_set_times(fd: fd, flags: lookupflags, path: string, atim: timestamp, mtim: timestamp, fst_flags: fstflags) -> errno
// path_link(old_fd: fd, old_flags: lookupflags, old_path: string, new_fd: fd, new_path: string) -> errno
// path_open(fd: fd, dirflags: lookupflags, path: string, oflags: oflags, fs_rights_base: rights, fs_rights_inheriting: rights, fdflags: fdflags) -> (errno, fd)
// path_readlink(fd: fd, path: string, buf: Pointer<u8>, buf_len: size) -> (errno, size)
// path_remove_directory(fd: fd, path: string) -> errno
// path_rename(fd: fd, old_path: string, new_fd: fd, new_path: string) -> errno
// path_symlink(old_path: string, fd: fd, new_path: string) -> errno
// path_unlink_file(fd: fd, path: string) -> errno
// poll_oneoff(in: ConstPointer<subscription>, out: Pointer<event>, nsubscriptions: size) -> (errno, size)
// proc_exit(rval: exitcode)
// proc_raise(sig: signal) -> errno
// sched_yield() -> errno
// random_get(buf: Pointer<u8>, buf_len: size) -> errno
// sock_recv(fd: fd, ri_data: iovec_array, ri_flags: riflags) -> (errno, size, roflags)
// sock_send(fd: fd, si_data: ciovec_array, si_flags: siflags) -> (errno, size)
// sock_shutdown(fd: fd, how: sdflags) -> errno

var __WASI_O_CREAT = int32(1)
var __WASI_O_DIRECTORY = int32(2)

// var __WASI_O_EXCL = int32(4)
// var __WASI_O_TRUNC = int32(8)

func wasi_stubUnimplemented(_ interface{}, _ memc.RuntimeHandler, _ []interface{}) ([]interface{}, error) {
	// Return ENOSYS = 52
	// Function not implemented
	returns := make([]interface{}, 1)
	returns[0] = int32(52)
	return returns, nil
}

// 1) args_get(argv: Pointer<Pointer<u8>>, argv_buf: Pointer<u8>) -> errno
func wasi_argsGet(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_argsGet", "params", params)
	returns := make([]interface{}, 1)
	args := rnh.GetVm().WasiArgs()
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}

	argvPtr := params[0].(int32)
	argvBufPtr := params[1].(int32)

	if len(args) == 0 {
		// Write a single NULL pointer to mark end of the argv array
		if err := wasimem.WriteUint32Le(mem, argvPtr, 0); err != nil {
			return nil, err
		}
		returns[0] = int32(0) // errno=0 (success)
		return returns, nil
	}

	currentArgvPtr := argvPtr
	currentArgvBufPtr := argvBufPtr

	for _, a := range args {
		// write pointer (currentArgvBufPtr) into memory at currentArgvPtr
		err = wasimem.WriteUint32Le(mem, currentArgvPtr, uint32(currentArgvBufPtr))
		if err != nil {
			return nil, err
		}

		// write the actual string into memory + a null terminator
		data := append([]byte(a), []byte{0}...)
		err = mem.Write(currentArgvBufPtr, data)
		if err != nil {
			return nil, err
		}

		currentArgvPtr += 4 // 32-bit pointers
		currentArgvBufPtr += int32(len(a) + 1)
	}

	// write a NULL pointer after the last arg in the argv list
	err = wasimem.WriteUint32Le(mem, currentArgvPtr, uint32(0))
	if err != nil {
		return nil, err
	}
	returns[0] = int32(0)
	return returns, nil
}

// 2) args_sizes_get() -> (errno, size, size)
func wasi_argsSizesGet(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_argsSizesGet", "params", params)
	args := rnh.GetVm().WasiArgs()
	// We need total size of all args plus a null terminator per arg
	totalSize := 0
	for _, a := range args {
		totalSize += len(a) + 1
	}
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	LoggerExtended(ctx.c).Debug("wasi_argsGet", "count", len(args))
	err = wasimem.WriteUint32Le(mem, params[0].(int32), uint32(len(args)))
	if err != nil {
		return nil, err
	}
	err = wasimem.WriteUint32Le(mem, params[1].(int32), uint32(totalSize))
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = int32(0) // errno=0
	return returns, nil
}

// 3) environ_get(environ: Pointer<Pointer<u8>>, environ_buf: Pointer<u8>) -> errno
func wasi_environGet(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_environGet", "params", params)
	returns := make([]interface{}, 1)
	envs := rnh.GetVm().WasiEnvs()
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}

	argvPtr := params[0].(int32)
	argvBufPtr := params[1].(int32)

	if len(envs) == 0 {
		// Write a single NULL pointer to mark end of the argv array
		if err := wasimem.WriteUint32Le(mem, argvPtr, 0); err != nil {
			return nil, err
		}
		returns[0] = int32(0) // errno=0 (success)
		return returns, nil
	}

	currentArgvPtr := argvPtr
	currentArgvBufPtr := argvBufPtr

	for _, a := range envs {
		// write pointer (currentArgvBufPtr) into memory at currentArgvPtr
		err = wasimem.WriteUint32Le(mem, currentArgvPtr, uint32(currentArgvBufPtr))
		if err != nil {
			return nil, err
		}

		// write the actual string into memory + a null terminator
		err = mem.Write(currentArgvBufPtr, append([]byte(a), []byte{0}...))
		if err != nil {
			return nil, err
		}

		currentArgvPtr += 4 // 32-bit pointers
		currentArgvBufPtr += int32(len(a) + 1)
	}

	// write a NULL pointer after the last arg in the argv list
	err = wasimem.WriteUint32Le(mem, currentArgvPtr, uint32(0))
	if err != nil {
		return nil, err
	}
	returns[0] = int32(0)
	return returns, nil
}

// 4) environ_sizes_get() -> (errno, size, size)
func wasi_environSizesGet(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_environSizesGet", "params", params)
	envs := rnh.GetVm().WasiEnvs()
	// We need total size of all args plus a null terminator per arg
	totalSize := 0
	for _, a := range envs {
		totalSize += len(a) + 1
	}
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	err = wasimem.WriteUint32Le(mem, params[0].(int32), uint32(len(envs)))
	if err != nil {
		return nil, err
	}
	err = wasimem.WriteUint32Le(mem, params[1].(int32), uint32(totalSize))
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	return returns, nil
}

// 5) clock_res_get(id: clockid) -> (errno, timestamp)
func wasi_clockResGet(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_clockResGet", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 6) clock_time_get(id: clockid, precision: timestamp) -> (errno, timestamp)
func wasi_clockTimeGet(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_clockTimeGet", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 7) fd_advise(fd: fd, offset: filesize, len: filesize, advice: advice) -> errno
func wasi_fdAdvise(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdAdvise", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 8) fd_allocate(fd: fd, offset: filesize, len: filesize) -> errno
func wasi_fdAllocate(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdAllocate", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 9) fd_close(fd: fd) -> errno
func wasi_fdClose(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdClose", "params", params)
	fd := params[0].(int32)
	returns := make([]interface{}, 1)

	// Check if this FD exists in our openFiles
	if _, ok := ctx.openFiles[fd]; !ok {
		// EBADF => 8 (Bad file descriptor)
		returns[0] = int32(8)
		return returns, nil
	}

	// Remove it from our map of open files
	delete(ctx.openFiles, fd)

	// Return success (errno=0)
	returns[0] = int32(0)
	return returns, nil
}

// 10) fd_datasync(fd: fd) -> errno
func wasi_fdDatasync(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdDatasync", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 11) fd_fdstat_get(fd: fd) -> (errno, fdstat)
func wasi_fdFdstatGet(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdFdstatGet", "params", params)
	fd := params[0].(int32)
	fdstatPtr := params[1].(int32)

	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}

	openFiles := ctx.GetOpenFiles(rnh.GetVm())
	fileMap := ctx.GetFileMap(rnh.GetVm())
	openF, ok := openFiles[fd]
	returns := make([]interface{}, 1) // [errno]

	if !ok {
		// EBADF => 8
		returns[0] = int32(8)
		return returns, nil
	}

	// Let's assume everything is a "regular file" if found in FileMapping
	// If you had directories, you'd set a different filetype
	_, ok = fileMap[openF.path]
	if !ok {
		// Not found => ENOENT => 44 or EIO => 29
		returns[0] = int32(44)
		return returns, nil
	}

	// Prepare fields for __wasi_fdstat_t
	// We'll do a minimal approach:
	// fs_filetype = 4 => __WASI_FILETYPE_REGULAR_FILE
	// fs_flags = 0
	// pad = 0
	// fs_rights_base = 0 for minimal (or ~0 for "all rights")
	// fs_rights_inheriting = 0 // or 0xffffffff
	fsFiletype := uint8(4) // regular file
	fsFlags := uint16(0)
	pad := uint8(0)
	fsRightsBase := uint64(0)
	fsRightsInheriting := uint64(0)

	// We'll write them out to memory in order:
	//  offset 0: fs_filetype (1 byte)
	//  offset 1: fs_flags (2 bytes)
	//  offset 3: pad (1 byte)
	//  offset 4: fs_rights_base (8 bytes)
	//  offset 12: fs_rights_inheriting (8 bytes)
	//  total => 20 bytes, but alignment might push it to 24.

	err = mem.Write(fdstatPtr, []byte{fsFiletype})
	if err != nil {
		return nil, err
	}
	err = wasimem.WriteUint16Le(mem, fdstatPtr+1, fsFlags)
	if err != nil {
		return nil, err
	}
	err = mem.Write(fdstatPtr+3, []byte{pad})
	if err != nil {
		return nil, err
	}
	err = wasimem.WriteUint64Le(mem, fdstatPtr+4, fsRightsBase)
	if err != nil {
		return nil, err
	}

	// 5) fs_rights_inheriting
	err = wasimem.WriteUint64Le(mem, fdstatPtr+12, fsRightsInheriting)
	if err != nil {
		return nil, err
	}

	returns[0] = int32(0)
	return returns, nil
}

// 12) fd_fdstat_set_flags(fd: fd, flags: fdflags) -> errno
func wasi_fdFdstatSetFlags(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdFdstatSetFlags", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 13) fd_fdstat_set_rights(fd: fd, fs_rights_base: rights, fs_rights_inheriting: rights) -> errno
func wasi_fdFdstatSetRights(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdFdstatSetRights", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 14) fd_filestat_get(fd: fd) -> (errno, filestat)
func wasi_fdFilestatGet(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdFilestatGet", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 15) fd_filestat_set_size(fd: fd, size: filesize) -> errno
func wasi_fdFilestatSetSize(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdFilestatSetSize", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 16) fd_filestat_set_times(fd: fd, atim: timestamp, mtim: timestamp, fst_flags: fstflags) -> errno
func wasi_fdFilestatSetTimes(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdFilestatSetTimes", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 17) fd_pread(fd: fd, iovs: iovec_array, offset: filesize) -> (errno, size)
func wasi_fdPread(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdPread", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 18) fd_prestat_get(fd: fd) -> (errno, prestat)
func wasi_fdPrestatGet(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdPrestatGet", "params", params)
	returns := make([]interface{}, 1)
	vm := rnh.GetVm()
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	fd := params[0].(int32)
	resptr := params[1].(int32)
	preopens := ctx.GetOpenFiles(vm)

	// // By convention, FDs 0, 1, 2 = stdin, stdout, stderr
	// if fd < 3 || fd > int32(len(preopens)+2) {
	// 	returns[0] = int32(8) // EBADF = 8 in WASI
	// 	return returns, nil
	// }
	// preopen := preopens[fd-3]
	preopen, ok := preopens[fd]
	if !ok {
		returns[0] = int32(8) // EBADF = 8 in WASI
		return returns, nil
	}

	// Write a __wasi_prestat for a directory:
	//   pr_type = 0 (directory)
	//   3 bytes of zero padding
	//   pr_name_len = nameLen

	// pr_type (1 byte) + 3 bytes of 0 padding
	// 0 => __WASI_PREOPENTYPE_DIR
	err = mem.Write(resptr, []byte{0, 0, 0, 0})
	if err != nil {
		return nil, err
	}

	// pr_name_len (4 bytes)
	err = wasimem.WriteUint32Le(mem, resptr+4, uint32(len(preopen.path)))
	if err != nil {
		return nil, err
	}

	returns[0] = int32(0)
	return returns, nil
}

// 19) fd_prestat_dir_name(fd: fd, path: Pointer<u8>, path_len: size) -> errno
func wasi_fdPrestatDirName(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdPrestatDirName", "params", params)
	returns := make([]interface{}, 1)
	vm := rnh.GetVm()
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}

	fd := params[0].(int32)
	pathPtr := params[1].(int32)
	pathLen := params[2].(int32)

	preopens := ctx.GetOpenFiles(vm)
	// if fd < 3 || fd > int32(len(preopens)+2) {
	// 	returns[0] = int32(8) // EBADF
	// 	return returns, nil
	// }
	// preopen := preopens[fd-3]

	preopen, ok := preopens[fd]
	if !ok {
		returns[0] = int32(8) // EBADF = 8 in WASI
		return returns, nil
	}

	if int32(len(preopen.path)) > pathLen {
		// EOVERFLOW if there's not enough space to write the full path
		returns[0] = int32(75) // EOVERFLOW = 75
		return returns, nil
	}

	// Write the directory path into [pathPtr, pathPtr+len(preopen)]
	err = mem.Write(pathPtr, []byte(preopen.path))
	if err != nil {
		return nil, err
	}

	returns[0] = int32(0)
	return returns, nil
}

// 20) fd_pwrite(fd: fd, iovs: ciovec_array, offset: filesize) -> (errno, size)
func wasi_fdPwrite(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdPwrite", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 21) fd_read(fd: fd, iovs: iovec_array) -> (errno, size)
func wasi_fdRead(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdRead", "params", params)
	fd := params[0].(int32)       // file descriptor
	iovsPtr := params[1].(int32)  // pointer to array of iovec structs
	iovsLen := params[2].(int32)  // number of iovecs
	outNread := params[3].(int32) // pointer in memory where we store bytes-read

	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}

	fileMap := ctx.GetFileMap(rnh.GetVm())
	openFiles := ctx.GetOpenFiles(rnh.GetVm())

	// 4) Check if fd is valid
	openF, ok := openFiles[fd]
	returns := make([]interface{}, 1) // We'll return just errno: i32
	if !ok {
		// EBADF => 8
		returns[0] = int32(8)
		return returns, nil
	}

	// 5) Get the file content
	content, ok := fileMap[openF.path]
	if !ok {
		// ENOENT => 44
		returns[0] = int32(44)
		return returns, nil
	}

	contentBytes := []byte(content)
	var totalRead int64

	// 6) Iterate over each iovec
	iovOffset := iovsPtr
	for i := int32(0); i < iovsLen; i++ {
		// Each iovec is (bufPtr: u32, bufLen: u32), 8 bytes total in 32-bit
		bufPtr, err1 := wasimem.ReadUint32Le(mem, iovOffset)
		if err1 != nil {
			return nil, err1
		}
		bufLen, err2 := wasimem.ReadUint32Le(mem, iovOffset+4)
		if err2 != nil {
			return nil, err2
		}
		iovOffset += 8

		// If we've already reached or passed EOF, read 0 bytes into this iovec
		if openF.offset >= int64(len(contentBytes)) {
			// Nothing more to read
			break
		}

		// Calculate how many bytes we can read
		bytesAvailable := int64(len(contentBytes)) - openF.offset
		toRead := int64(bufLen)
		if toRead > bytesAvailable {
			toRead = bytesAvailable
		}

		// Slice out [offset : offset+toRead]
		chunk := contentBytes[openF.offset : openF.offset+toRead]

		// Write that chunk into wasm memory at bufPtr
		if err3 := mem.Write(int32(bufPtr), chunk); err3 != nil {
			return nil, err3
		}

		// Update offset and totalRead
		openF.offset += toRead
		totalRead += toRead
	}

	// 7) Write the total number of bytes read into out_nread
	//    Make sure to clamp it to a 32-bit if needed
	if totalRead > (1<<32 - 1) {
		// If you want, handle an overflow scenario, but usually very large reads won't happen in practice
		totalRead = (1<<32 - 1)
	}
	if err := wasimem.WriteUint32Le(mem, outNread, uint32(totalRead)); err != nil {
		return nil, err
	}

	// 8) Return errno=0 (success)
	returns[0] = int32(0)
	return returns, nil
}

// 22) fd_readdir(fd: fd, buf: Pointer<u8>, buf_len: size, cookie: dircookie) -> (errno, size)
func wasi_fdReaddir(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdReaddir", "params", params)
	fd := params[0].(int32)
	bufPtr := params[1].(int32)
	bufLen := params[2].(int32)
	cookie := params[3].(int64)
	outNread := params[4].(int32)

	// We'll return just one i32 (errno)
	returns := make([]interface{}, 1)

	mem, err := rnh.GetMemory()
	if err != nil {
		// EFAULT => 21 if we can't access memory
		returns[0] = int32(21)
		return returns, nil
	}

	of, ok := ctx.openFiles[fd]
	if !ok {
		// EBADF => 8
		returns[0] = int32(8)
		return returns, nil
	}
	if !of.isdir {
		// ENOTDIR => 54
		returns[0] = int32(54)
		return returns, nil
	}

	// We'll gather the "entries" in this directory. For a simple approach,
	// treat everything in DirMapping or FileMapping that starts with of.path + "/" as a child.
	// Then we skip the part equal to of.path + "/", and keep the remainder as the name.

	entries := listDirectoryEntries(ctx, of.path)

	// Sort entries by name if you want stable ordering:
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	// The 'cookie' effectively is the index in `entries`. If cookie=2, skip first 2 entries, etc.
	startIndex := int(cookie)
	if startIndex < 0 {
		startIndex = 0
	}
	if startIndex > len(entries) {
		// If the cookie is beyond the end, we write nothing.
		// It's valid to just return 0 bytes read, errno=0
		_ = wasimem.WriteUint32Le(mem, outNread, 0)
		returns[0] = int32(0)
		return returns, nil
	}

	var totalWritten int32
	currentOffset := bufPtr

	// We'll define a 24-byte struct for each dirent, then append the file name bytes
	const direntSize = 24 // 8(d_next) + 8(d_ino) + 4(d_namlen) + 1(d_type) + 3(padding)

	for i := startIndex; i < len(entries); i++ {
		e := entries[i]

		nameLen := len(e.Name)
		recordSize := direntSize + int32(nameLen) // total bytes for this entry

		// Check if we have enough space in the buffer
		if totalWritten+recordSize > bufLen {
			// Not enough room => stop listing
			break
		}

		// 1) Write d_next (8 bytes, i+1 => next cookie)
		dNext := uint64(i + 1)
		if err := wasimem.WriteUint64Le(mem, currentOffset, dNext); err != nil {
			returns[0] = int32(21) // EFAULT
			return returns, nil
		}

		// 2) Write d_ino (8 bytes). We'll just use 0 or a placeholder
		dIno := uint64(0)
		if err := wasimem.WriteUint64Le(mem, currentOffset+8, dIno); err != nil {
			returns[0] = int32(21)
			return returns, nil
		}

		// 3) Write d_namlen (4 bytes)
		if err := wasimem.WriteUint32Le(mem, currentOffset+16, uint32(nameLen)); err != nil {
			returns[0] = int32(21)
			return returns, nil
		}

		// 4) Write d_type (1 byte). 3=dir, 4=file
		filetype := uint8(4)
		if e.IsDir {
			filetype = 3
		}
		if err := mem.Write(currentOffset+20, []byte{filetype}); err != nil {
			returns[0] = int32(21)
			return returns, nil
		}
		// The 3 padding bytes [21..23] remain zero by default if your mem.Write(...) starts from zero memory.
		// Or we can explicitly write them as zero if you want.

		// 5) Write the name bytes right after the 24-byte struct
		nameOffset := currentOffset + direntSize
		if err := mem.Write(nameOffset, []byte(e.Name)); err != nil {
			returns[0] = int32(21)
			return returns, nil
		}

		// Advance
		currentOffset += recordSize
		totalWritten += recordSize
	}

	// Write how many bytes we wrote in total to outNread
	if err := wasimem.WriteUint32Le(mem, outNread, uint32(totalWritten)); err != nil {
		returns[0] = int32(21)
		return returns, nil
	}

	// Return success
	returns[0] = int32(0)
	return returns, nil
}

// 23) fd_renumber(fd: fd, to: fd) -> errno
func wasi_fdRenumber(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdRenumber", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 24) fd_seek(fd: fd, offset: filedelta, whence: whence) -> (errno, filesize)
func wasi_fdSeek(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdSeek", "params", params)

	fd := params[0].(int32)
	offset := params[1].(int64)
	whence := params[2].(int32)
	resultPtr := params[3].(int32) // where to store the new offset in WASM memory

	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}

	fileMap := ctx.GetFileMap(rnh.GetVm())
	openFiles := ctx.GetOpenFiles(rnh.GetVm())

	openF, ok := openFiles[fd]
	returns := make([]interface{}, 1)

	if !ok {
		// EBADF => 8
		returns[0] = int32(8)
		return returns, nil
	}

	// 4) Compute new offset
	var base int64
	switch whence {
	case 0: // SEEK_SET
		base = 0
	case 1: // SEEK_CUR
		base = openF.offset
	case 2: // SEEK_END
		// If file is in ctx.FileMapping, get its length
		fileContent, ok := fileMap[openF.path]
		if !ok {
			// EIO => 29, or maybe ENOENT => 44
			returns[0] = int32(44)
			return returns, nil
		}
		base = int64(len(fileContent))
	default:
		// EINVAL => 28
		returns[0] = int32(28)
		return returns, nil
	}

	newOffset := base + offset
	if newOffset < 0 {
		// ESPIPE => 70 or EINVAL => 28, depending on how you want to handle negative seeks
		returns[0] = int32(28)
		return returns, nil
	}

	// 5) Update offset in openFile
	openF.offset = newOffset

	// 6) Write new offset to memory at resultPtr (uint64)
	// Using your “wasimem” helper or similar:
	err = wasimem.WriteUint64Le(mem, resultPtr, uint64(newOffset))
	if err != nil {
		return nil, err
	}

	// success => errno=0
	returns[0] = int32(0)
	return returns, nil

}

// 25) fd_sync(fd: fd) -> errno
func wasi_fdSync(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdSync", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 26) fd_tell(fd: fd) -> (errno, filesize)
func wasi_fdTell(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdTell", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 27) fd_write(fd: fd, iovs: ciovec_array) -> (errno, size)
func wasi_fdWrite(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_fdWrite", "params", params)
	fd := params[0].(int32)
	iovsPtr := params[1].(int32) // pointer to the iovec array
	iovsLen := params[2].(int32) // number of iovecs
	nwrittenPtr := params[3].(int32)

	returns := make([]interface{}, 1)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}

	vm := rnh.GetVm()
	fileMap := ctx.GetFileMap(vm)
	openFiles := ctx.GetOpenFiles(vm)

	openF, ok := openFiles[fd]
	if !ok {
		// EBADF => 8 (Bad file descriptor)
		returns[0] = int32(8)
		return returns, nil
	}

	// 4) Find the file’s content in our in-memory mapping
	content, ok := fileMap[openF.path]
	if !ok {
		// ENOENT => 44 (No such file or directory)
		returns[0] = int32(44)
		return returns, nil
	}

	var totalWritten int64 = 0

	// 5) Each WASI `__wasi_ciovec_t` is (buf_ptr: u32, buf_len: u32) in a 32-bit environment
	//    so each element is 8 bytes. We'll iterate iovsLen times, reading from memory.
	iovOffset := iovsPtr

	for i := int32(0); i < iovsLen; i++ {
		// read buf_ptr
		bufPtr, err1 := wasimem.ReadUint32Le(mem, iovOffset)
		if err1 != nil {
			return nil, err1
		}
		// read buf_len
		bufLen, err2 := wasimem.ReadUint32Le(mem, iovOffset+4)
		if err2 != nil {
			return nil, err2
		}
		iovOffset += 8

		// 6) Read the actual data from the Wasm memory
		data, err3 := mem.Read(int32(bufPtr), int32(bufLen))
		if err3 != nil {
			return nil, err3
		}

		// 7) Write into our in-memory file at the current offset
		start := openF.offset
		end := start + int64(bufLen)

		// If end extends beyond the current content, expand it
		if end > int64(len(content)) {
			extra := int(end) - len(content)
			content = append(content, make([]byte, extra)...) // Extend with zero bytes
		}

		// Replace that segment
		contentBytes := []byte(content)
		copy(contentBytes[start:end], data)
		content = contentBytes

		// Advance file offset
		openF.offset = end

		// Tally how many bytes we wrote this iteration
		totalWritten += int64(bufLen)
	}

	// 8) Update the in-memory file content
	ctx.SetFileMap(vm, openF.path, content)

	// 9) Write totalWritten into `nwrittenPtr`
	if err := wasimem.WriteUint32Le(mem, nwrittenPtr, uint32(totalWritten)); err != nil {
		return nil, err
	}

	if openF.path == "stderr" {
		ctx.c.Logger(ctx.c.Ctx).Error(string(content))
	}
	if openF.path == "stdout" {
		LoggerExtended(ctx.c).Debug(string(content))
	}

	returns[0] = int32(0)
	return returns, nil
}

// 28) path_create_directory(fd: fd, path: string) -> errno
func wasi_pathCreateDirectory(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_pathCreateDirectory", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 29) path_filestat_get(fd: fd, flags: lookupflags, path: string) -> (errno, filestat)
func wasi_pathFilestatGet(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_pathFilestatGet", "params", params)
	fd := params[0].(int32)        // directory FD
	_ = params[1].(int32)          // WASI lookup flags (we might ignore here)
	pathPtr := params[2].(int32)   // pointer to the path string in guest memory
	pathLen := params[3].(int32)   // length of that path
	resultPtr := params[4].(int32) // where to write the filestat

	returns := make([]interface{}, 1)
	vm := rnh.GetVm()
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}

	// 4) Read the actual path string from the guest memory
	pathBytes, err := mem.Read(pathPtr, pathLen)
	if err != nil {
		// EFAULT => invalid memory => 21 in WASI, or something appropriate
		returns[0] = int32(21)
		return returns, nil
	}
	guestPath := string(pathBytes)

	// 5) Look up the directory FD to confirm it's valid
	//    Usually you'd have a map like: fd => preopenDirPath
	preopens := ctx.GetOpenFiles(vm)
	fileMap := ctx.GetFileMap(vm)
	preopen, ok := preopens[fd]
	// preopenDir, ok := ctx.PreopenDirs[fd]
	if !ok {
		// EBADF => 8
		returns[0] = int32(8)
		return returns, nil
	}

	// 6) Combine preopenDir + guestPath to find the real host path or
	//    your in-memory representation. For a simple example, just do:
	fullPath := preopen.path
	if guestPath != "." {
		fullPath = preopen.path + "/" + guestPath
	}

	// fullPath = filepath.Clean(fullPath) ?

	// 7) If the file doesn't exist, return ENOENT => 44
	fileContent, ok := fileMap[fullPath]
	if !ok {
		// Possibly also check if it's a directory if you store them separately
		returns[0] = int32(44)
		return returns, nil
	}

	isDir := false
	if ctx.dirMapping[fullPath] {
		isDir = true
	} else if _, fileOk := fileMap[fullPath]; fileOk {
		isDir = false
	} else {
		// Doesn't exist
		returns[0] = int32(44) // ENOENT
		return returns, nil
	}

	// 8) Prepare filestat fields
	// We'll do a minimal approach:
	//   dev=0, ino=0, filetype=__WASI_FILETYPE_REGULAR_FILE=4, size=len(fileContent)
	//   times=0, nlink=1, etc.

	var dev uint64 = 0
	var ino uint64 = 0
	var nlink uint64 = 1
	size := uint64(len(fileContent))
	var atim uint64 = 0
	var mtim uint64 = 0
	var ctim uint64 = 0

	// __WASI_FILETYPE_REGULAR_FILE=4, directory=3, etc.
	var filetype uint8 = 4
	if isDir {
		filetype = uint8(3)
	}

	// 9) Write the struct into memory.
	//    The layout (in bytes) typically is:
	//    0..7   dev (u64)
	//    8..15  ino (u64)
	//    16     filetype (u8)
	//    17..23 padding
	//    24..31 nlink (u64)
	//    32..39 size (u64)
	//    40..47 atim (u64)
	//    48..55 mtim (u64)
	//    56..63 ctim (u64) - total 64 bytes (some WASI docs say 56, but alignment might push it to 64).
	//    Check your specific environment if it’s 56 or 64 total.

	// Let's define offsets as constants for clarity:
	const (
		offsetDev      = 0
		offsetIno      = 8
		offsetFiletype = 16
		offsetNlink    = 24
		offsetSize     = 32
		offsetAtim     = 40
		offsetMtim     = 48
		offsetCtim     = 56
	)

	// Write dev (u64)
	if err := wasimem.WriteUint64Le(mem, resultPtr+offsetDev, dev); err != nil {
		return nil, err
	}

	// Write ino (u64)
	if err := wasimem.WriteUint64Le(mem, resultPtr+offsetIno, ino); err != nil {
		return nil, err
	}

	// Write filetype (u8)
	if err := mem.Write(resultPtr+offsetFiletype, []byte{filetype}); err != nil {
		return nil, err
	}

	// nlink (u64)
	if err := wasimem.WriteUint64Le(mem, resultPtr+offsetNlink, nlink); err != nil {
		return nil, err
	}

	// size (u64)
	if err := wasimem.WriteUint64Le(mem, resultPtr+offsetSize, size); err != nil {
		return nil, err
	}

	// atim (u64)
	if err := wasimem.WriteUint64Le(mem, resultPtr+offsetAtim, atim); err != nil {
		return nil, err
	}

	// mtim (u64)
	if err := wasimem.WriteUint64Le(mem, resultPtr+offsetMtim, mtim); err != nil {
		return nil, err
	}

	// ctim (u64)
	if err := wasimem.WriteUint64Le(mem, resultPtr+offsetCtim, ctim); err != nil {
		return nil, err
	}

	// 10) Return errno=0 for success
	returns[0] = int32(0)
	return returns, nil
}

//  30. path_filestat_set_times(fd: fd, flags: lookupflags, path: string,
//     atim: timestamp, mtim: timestamp, fst_flags: fstflags) -> errno
func wasi_pathFilestatSetTimes(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_pathFilestatSetTimes", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

//  31. path_link(old_fd: fd, old_flags: lookupflags, old_path: string,
//     new_fd: fd, new_path: string) -> errno
func wasi_pathLink(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_pathLink", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

//  32. path_open(fd: fd, dirflags: lookupflags, path: string,
//     oflags: oflags, fs_rights_base: rights, fs_rights_inheriting: rights,
//     fdflags: fdflags) -> (errno, fd)
func wasi_pathOpen(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_pathOpen", "params", params)
	// The declared parameters are:
	//  1) dirFd (i32)
	//  2) dirFlags (i32)
	//  3) pathPtr (i32)
	//  4) pathLen (i32)
	//  5) oflags (i32)
	//  6) fsRightsBase (i64)
	//  7) fsRightsInheriting (i64)
	//  8) fdflags (i32)
	//  9) newFdPtr (i32) - pointer in guest memory to store the newly opened FD

	dirFd := params[0].(int32)
	// dirFlags := params[1].(int32)
	pathPtr := params[2].(int32)
	pathLen := params[3].(int32)
	oflags := params[4].(int32)
	// fsRightsBase := params[5].(int64)
	// fsRightsInheriting := params[6].(int64)
	// fdflags := params[7].(int32)
	newFdPtr := params[8].(int32)

	returns := make([]interface{}, 1)

	mem, err := rnh.GetMemory()
	if err != nil {
		returns[0] = int32(21)
		return returns, nil
	}

	vm := rnh.GetVm()
	preopens := ctx.GetOpenFiles(vm)
	fileMap := ctx.GetFileMap(vm)

	// 3) Read the path string from the guest memory
	pathBytes, err := mem.Read(pathPtr, pathLen)
	if err != nil {
		// Return EFAULT => 21
		returns[0] = int32(21)
		return returns, nil
	}
	guestPath := string(pathBytes)

	// 4) Look up the directory FD in PreopenDirs (or a map of open directories).
	baseDirPath, ok := preopens[dirFd]
	if !ok || !baseDirPath.isdir {
		// EBADF => 8
		returns[0] = int32(8)
		return returns, nil
	}

	// 5) Combine baseDirPath + guestPath to find the real path
	//    or the key used in your in-memory FS.
	fullPath := baseDirPath.path
	if guestPath != "" && guestPath != "." {
		if fullPath != "" && !endsWithSlash(fullPath) {
			fullPath += "/"
		}
		fullPath += guestPath
	}

	// 6) Check if the guest wants to open a directory (__WASI_O_DIRECTORY)
	wantsDir := (oflags & __WASI_O_DIRECTORY) != 0

	// 7) Check if the path exists in your in-memory representation.
	//    If it doesn't exist and O_CREAT is set, you might create it.
	//    Or if it's a directory, check DirMapping, etc.
	isDir := false
	if ctx.dirMapping[fullPath] {
		isDir = true
	} else if _, fileOk := fileMap[fullPath]; fileOk {
		isDir = false
	} else {
		// Doesn't exist
		if (oflags&__WASI_O_CREAT) != 0 && !wantsDir {
			// create an empty file
			ctx.fileMapping[fullPath] = []byte{}
		} else {
			// ENOENT => 44
			returns[0] = int32(44)
			return returns, nil
		}
	}

	// If the path is a directory but the guest wanted a file, or vice versa, handle that:
	if isDir && !wantsDir {
		// EISDIR => 31
		returns[0] = int32(31)
		return returns, nil
	}
	if !isDir && wantsDir {
		// ENOTDIR => 54
		returns[0] = int32(54)
		return returns, nil
	}

	// 8) Create a new FD for the opened file/dir
	newFd := ctx.nextfd
	ctx.nextfd++

	// Insert into OpenFiles map
	ctx.SetOpenFile(vm, newFd, &openFile{
		path:   fullPath,
		offset: 0,
		isdir:  isDir,
		// store other flags if needed
	})

	// If you want to handle read/write perms, you’d store fsRightsBase or fdflags, etc.

	// 9) Write the new FD into guest memory at newFdPtr
	if err := wasimem.WriteUint32Le(mem, newFdPtr, uint32(newFd)); err != nil {
		// EFAULT => 21 or some other code
		returns[0] = int32(21)
		return returns, nil
	}

	returns[0] = int32(0)
	return returns, nil
}

// 33) path_readlink(fd: fd, path: string, buf: Pointer<u8>, buf_len: size) -> (errno, size)
func wasi_pathReadlink(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_pathReadlink", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 34) path_remove_directory(fd: fd, path: string) -> errno
func wasi_pathRemoveDirectory(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_pathRemoveDirectory", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 35) path_rename(fd: fd, old_path: string, new_fd: fd, new_path: string) -> errno
func wasi_pathRename(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_pathRename", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 36) path_symlink(old_path: string, fd: fd, new_path: string) -> errno
func wasi_pathSymlink(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_pathSymlink", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 37) path_unlink_file(fd: fd, path: string) -> errno
func wasi_pathUnlinkFile(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_pathUnlinkFile", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 38) poll_oneoff(in: ConstPointer<subscription>, out: Pointer<event>, nsubscriptions: size) -> (errno, size)
func wasi_pollOneoff(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_pollOneoff", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 39) proc_exit(rval: exitcode)
func wasi_procExit(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_procExit", "params", params)
	// Typically triggers a runtime exit, but here we just log and return.
	return nil, nil
}

// 40) proc_raise(sig: signal) -> errno
func wasi_procRaise(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_procRaise", "params", params)
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	return returns, nil
}

// 41) sched_yield() -> errno
func wasi_schedYield(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_schedYield", "params", params)
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	return returns, nil
}

// 42) random_get(buf: Pointer<u8>, buf_len: size) -> errno
func wasi_randomGet(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_randomGet", "params", params)
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	return returns, nil
}

func wasi_sockAccept(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_sockAccept", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 43) sock_recv(fd: fd, ri_data: iovec_array, ri_flags: riflags) -> (errno, size, roflags)
func wasi_sockRecv(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_sockRecv", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 44) sock_send(fd: fd, si_data: ciovec_array, si_flags: siflags) -> (errno, size)
func wasi_sockSend(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_sockSend", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

// 45) sock_shutdown(fd: fd, how: sdflags) -> errno
func wasi_sockShutdown(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_sockShutdown", "params", params)
	return wasi_stubUnimplemented(_context, rnh, params)
}

type DirEntry struct {
	Name  string
	IsDir bool
}

type openFile struct {
	isdir  bool
	path   string
	offset int64
	// add other fields if needed (read/write permissions, etc.)
}

type WasiContext struct {
	c      *Context
	inited bool
	nextfd int32
	// fd => openFile
	// used for preopened folders and files
	openFiles map[int32]*openFile
	// these are temporary execution changes
	// filepath => content
	fileMapping map[string][]byte // contains dir paths with empty content
	dirMapping  map[string]bool
}

func (wc *WasiContext) InitContext(vm memc.IVm) {
	// init stdin, stdout, stderr
	wc.fileMapping["stdin"] = []byte{}
	wc.fileMapping["stdout"] = []byte{}
	wc.fileMapping["stderr"] = []byte{}
	wc.openFiles[int32(0)] = &openFile{path: "stdin"}
	wc.openFiles[int32(1)] = &openFile{path: "stdout"}
	wc.openFiles[int32(2)] = &openFile{path: "stderr"}

	// initialize filedescriptor map with wasi config preopened folders
	preopens := vm.WasiPreopens()
	externalToInner := map[string]string{}
	for i, pair := range preopens {
		// e.g. some/wasi/path:some/absolute/dir/path
		parts := strings.Split(pair, ":")
		externalToInner[parts[1]] = parts[0]
		wc.openFiles[int32(3+i)] = &openFile{path: parts[0], isdir: true}
		wc.fileMapping[parts[0]] = []byte{}
		wc.dirMapping[parts[0]] = true
	}

	// initialize filemap with wasi config filemap
	fileMap := vm.WasiFileMap()
	for externalPath := range fileMap {
		path := externalPath
		for externalDir, innerDir := range externalToInner {
			if strings.Contains(path, externalDir) {
				path = strings.Replace(path, externalDir, innerDir, int(1))
				break
			}
		}
		if _, ok := wc.fileMapping[path]; !ok {
			wc.fileMapping[path] = fileMap[externalPath]
		}
	}

	wc.nextfd = int32(3 + len(preopens))
	wc.inited = true
}

func (wc *WasiContext) GetFileMap(vm memc.IVm) map[string][]byte {
	if !wc.inited {
		wc.InitContext(vm)
	}
	return wc.fileMapping
}

func (wc *WasiContext) SetFileMap(vm memc.IVm, path string, content []byte) {
	if !wc.inited {
		wc.InitContext(vm)
	}
	wc.fileMapping[path] = content
}

func (wc *WasiContext) GetOpenFiles(vm memc.IVm) map[int32]*openFile {
	if !wc.inited {
		wc.InitContext(vm)
	}
	return wc.openFiles
}

func (wc *WasiContext) SetOpenFile(vm memc.IVm, fd int32, f *openFile) {
	if !wc.inited {
		wc.InitContext(vm)
	}
	wc.openFiles[fd] = f
}

func BuildWasiEnv(_context *Context, rnh memc.RuntimeHandler) (interface{}, error) {
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
			wasi_clockTimeGet,
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
			wasi_randomGet,
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

func endsWithSlash(s string) bool {
	return len(s) > 0 && s[len(s)-1] == '/'
}

func listDirectoryEntries(ctx *WasiContext, basePath string) []DirEntry {
	var result []DirEntry

	prefix := basePath
	if !endsWithSlash(prefix) {
		prefix += "/"
	}

	// Add '.' and '..' entries
	result = append(result, DirEntry{Name: ".", IsDir: true})
	result = append(result, DirEntry{Name: "..", IsDir: true})

	// We'll gather every path in DirMapping/FileMapping that starts with prefix.
	// Then the part after prefix is the name. If that name has further slashes, it’s in a subdir.

	// For directories
	for dir := range ctx.dirMapping {
		if strings.HasPrefix(dir, prefix) {
			remainder := dir[len(prefix):]
			// If remainder has slashes, it's in a subdirectory, so skip if you only want immediate children
			if !strings.ContainsRune(remainder, '/') && remainder != "" {
				result = append(result, DirEntry{Name: remainder, IsDir: true})
			}
		}
	}
	// For files
	for file := range ctx.fileMapping {
		_, isDir := ctx.dirMapping[file]
		if strings.HasPrefix(file, prefix) && !isDir {
			remainder := file[len(prefix):]
			if !strings.ContainsRune(remainder, '/') && remainder != "" {
				result = append(result, DirEntry{Name: remainder, IsDir: false})
			}
		}
	}
	return result
}
