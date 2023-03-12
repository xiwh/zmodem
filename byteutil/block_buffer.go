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
	readSize   int64
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
		if t.closed {
			var n = 0
			var err = io.EOF
			if t.buf.Len() != 0 {
				//虽然已经关闭，但是可能还有未读取完的buf，需要读取出来
				n, err = t.buf.Read(b)
				atomic.AddInt64(&t.readSize, int64(n))
			}
			if t.buf.Len() == 0 && t.expectSize >= 0 && t.readSize < t.expectSize {
				//如果设置了预期大小，而读取完毕时并没有读取到这么字节，将返回ErrUnexpectedEOF
				err = io.ErrUnexpectedEOF
			}
			t.lock.Unlock()
			return n, err
		}
		i, err := t.buf.Read(b)
		if err == nil {
			atomic.AddInt64(&t.readSize, int64(i))
			if t.buf.Len() == 0 {
				running := true
				//读取完buf，清空channel,防止堵塞写入
				for running {
					select {
					case <-t.ch:
						break
					default:
						running = false
						break
					}
				}
			} else {
				select {
				case <-t.ch:
					break
				}
			}

			t.lock.Unlock()
			return i, err
		} else if err != io.EOF {
			t.lock.Unlock()
			return 0, err
		}
		t.lock.Unlock()

		signal := <-t.ch
		if signal == 0 {
			//通过关闭信号判断关闭，不能直接在close方法直接关闭，否则如果有线程在read将直接panic
			close(t.ch)
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
