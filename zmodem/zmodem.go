package zmodem

import (
	"context"
	"errors"
	"github.com/xiwh/zmodem/collectionutil"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

type Status uint8

const (
	StatusIdle    Status = iota
	StatusReceive Status = iota
	StatusSend    Status = iota
)

type ZModemConsumer struct {
	OnUpload        func() *ZModemFile
	OnUploadSkip    func(file *ZModemFile)
	OnCheckDownload func(file *ZModemFile)
	OnDownload      func(file *ZModemFile, reader io.ReadCloser) error
	//数据写入
	Writer io.Writer
	//终端回显写入
	EchoWriter io.Writer
}

type ZModem struct {
	unreadBuf        []byte
	consumer         ZModemConsumer
	waitCtx          context.Context
	waitNotifier     func()
	lock             *sync.RWMutex
	status           Status
	lastDownloadFile *ZModemFile
	lastUploadFile   *ZModemFile
	running          *atomic.Bool
	sendFileEOF      bool
}

func New(consumer ZModemConsumer) *ZModem {
	return &ZModem{
		consumer:  consumer,
		unreadBuf: make([]byte, 0, 0xff),
		status:    StatusIdle,
		lock:      new(sync.RWMutex),
		running:   new(atomic.Bool),
	}
}

func (t *ZModem) GetStatus() Status {
	return t.status
}

func (t *ZModem) sendFrame(f frame) error {
	result, err := f.marshal()
	if err != nil {
		return err
	}
	//log("发送帧")
	//log(hex.Dump(result))
	//log(f.ToString() + "\n")

	_, err = t.consumer.Writer.Write(result)
	return err
}

func (t *ZModem) sendSubPacket(packet subPacket, frameEncoding FrameEncoding, appendXON bool) error {
	result, err := packet.marshal(frameEncoding, appendXON)
	if err != nil {
		return err
	}
	//log("发送子包")
	//log(hex.Dump(result))
	_, err = t.consumer.Writer.Write(result)
	return err
}

func (t *ZModem) readFrame() (f frame, err error) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	unreadBufLen := len(t.unreadBuf)
	if unreadBufLen == 0 {
		err = errors.New("data is empty")
		return f, err
	}
	var n int
	f, n, err = unmarshalFrame(t.unreadBuf)
	if err != nil {
		return f, err
	}
	//将处理完的数据还回buf，等待下次处理
	t.unreadBuf = t.unreadBuf[n:]

	return f, err
}

func (t *ZModem) readSubPacket(frameEncoding FrameEncoding) (s subPacket, err error) {
	for i := 0; i < 6; i++ {
		//第一次不等直接尝试读取，如果失败之后最多再试5次每次最多等0.5s
		if i > 0 {
			err = t.waitWrite(time.Millisecond * 500)
			if err != nil {
				continue
			}
		}

		t.lock.RLock()
		unreadBufLen := len(t.unreadBuf)
		if unreadBufLen == 0 {
			err = errors.New("data is empty")
			t.lock.RUnlock()
			continue
		}
		t.lock.RUnlock()

		t.lock.Lock()
		var n int
		s, n, err = unmarshalSubPacket(frameEncoding, t.unreadBuf)
		if err != nil {
			t.lock.Unlock()
			continue
		}

		//将处理完的数据还回buf，等待下次处理
		t.unreadBuf = t.unreadBuf[n:]
		t.lock.Unlock()
		return s, err
	}
	return s, err
}

func (t *ZModem) Write(data []byte) (int, error) {
	//println(fmt.Sprintf("收到数据,长度:%d", len(data)))
	//println(hex.Dump(data))
	//println(hex.EncodeToString(data))

	t.lock.Lock()

	n := len(data)
	var err error
	if t.status == StatusIdle {
		if n >= 21 {
			frameData := data[n-21 : n]
			f, _, err := unmarshalFrame(frameData)
			if err == nil {
				//清零重置buf
				t.unreadBuf = t.unreadBuf[:0]
				if f.frameType == ZRQINIT {
					data = frameData
					t.status = StatusReceive
				} else if f.frameType == ZRINIT {
					data = frameData
					t.status = StatusSend
				}
			}
		}
	}
	t.lock.Unlock()

	if t.status == StatusIdle {
		//未进入zmodem协议中忽略处理直接输出回显到终端
		return t.consumer.EchoWriter.Write(data)
	}

	if collectionutil.HasSuffix(data, SES_ABORT_SEQ) {
		//收到强制结束命令，则强制结束
		t.release()
		return n, nil
	}

	for true {
		t.lock.Lock()
		if len(t.unreadBuf) >= 0xfffff {
			//buf消费能力跟不上buf满了(1MB)，等待buf有余量时再写入
			t.lock.Unlock()
			time.Sleep(25 * time.Millisecond)
			continue
		} else {
			t.unreadBuf = append(t.unreadBuf, data...)
			if t.waitCtx != nil {
				t.waitNotifier()
				t.waitNotifier = nil
				t.waitCtx = nil
			}
			t.lock.Unlock()
			break
		}
	}

	if t.status == StatusReceive {
		go t.handleReceive()
	} else if t.status == StatusSend {
		go t.handleSend()
	}

	return n, err
}

func (t *ZModem) waitWrite(timeout time.Duration) error {
	t.lock.Lock()
	waitCtx := t.waitCtx
	if waitCtx != nil {
		t.lock.Unlock()
		<-waitCtx.Done()
		//释放再加锁避免等待堵塞整个锁
		t.lock.Lock()
	}
	waitCtx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	t.waitCtx = waitCtx
	t.waitNotifier = cancelFunc
	t.lock.Unlock()
	<-waitCtx.Done()
	err := waitCtx.Err()
	if err == context.Canceled {
		err = nil
	}
	return err
}

func (t *ZModem) sendClose() {
	_, _ = t.consumer.Writer.Write(SES_ABORT_SEQ)
}

func (t *ZModem) close() {
	t.sendClose()
	if t.status != StatusIdle {
		t.release()
	}
}

func (t *ZModem) release() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.status != StatusIdle {
		t.sendFileEOF = false
		if t.lastDownloadFile != nil {
			if t.lastDownloadFile.buf != nil {
				_ = t.lastDownloadFile.buf.Close()
			}
			t.lastDownloadFile = nil
		}
		if t.lastUploadFile != nil {
			if t.lastUploadFile.buf != nil {
				_ = t.lastUploadFile.buf.Close()
			}
			t.lastUploadFile = nil
		}
		t.unreadBuf = make([]byte, 0, 0xff)
		t.status = StatusIdle
		t.running.Store(false)
	}
}

//var logFile, _ = os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
//
//func log(str string) {
//	_, err := logFile.WriteString(str + "\n")
//	if err != nil {
//		println(err)
//	}
//	//println(str)
//}
