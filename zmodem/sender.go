package zmodem

import (
	"encoding/hex"
	"fmt"
	"io"
)

func (t *ZModem) handleSend() {
	if !t.running.CompareAndSwap(false, true) {
		//避免多个线程一起执行
		return
	}
	defer t.running.CompareAndSwap(true, false)
	for true {
		dataFrame, err := t.readFrame()
		if err != nil {
			//解析错误属于正常现象，因为可能一个大数据包被分成两段发过来了，需要等待第二段到位才能够正常解析
			log(fmt.Sprintf("err:%s,data:%s", err.Error(), hex.Dump(t.unreadBuf)))
			return
		}
		log("解析到接收帧")
		log(dataFrame.ToString() + "\n")
		switch dataFrame.frameType {
		case ZRINIT:
			println("ZRINIT")
			if t.lastUploadFile != nil {
				//传输完成
				err = t.sendFrame(newHexFrame(ZFIN, DEFAULT_HEADER_DATA))
				println("ffff")
				break
			}
			zmodemFile := t.consumer.OnUpload()
			if zmodemFile == nil {
				t.close()
				return
			}
			err = t.sendFrame(newBinFrame(ZFILE, DEFAULT_HEADER_DATA))
			if err != nil {
				return
			}
			t.lastUploadFile = zmodemFile
			err = t.sendSubPacket(newSubPacket(ZCRCW, zmodemFile.marshal()), ZBIN, true)
			if zmodemFile == nil {
				t.close()
				return
			}
			break
		case ZRPOS:
			if t.lastUploadFile == nil {
				t.close()
				return
			}
			//发送文件内容
			err = t.sendFrame(newBinFrame(ZDATA, DEFAULT_HEADER_DATA))
			if err != nil {
				t.close()
				return
			}
			//8k一个包发送
			buf := make([]byte, 1024)
			for true {
				n, err := io.ReadFull(t.lastUploadFile.buf, buf)
				if err != nil {
					if err == io.EOF || err == io.ErrUnexpectedEOF {
						//正常读取完毕
						err = t.sendSubPacket(newSubPacket(ZCRCE, buf[:n]), ZBIN, false)
						if err != nil {
							t.close()
							return
						}
						err = t.sendFrame(newBinFrame(ZEOF, []byte{0x7d, 0x21, 0, 0}))
						if err != nil {
							t.close()
							return
						}
						break
					} else {
						//读取出错
						t.close()
						return
					}
				} else {
					err = t.sendSubPacket(newSubPacket(ZCRCG, buf[:n]), ZBIN, false)
					if err != nil {
						t.close()
						return
					}
				}
			}
		case ZSKIP:
			//跳过
			if t.lastUploadFile == nil {
				t.close()
				return
			}
			t.consumer.OnUploadSkip(t.lastUploadFile)
			err = t.sendFrame(newHexFrame(ZFIN, DEFAULT_HEADER_DATA))
			if t.lastUploadFile == nil {
				t.close()
				return
			}
			t.release()
		case ZFIN:
			//完成
			_, _ = t.consumer.Writer.Write([]byte{'O', 'O'})
			t.status = StatusIdle
			t.release()
			return
		default:
			t.close()
			return
		}

	}
	return
}
