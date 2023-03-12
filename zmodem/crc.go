package zmodem

import (
	"github.com/sigurn/crc16"
	"hash"
	"hash/crc32"
)

var crcTable = crc16.MakeTable(crc16.CRC16_XMODEM)

type crcMix struct {
	hash16 crc16.Hash16
	hash32 hash.Hash32
}

func (t crcMix) update(data []byte) {
	if t.hash16 != nil {
		_, _ = t.hash16.Write(data)
	}
	if t.hash32 != nil {
		_, _ = t.hash32.Write(data)
	}
}

func (t crcMix) getSum16() uint16 {
	return t.hash16.Sum16()
}

func (t crcMix) getSum32() uint32 {
	return t.hash32.Sum32()
}

func newCrc16() crcMix {
	return crcMix{
		hash16: crc16.New(crcTable),
	}
}

func newCrc32() crcMix {
	return crcMix{
		hash32: crc32.NewIEEE(),
	}
}

func getCrc16(data []byte) uint16 {
	return crc16.Checksum(data, crcTable)
}

func getCrc32(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}
