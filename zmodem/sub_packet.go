package zmodem

import (
	"errors"
	"github.com/xiwh/zmodem/collectionutil"
)

type subPacket struct {
	packetType SubPacketType
	data       []byte
	checksum   uint32
	isEnd      bool
}

func newSubPacket(packetType SubPacketType, data []byte) subPacket {
	crc := newCrc16()
	crc.update(data)
	crc.update([]byte{byte(packetType)})
	return subPacket{
		packetType: packetType,
		data:       data,
		checksum:   uint32(crc.getSum16()),
	}
}

func newBin32SubPacket(packetType SubPacketType, data []byte) subPacket {
	return subPacket{
		packetType: packetType,
		data:       data,
		checksum:   getCrc32(data),
	}
}

func (t subPacket) marshal(frameEncoding FrameEncoding, appendXON bool) ([]byte, error) {
	var data []byte

	if appendXON {
		data = t.data
		data = append(data, byte(ZDLE), byte(t.packetType))

		if frameEncoding == ZBIN {
			data = append(data, escape([]byte{byte(t.checksum >> 8), byte(t.checksum & 0xff)})...)
		} else if frameEncoding == ZBIN32 {
			data = append(data, escape([]byte{byte(t.checksum >> 24), byte((t.checksum >> 16) & 0xff), byte((t.checksum >> 8) & 0xff), byte(t.checksum & 0xff)})...)
		} else {
			return nil, errors.New("marshal failed")
		}

		data = append(data, byte(XON))
	} else {
		data = escape(t.data)
		data = append(data, byte(ZDLE), byte(t.packetType))

		if frameEncoding == ZBIN {
			data = append(data, escape([]byte{byte(t.checksum >> 8), byte(t.checksum & 0xff)})...)
		} else if frameEncoding == ZBIN32 {
			data = append(data, escape([]byte{byte(t.checksum >> 24), byte((t.checksum >> 16) & 0xff), byte((t.checksum >> 8) & 0xff), byte(t.checksum & 0xff)})...)
		} else {
			return nil, errors.New("marshal failed")
		}
	}

	return data, nil
}

func unmarshalSubPacket(frameEncoding FrameEncoding, data []byte) (packet subPacket, n int, err error) {
	dataLen := len(data)
	endIdx := collectionutil.IndexFunc(data, func(b byte, idx int) bool {
		//上一个为ZDLE紧跟着SubPacketType的字符(表示此子包已经到结尾了)
		if idx != 0 {
			if data[idx-1] == byte(ZDLE) {
				t := SubPacketType(b)
				return t == ZCRCE || t == ZCRCQ || t == ZCRCG || t == ZCRCW || t == ZRUB0 || t == ZRUB1
			}
		}
		return false
	})
	if endIdx == -1 {
		return packet, n, errors.New("packet format error1")
	}

	packet.packetType = SubPacketType(data[endIdx])
	n = endIdx + 1
	subData := data[:n]
	subData = unescape(subData)
	subDataLen := len(subData)
	packet.data = subData[:subDataLen-1]

	var dataChecksum uint16 = 0
	hasCrc1 := false
	var lastB byte = 0
	//dataChecksum := (uint16(data[n-2]) << 8) | uint16(data[n-1])
	for i := n; i <= n+4; i++ {
		//远古屎山兼容lszrz, crc16也需要转义，那么crc两个字节就是非定长的需要动态计算，可能是2,3,4个字节
		if i >= dataLen {
			//数组越界
			return packet, n, errors.New("packet format error2")
		}
		b := data[i]
		if b != byte(ZDLE) {
			if lastB == byte(ZDLE) && ((b^0x40)&0x60) == 0 {
				b = b ^ 0x40
				lastB = 0
			} else {
				lastB = b
			}
			if hasCrc1 {
				dataChecksum |= uint16(b)
				n = i + 1
				break
			} else {
				hasCrc1 = true
				dataChecksum = uint16(b) << 8
			}
		} else {
			lastB = b
		}
	}
	checksum := getCrc16(subData)
	packet.checksum = uint32(checksum)

	if dataChecksum != checksum {
		return packet, n, errors.New("crc valid failed")
	}

	//判断子包结尾的下一个字节是否是xon，如果是则表示当前数据帧没有后续子包,否则表示还有子包
	if n < dataLen {
		if data[n] == byte(XON) {
			packet.isEnd = true
			n += 1
		}
	}
	if !packet.isEnd {
		//ZFILE最后一个子包标记为ZCRCE
		packet.isEnd = packet.packetType == ZCRCE
	}
	return packet, n, err
}
