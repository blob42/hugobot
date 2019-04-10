package utils

import "encoding/binary"

func IntToBytes(x int64) []byte {
	buf := make([]byte, 4)
	binary.PutVarint(buf, x)
	return buf
}
