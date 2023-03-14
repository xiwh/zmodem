package zmodem

//func zsendline_init() {
//	var zsendline_tab = make()
//	for i := 0; i < 256; i++ {
//		if (i & 0140)
//			zsendline_tab[i] = 0;
//		else {
//			switch (i)
//			{
//			case ZDLE:
//			case XOFF: /* ^Q */
//			case XON: /* ^S */
//			case (XOFF | 0200):
//			case (XON | 0200):
//				zsendline_tab[i] = 1;
//				break;
//			case 020: /* ^P */
//			case 0220:
//				if (turbo_escape)
//					zsendline_tab[i] = 0;
//				else
//				zsendline_tab[i] = 1;
//				break;
//			case 015:
//			case 0215:
//				if (Zctlesc)
//					zsendline_tab[i] = 1;
//				else if (!turbo_escape)
//				zsendline_tab[i] = 2;
//				else
//				zsendline_tab[i] = 0;
//				break;
//			default:
//				if (Zctlesc)
//					zsendline_tab[i] = 1;
//				else
//				zsendline_tab[i] = 0;
//			}
//		}
//	}
//}

func escape(data []byte) []byte {
	dataLen := len(data)
	result := make([]byte, 0, dataLen*2)
	var _ byte
	var curr byte
	//数据转义，只要遇见特殊字节，在前面加一个ZDLE并且当前字节的值等于异或0x40
	for i := 0; i < dataLen; i++ {
		curr = data[i]

		/* Quick check for non control characters */
		if (curr & 0140) > 0 {
			_ = curr
			result = append(result, curr)
		} else {
			switch curr {
			case byte(ZDLE):
				curr = curr ^ 0100
				_ = curr
				result = append(result, byte(ZDLE), curr)
				break
			case 021:
				curr = curr ^ 0100
				_ = curr
				result = append(result, byte(ZDLE), curr)
				break
			case 023:
				curr = curr ^ 0100
				_ = curr
				result = append(result, byte(ZDLE), curr)
				break
			case 0221:
				curr = curr ^ 0100
				_ = curr
				result = append(result, byte(ZDLE), curr)
				break
			case 0223:
				curr = curr ^ 0100
				_ = curr
				result = append(result, byte(ZDLE), curr)
				break
			default:
				if false && (curr&0140) == 0 {
					curr = curr ^ 0100
					_ = curr
					result = append(result, byte(ZDLE), curr)
				} else {
					_ = curr
					result = append(result, curr)
				}
				//break
			}
		}
	}

	//00000050  b7 2f 5f bf 27 90 35 47  03 e9 d6 f4 e6 75 c2 76  |./_.'.5G.....u.v|
	//00000060  3b 88 f5 11 06 33 9a e1  68 ba e9 64 22 4d 67 47  |;....3..h..d"MgG|
	//00000070  3f d8 87 83 b7 ae 87 a6  ef e0 30 1a b0 3d 8c ee  |?.........0..=..|
	//00000080  30 b4 26 de 3c d8 be 19  5e e1 d1 0d fb 31 81 17  |0.&.<...^....1..|
	//00000090  eb b7 e0 86 f8 76 07 0f  7b d7 d9 47 db 36 01 20  |.....v..{..G.6. |
	//000000a0  81 66 30 f0 6c 86 bd f5  de 74 f0 3c b8 a3 ed e8  |.f0.l....t.<....|
	//000000b0  e0 b7 8d a7 87 21 90 dd  ce bd d8 fe 09 5a d7 77  |.....!.......Z.w|
	//000000c0  36 0c 8d 71 68 6f fc 8f  c9 04 3e c1 7b 49 23 b8  |6..qho....>.{I#.|
	//
	//
	//00000050  b7 2f 5f bf 27 90 35 47  03 e9 d6 f4 e6 75 c2 76  |./_.'.5G.....u.v|
	//00000060  3b 88 f5 18 51 06 33 9a  e1 68 ba e9 64 22 4d 67  |;...Q.3..h..d"Mg|
	//00000070  47 3f d8 87 83 b7 ae 87  a6 ef e0 30 1a b0 3d 8c  |G?.........0..=.|
	//00000080  ee 30 b4 26 de 3c d8 be  19 5e e1 d1 0d fb 31 81  |.0.&.<...^....1.|
	//00000090  17 eb b7 e0 86 f8 76 07  0f 7b d7 d9 47 db 36 01  |......v..{..G.6.|
	//000000a0  20 81 66 30 f0 6c 86 bd  f5 de 74 f0 3c b8 a3 ed  | .f0.l....t.<...|
	//000000b0  e8 e0 b7 8d a7 87 21 90  dd ce bd d8 fe 09 5a d7  |......!.......Z.|
	//000000c0  77 36 0c 8d 71 68 6f fc  8f c9 04 3e c1 7b 49 23  |w6..qho....>.{I#|
	//000000d0  b8 c7 8b 96 d6 75 d4 77  18 58 3d 39 f0 0d 69 0c  |.....u.w.X=9..i.|

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
