package myioutil

import (
	"bytes"
	"io"
)

type writeFuncStruct struct {
	f func(p []byte) (n int, err error)
}

func (w *writeFuncStruct) Write(p []byte) (n int, err error) {
	return w.f(p)
}

// CopyFixedSize 读取满固定大小的字节或读取遇到错误再进行写入
// 推荐使用在网络传输层，因为在互联网环境中频繁发送小数据包相对一次性发送大数据包效率能提高几倍
func CopyFixedSize(writer io.Writer, src io.Reader, size int64) (int64, error) {
	var total int64 = 0
	buf := bytes.NewBuffer(make([]byte, 0, size))
	for true {
		n, err := io.CopyN(buf, src, size)
		if err != nil {
			if n != 0 {
				n, _ = io.CopyN(writer, buf, n)
				total += n
			}
			return total, err
		} else {
			n, err = io.CopyN(writer, buf, n)
			total += n
			if err != nil {
				return total, err
			}
		}
		buf.Reset()
	}
	return 0, io.EOF
}

func WriteFunc(f func(p []byte) (n int, err error)) io.Writer {
	return &writeFuncStruct{
		f,
	}
}
