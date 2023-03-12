package zmodem

func escape(data []byte) []byte {
	dataLen := len(data)
	result := make([]byte, 0, dataLen*2)
	//数据转义，只要遇见特殊字节，在前面加一个ZDLE并且当前字节的值等于异或0x40
	for i := 0; i < dataLen; i++ {
		current := data[i]
		if current&0x60 == 0 {
			result = append(result, byte(ZDLE), current^0x40)
		} else {
			result = append(result, current)
		}
	}
	return result
}

func unescape(data []byte) []byte {
	dataLen := len(data)
	result := make([]byte, 0, dataLen)
	var last byte
	//数据反转义，只要遇见ZDLE就忽略，并且下一个字节的值等于异或0x40
	for i := 0; i < dataLen; i++ {
		current := data[i]
		if current != byte(ZDLE) {
			if last == byte(ZDLE) && ((current^0x40)&0x60) == 0 {
				last = 0
				result = append(result, current^0x40)
			} else {
				result = append(result, current)
				last = current
			}
		} else {
			last = current
		}
	}
	return result
}
