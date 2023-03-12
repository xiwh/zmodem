package zmodem

type ZModemChar byte

type FrameEncoding ZModemChar
type FrameType ZModemChar
type SubPacketType ZModemChar

var SES_ABORT_SEQ = []byte{24, 24, 24, 24, 24, 24, 24, 24, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 0}
var DEFAULT_HEADER_DATA = []byte{0, 0, 0, 0}

const (
	ZPAD ZModemChar = '*'
	ZDLE ZModemChar = 0x18

	ZBIN   FrameEncoding = 'A'
	ZHEX   FrameEncoding = 'B'
	ZBIN32 FrameEncoding = 'C'

	ZCRCE SubPacketType = 'h'
	ZCRCG SubPacketType = 'i'
	ZCRCQ SubPacketType = 'j'
	ZCRCW SubPacketType = 'k'
	ZRUB0 SubPacketType = 'l'
	ZRUB1 SubPacketType = 'm'

	ZRQINIT    FrameType = 0
	ZRINIT     FrameType = 1
	ZSINIT     FrameType = 2
	ZACK       FrameType = 3
	ZFILE      FrameType = 4
	ZSKIP      FrameType = 5
	ZNAK       FrameType = 6
	ZABORT     FrameType = 7
	ZFIN       FrameType = 8
	ZRPOS      FrameType = 9
	ZDATA      FrameType = 10
	ZEOF       FrameType = 11
	ZFERR      FrameType = 12
	ZCRC       FrameType = 13
	ZCHALLENGE FrameType = 14
	ZCOMPL     FrameType = 15
	ZCAN       FrameType = 16
	ZFREECNT   FrameType = 17
	ZCOMMAND   FrameType = 18
	ZSTDERR    FrameType = 19

	SOH     ZModemChar = 0x01
	STX     ZModemChar = 0x02
	EOT     ZModemChar = 0x04
	ENQ     ZModemChar = 0x05
	ACK     ZModemChar = 0x06
	LF      ZModemChar = 0x0a
	LFXOR80 ZModemChar = 0x8a
	CR      ZModemChar = 0x0d
	XON     ZModemChar = 0x11
	XOFF    ZModemChar = 0x13
	NAK     ZModemChar = 0x15
	CAN     ZModemChar = 0x18
)
