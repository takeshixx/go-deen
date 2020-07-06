package hashs

import (
	"hash/adler32"
	"hash/crc32"
	"hash/crc64"
)

func CalculateAdler32(data []byte) (checksum uint32) {
	checksum = adler32.Checksum(data)
	return
}

func CalculateCrc32IEEE(data []byte) (checksum uint32) {
	checksum = crc32.ChecksumIEEE(data)
	return
}

func CalculateCrc32Castagnoli(data []byte) (checksum uint32) {
	tab := crc32.MakeTable(crc32.Castagnoli)
	checksum = crc32.Checksum(data, tab)
	return
}

func CalculateCrc32Koopman(data []byte) (checksum uint32) {
	tab := crc32.MakeTable(crc32.Koopman)
	checksum = crc32.Checksum(data, tab)
	return
}

func CalculateCrc64ISO(data []byte) (checksum uint64) {
	tab := crc64.MakeTable(crc64.ISO)
	checksum = crc64.Checksum(data, tab)
	return
}

func CalculateCrc64ECMA(data []byte) (checksum uint64) {
	tab := crc64.MakeTable(crc64.ECMA)
	checksum = crc64.Checksum(data, tab)
	return
}
