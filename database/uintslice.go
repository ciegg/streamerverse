package database

import (
	"hash/crc64"
	"sort"
)

var table = crc64.MakeTable(crc64.ECMA)

type UintSlice []uint64

func (x UintSlice) Len() int           { return len(x) }
func (x UintSlice) Less(i, j int) bool { return x[i] < x[j] }
func (x UintSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x UintSlice) Search(a uint64) int {
	return sort.Search(x.Len(), func(i int) bool { return x[i] >= a })
}
func (x UintSlice) Sort() { sort.Sort(x) }

func NewUintSlice(chatroom []string) UintSlice {
	compressed := make([]uint64, 0, len(chatroom))
	for _, chatter := range chatroom {
		hash := crc64.Checksum([]byte(chatter), table)
		compressed = append(compressed, hash)
	}

	return compressed
}

/*
func Compare(aChatroom, bChatroom UintSlice) int {
	var smaller UintSlice
	var bigger UintSlice

	if aChatroom.Len() > bChatroom.Len() {
		bigger = aChatroom
		smaller = bChatroom
	} else {
		bigger = bChatroom
		smaller = aChatroom
	}

	count := 0

	for _, chatter := range smaller {
		index := bigger.Search(chatter)
		if index < bigger.Len() && bigger[index] == chatter {
			count += 1
		}
	}

	return count
}
*/
