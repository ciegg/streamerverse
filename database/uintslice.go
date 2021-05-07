package database

import (
	"hash/crc64"
)

var table = crc64.MakeTable(crc64.ECMA)

func HashViewers(chatroom []string) []int64 {
	compressed := make([]int64, 0, len(chatroom))
	for _, chatter := range chatroom {
		hash := crc64.Checksum([]byte(chatter), table)
		compressed = append(compressed, int64(hash))
	}

	return compressed
}
