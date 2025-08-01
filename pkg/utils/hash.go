package utils

import (
	"fmt"
	"hash/fnv"
)

func CalcFnvHash(src string, prefix string, div uint32, mod uint32) (uint32, string) {
	h := fnv.New32a()
	h.Write([]byte(src))
	hashValue := h.Sum32()
	tableIndex := hashValue % (div * mod) // 001-010-020-128
	dbx := tableIndex / mod
	tbx := tableIndex - dbx*mod
	return dbx, fmt.Sprintf("%s_%04d", prefix, tbx)
}
