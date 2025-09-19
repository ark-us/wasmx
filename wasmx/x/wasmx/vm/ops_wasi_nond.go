package vm

import (
	"crypto/rand"
	"io"
	"time"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
	wasimem "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/wasi"
)

// 42) random_get(buf: Pointer<u8>, buf_len: size) -> errno
func wasi_randomGet_NonD(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	if len(params) != 2 {
		return []interface{}{WasiErrnoInval}, nil
	}
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_randomGet", "params", params)
	returns := make([]interface{}, 1)
	mem, err := rnh.GetMemory()
	if err != nil {
		returns[0] = WasiErrnoFault
		return returns, nil
	}

	bufPtr, ok := params[0].(int32)
	if !ok {
		return []interface{}{WasiErrnoInval}, nil
	}
	bufLen, ok := params[1].(int32)
	if !ok {
		return []interface{}{WasiErrnoInval}, nil
	}
	if bufLen == 0 {
		return []interface{}{WasiErrnoSuccess}, nil
	}

	// Fill in chunks to avoid huge allocations.
	const chunkSize = 64 * 1024
	remaining := bufLen
	offset := int32(0)
	for remaining > 0 {
		n := int(remaining)
		if n > chunkSize {
			n = chunkSize
		}
		tmp := make([]byte, n)
		if _, err := io.ReadFull(rand.Reader, tmp); err != nil {
			return []interface{}{WasiErrnoIo}, nil
		}
		if err := mem.Write(int32(bufPtr+offset), tmp); err != nil {
			return []interface{}{WasiErrnoFault}, nil
		}
		remaining -= int32(n)
		offset += int32(n)
	}
	returns[0] = WasiErrnoSuccess
	return returns, nil
}

// this is non-deterministic
// we need it for consensus algorithms & other single consensus contracts
// 6) clock_time_get(id: clockid, precision: timestamp) -> (errno, timestamp)
func wasi_clockTimeGet_NonD(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	// TODO move!
	ctx := _context.(*WasiContext)
	LoggerExtended(ctx.c).Debug("wasi_clockTimeGet", "params", params)
	returns := make([]interface{}, 1)

	// Params as per WASI:
	//   0: clockid (u32)
	//   1: precision (u64)  -- hosts usually ignore; you can round if you want
	//   2: result pointer (u32) to write u64 ns
	if len(params) != 3 {
		// Return EINVAL if your dispatcher allows.
		returns[0] = WasiErrnoInval
		return returns, nil
	}
	clockid := params[0].(int32)
	// precision, _ := params[1].(int64)
	resultPtr := params[2].(int32)

	var ns uint64
	switch clockid {
	case WasiClockRealtime:
		// Nanoseconds since UNIX epoch, UTC.
		ns = uint64(time.Now().UTC().UnixNano())
	default:
		return []interface{}{WasiErrnoNotSup}, nil
	}

	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	err = wasimem.WriteUint64Le(mem, resultPtr, ns)
	if err != nil {
		return nil, err
	}
	return []interface{}{WasiErrnoSuccess}, nil
}
