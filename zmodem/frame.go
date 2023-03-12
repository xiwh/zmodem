package zmodem

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
)

type frame struct {
	encoding       FrameEncoding
	frameType      FrameType
	headerData     []byte
	headerChecksum uint32
	hasSubPacket   bool
}

func (t frame) marshal() ([]byte, error) {
	if t.encoding == ZHEX {
		//需要封装成16进制的数据内容 frameType + 实际数据 + crc16
		rawDataBytes := append(append([]byte{byte(t.frameType)}, t.headerData...), byte((t.headerChecksum>>8)&0xff), byte(t.headerChecksum&0xff))
		//转换为16进制字符串
		hexStrBytes := []byte(hex.EncodeToString(rawDataBytes))
		hexStrByteLen := len(hexStrBytes)
		//数据长度为固定3个数据头字节+数据长度+3个结尾符
		result := make([]byte, 0, 3+hexStrByteLen+4)
		result = append(result, byte(ZPAD), byte(ZPAD), byte(ZDLE), byte(ZHEX))

		for i := 0; i < hexStrByteLen; i++ {
			result = append(result, hexStrBytes[i])
		}

		result = append(result,
			byte(CR),
			byte(LFXOR80),
			byte(XON),
		)

		return result, nil
	} else if t.encoding == ZBIN {
		data := make([]byte, 0, 64)
		data = append(data, '*', byte(ZDLE), byte(ZBIN), byte(t.frameType))
		data = append(data, t.headerData...)
		data = append(data, byte(t.headerChecksum>>8), byte(t.headerChecksum&0xff))
		return data, nil
	}

	return nil, errors.New("marshal failed")
}

func (t frame) ToString() string {
	if t.encoding == ZHEX {
		return fmt.Sprintf("hex ptype:%x,hdata:%s,hcrc:%d", t.frameType, hex.EncodeToString(t.headerData), t.headerChecksum)
	} else if t.encoding == ZBIN {
		return fmt.Sprintf("bin ptype:%d,hdata:%q,hcrc:%d",
			t.frameType,
			hex.EncodeToString(t.headerData),
			t.headerChecksum,
		)
	} else {
		return fmt.Sprintf("bin32 ptype:%x,hdata:%q,hcrc:%d",
			t.frameType,
			hex.EncodeToString(t.headerData),
			t.headerChecksum,
		)
	}
}

func readBufByte(buf *bytes.Buffer) (b byte) {
	b, _ = buf.ReadByte()
	return b
}

func newBinFrame(frameType FrameType, info []byte) (frameData frame) {
	frameData.frameType = frameType
	frameData.headerData = info
	frameData.encoding = ZBIN
	frameData.headerChecksum = uint32(getCrc16(append([]byte{byte(frameType)}, info...)))
	return frameData
}

func newHexFrame(frameType FrameType, info []byte) (frameData frame) {
	frameData.frameType = frameType
	frameData.headerData = info
	frameData.encoding = ZHEX
	frameData.headerChecksum = uint32(getCrc16(append([]byte{byte(frameType)}, info...)))
	return frameData
}

func unmarshalFrame(data []byte) (dataFrame frame, n int, err error) {
	dataLen := len(data)
	if dataLen < 6 {
		return dataFrame, n, errors.New("packet format error1")
	}
	firstCr := false
	if data[0] == byte(CR) {
		//兼容lszrz某些时候包头多了一个cr,需要忽略
		firstCr = true
		data = data[1:]
		dataLen -= 1
	}
	firstChar := ZModemChar(data[0])
	secondChar := ZModemChar(data[1])
	thirdChar := data[2]
	fourthChar := data[3]

	if firstChar == ZPAD && secondChar == ZPAD && ZModemChar(thirdChar) == ZDLE && FrameEncoding(fourthChar) == ZHEX {
		//HEX帧
		dataFrame.encoding = ZHEX
		//查找HEX帧结尾，可能是CR,LF也可能是CR,LF,XON
		endIdx := slices.Index(data, byte(CR))
		if endIdx == -1 && endIdx < dataLen-1 {
			return dataFrame, n, errors.New("packet format error2")
		}
		//可能是LF也可能是LF异或0x80
		hasLF := data[endIdx+1] == byte(LF) || data[endIdx+1] == byte(LFXOR80)
		if endIdx < dataLen-2 && hasLF && data[endIdx+2] == byte(XON) {
			n = endIdx + 3
		} else if endIdx < dataLen-1 && hasLF {
			n = endIdx + 2
		} else {
			return dataFrame, n, errors.New("packet format error3")
		}
		hexStr := string(data[4:endIdx])
		//解析16进制字符串转字节数组
		hexData, err := hex.DecodeString(hexStr)
		if err != nil {
			return dataFrame, n, err
		}
		hexDataLen := len(hexData)
		if hexDataLen < 2 {
			return dataFrame, n, errors.New("packet format error4")
		}
		//末尾两个字节为crc16值，对比是否正确
		headerChecksum := (uint16(hexData[hexDataLen-2]) << 8) | uint16(hexData[hexDataLen-1])
		//去除末尾两个crc字节，再计算crc
		hexData = hexData[:hexDataLen-2]
		checksum := getCrc16(hexData)
		//效验crc
		if headerChecksum != checksum {
			return dataFrame, n, errors.New("crc valid failed")
		}
		dataFrame.frameType = FrameType(hexData[0])
		dataFrame.headerData = hexData[1:]
		dataFrame.headerChecksum = uint32(checksum)
	} else if firstChar == ZPAD {
		if secondChar != ZDLE {
			return dataFrame, n, errors.New("packet format error5")
		}
		dataFrame.encoding = FrameEncoding(thirdChar)
		dataFrame.frameType = FrameType(fourthChar)

		if dataFrame.encoding == ZBIN {
			//16位crc二进制帧
			if dataLen < 10 {
				//至少10字节
				return dataFrame, n, errors.New("packet format error6")
			} else {
				n = 10
				dataFrame.headerData = data[4:8]
				headerDataChecksum := (uint16(data[8]) << 8) | uint16(data[9])
				headerChecksum := getCrc16(data[3:8])
				//效验crc
				if headerDataChecksum != headerChecksum {
					return dataFrame, n, errors.New("crc valid failed")
				}
				if dataLen >= 11 {
					dataFrame.hasSubPacket = data[10] != byte(XON)
					if !dataFrame.hasSubPacket {
						n = 11
					}
				}
			}
		} else if dataFrame.encoding == ZBIN32 {
			return dataFrame, n, errors.New("packet format error6")
			//32位crc二进制帧
		} else {
			return dataFrame, n, errors.New("packet format error6")
		}
	} else {
		return dataFrame, n, errors.New("packet format error7")
	}
	if firstCr {
		n += 1
	}
	return dataFrame, n, err
}
