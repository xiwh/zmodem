package zmodem

import (
	"github.com/xiwh/zmodem/byteutil"
)

func (t *ZModem) handleReceive() {
	if !t.running.CompareAndSwap(false, true) {
		//避免多个线程一起执行
		return
	}
	defer t.running.CompareAndSwap(true, false)
	for true {
		dataFrame, err := t.readFrame()
		if err != nil {
			//解析错误属于正常现象，因为可能一个大数据包被分成两段发过来了，需要等待第二段到位才能够正常解析
			return
		}
		//log("解析到接收帧")
		//log(dataFrame.ToString() + "\n")
		switch dataFrame.frameType {
		case ZRQINIT:
			err = t.sendFrame(newHexFrame(ZRINIT, DEFAULT_HEADER_DATA))
			if err != nil {
				return
			}
			break
		case ZFILE:
			packet, err := t.readSubPacket(dataFrame.encoding)
			if err != nil {
				//子包读取错误直接抛异常
				t.close()
				return
			}
			file, err := parseZModemFile(packet.data)
			if err != nil {
				return
			}
			t.consumer.OnCheckDownload(&file)
			if file.isSkip {
				t.lastDownloadFile = nil
				//是否跳过
				err = t.sendFrame(newHexFrame(ZSKIP, DEFAULT_HEADER_DATA))
				if err != nil {
					return
				}
				//文件全部跳过了
				t.close()
				return
			} else {
				//不跳过
				t.lastDownloadFile = &file
				err = t.sendFrame(newHexFrame(ZRPOS, DEFAULT_HEADER_DATA))
				if err != nil {
					return
				}
			}
			break
		case ZDATA:
			//文件传输中
			isEnd := false
			if t.lastDownloadFile == nil {
				t.close()
				return
			}
			buf := byteutil.NewBlockReadWriterBuf(make([]byte, 0, 0x2fff), int64(t.lastDownloadFile.Size))
			t.lastDownloadFile.buf = buf
			go func() {
				err := t.consumer.OnDownload(t.lastDownloadFile, buf)
				if err != nil {
					t.close()
					return
				}
			}()
			for !isEnd {
				//子包读取错误直接抛异常
				packet, err := t.readSubPacket(dataFrame.encoding)
				if err != nil {
					//子包读取错误直接抛异常
					t.close()
					_ = buf.Close()
					break
				}
				_, err = buf.Write(packet.data)
				if err != nil {
					_ = buf.Close()
					t.sendClose()
					continue
				}
				isEnd = packet.isEnd
			}
			if isEnd {
				_ = buf.Close()
			}
		case ZEOF:
			//文件传输完毕
			err = t.sendFrame(newHexFrame(ZRINIT, DEFAULT_HEADER_DATA))
			if err != nil {
				return
			}
			break
		case ZFIN:
			//完成
			_ = t.sendFrame(newHexFrame(ZFIN, DEFAULT_HEADER_DATA))
			t.release()
			return
		default:
			t.close()
			return
		}

	}
	return
}
