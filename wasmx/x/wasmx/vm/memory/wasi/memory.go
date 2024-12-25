package wasi

import (
	"encoding/binary"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func WriteUint16Le(mem memc.IMemory, offset int32, v uint16) error {
	bz := make([]byte, 2)
	binary.LittleEndian.PutUint16(bz, v)
	return mem.WriteRaw(int32(offset), bz)
}

func WriteUint32Le(mem memc.IMemory, offset int32, v uint32) error {
	bz := make([]byte, 4)
	binary.LittleEndian.PutUint32(bz, v)
	return mem.WriteRaw(int32(offset), bz)
}

func WriteUint64Le(mem memc.IMemory, offset int32, v uint64) error {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, v)
	return mem.WriteRaw(int32(offset), bz)
}

func ReadUint32Le(mem memc.IMemory, offset int32) (uint32, error) {
	data, err := mem.Read(offset, 4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(data), nil
}
