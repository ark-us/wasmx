package utils

import "strconv"

func Itoa(i int) string      { return strconv.FormatInt(int64(i), 10) }
func U64toa(i uint64) string { return strconv.FormatUint(i, 10) }
