package utils

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

func Itoa(i int) string      { return strconv.FormatInt(int64(i), 10) }
func U64toa(i uint64) string { return strconv.FormatUint(i, 10) }

func ParseUint8ArrayToI32BigEndian(b []byte) (int32, error) {
	if len(b) < 4 {
		return 0, fmt.Errorf("buffer too small: got %d, need 4", len(b))
	}
	u := binary.BigEndian.Uint32(b[len(b)-4:])
	return int32(u), nil // interpret same bits as signed
}
