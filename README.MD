**Zmodem protocol implementation, refer to: http://wiki.synchro.net/ref:zmodem , compatible with lszrz**

## install
    go get github.com/xiwh/zmodem
## Example
```golang
package main

import (
	"fmt"
	"github.com/xiwh/zmodem/byteutil"
	"github.com/xiwh/zmodem/zmodem"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("wsl")
	} else {
		cmd = exec.Command("bash")
	}
	cmdWriter := byteutil.NewBlockReadWriter(-1)
	zmodemInstance := zmodem.New(zmodem.ZModemConsumer{
		OnUploadSkip: func(file *zmodem.ZModemFile) {
			println(fmt.Sprintf("send file:%s skip", file.Filename))
		},
		OnUpload: func() *zmodem.ZModemFile {
			uploadFile, _ := zmodem.NewZModemLocalFile("~/.ssh/id_rsa.pub")
			return uploadFile
		},
		OnCheckDownload: func(file *zmodem.ZModemFile) {
			if pathExists(filepath.Join("test_dir", file.Filename)) {
				println(fmt.Sprintf("receive file:%s skip", file.Filename))
				file.Skip()
			}
		},
		OnDownload: func(file *zmodem.ZModemFile, reader io.ReadCloser) error {
			_ = os.Mkdir("test_dir", os.ModePerm)
			f, _ := os.OpenFile(filepath.Join("test_dir", file.Filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
			_, err := io.Copy(f, reader)
			if err == nil {
				println(fmt.Sprintf("receive file:%s sucesss", file.Filename))
			} else {
				println(fmt.Sprintf("receive file:%s failed:%s", file.Filename, err.Error()))
			}
			return err
		},
		Writer:     cmdWriter,
		EchoWriter: os.Stdout,
	})
	cmd.Stdout = zmodemInstance
	cmd.Stderr = os.Stderr
	cmd.Stdin = cmdWriter
	err := cmd.Start()
	if err != nil {
		println(err.Error())
		return
	} else {
		cmdWriter.Write([]byte("cd ~/.ssh\n"))
		cmdWriter.Write([]byte("ls\n"))
		cmdWriter.Write([]byte("sz ~/.ssh/id_rsa.pub\n"))
		//cmdWriter.Write([]byte("rz\n"))
	}
	_ = cmd.Wait()
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

```
