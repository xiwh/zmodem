package byteutil

import (
	"encoding/hex"
	"fmt"
	"io"
	"testing"
	"time"
)

func TestBlockBuffer(t *testing.T) {
	buf := NewBlockReadWriter(-1)
	go func() {
		for j := 0; j < 4; j++ {

			dd := make([]byte, 64)
			for i := 0; i < 64; i++ {
				dd[i] = byte(i + 1)
			}
			_, err := buf.Write(dd)
			if err != nil {
				fmt.Printf("W:%s", err)
			}
			_, err = buf.Write(dd)
			time.Sleep(1000000000)
		}
		buf.Close()
	}()

	var count = 0
	for true {
		dd := make([]byte, 20)
		i, err := buf.Read(dd)
		if err != nil {
			fmt.Printf("R:%s", err)
			if err != io.EOF {
				t.Fail()
			}
			return
		}
		count += i
		fmt.Printf("i:%d\n", i)
		println(hex.Dump(dd[0:i]))
	}

	if count != 3*32 {
		t.Fail()
	}
}
