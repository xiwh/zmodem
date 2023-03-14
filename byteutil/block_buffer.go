package byteutil

import (
	"bytes"
	"io"
	"sync"
	"sync/atomic"
)

type BlockBuffer struct {
	buf        *bytes.Buffer
	ch         chan int
	closed     bool
	lock       *sync.Mutex
	expectSize int64
	readedSize int64
}

func NewBlockReadWriter(expectSize int64) *BlockBuffer {
	ret := &BlockBuffer{
		bytes.NewBuffer(nil),
		make(chan int, 2),
		false,
		new(sync.Mutex),
		expectSize,
		0,
	}
	return ret
}

func NewBlockReadWriterBuf(buf []byte, expectSize int64) *BlockBuffer {
	ret := &BlockBuffer{
		bytes.NewBuffer(buf),
		make(chan int, 2),
		false,
		new(sync.Mutex),
		expectSize,
		0,
	}
	return ret
}

func (t *BlockBuffer) Write(data []byte) (nr int, err error) {
	if t.closed {
		return 0, io.EOF
	}
	defer func() {
		t.ch <- 1
	}()
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.buf.Write(data)

}

func (t *BlockBuffer) Read(b []byte) (int, error) {
	for true {
		t.lock.Lock()
		n, err := t.buf.Read(b)
		t.lock.Unlock()
		if n > 0 {
			//增加已读取字节数量
			atomic.AddInt64(&t.readedSize, int64(n))
		}
		if err == io.EOF {
			//buf被读取完了，清空channel，防止write陷入堵塞
			running := true
			for running {
				select {
				case <-t.ch:
					break
				default:
					running = false
					break
				}
			}
			if t.closed {
				//如果当前buffer也被关闭说明真的读取完毕
				if t.expectSize >= 0 {
					//如果设置了预期buffer大小，而实际并未读取到这么多字节那么返回ErrUnexpectedEOF
					if t.readedSize != t.expectSize {
						return n, io.ErrUnexpectedEOF
					}
				}
				return n, io.EOF
			} else {
				//如果当前buffer未被关闭，说明还没读取完，需要卡在这里等待写入后再读取
				<-t.ch
				continue
			}
		} else if err != nil {
			//发生异常
			return n, err
		} else {
			//读取到了数据
			return n, nil
		}
	}
	return 0, nil
}

func (t *BlockBuffer) Close() error {
	if t.closed {
		return nil
	}
	t.closed = true
	t.ch <- 0
	return nil
}
