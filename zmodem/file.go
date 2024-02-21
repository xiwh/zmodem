package zmodem

import (
	"errors"
	"github.com/xiwh/zmodem/byteutil"
	"golang.org/x/exp/slices"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type ZModemFile struct {
	Filename string `json:"filename"`
	Size     int    `json:"length"`
	ModTime  int    `json:"modTime"`
	FileMode int    `json:"fileMode"`
	No       int    `json:"no"`
	RemFiles int    `json:"remFiles"`
	RemSize  int    `json:"remSize"`
	isSkip   bool
	buf      io.ReadWriteCloser
}

func (t *ZModemFile) Skip() {
	t.isSkip = true
}

func (t *ZModemFile) marshal() []byte {
	data := make([]byte, 0, 256)

	data = append(data, []byte(t.Filename)...)
	data = append(data, 0)

	str := strings.Join([]string{
		strconv.Itoa(t.Size),
		strconv.FormatInt(int64(t.ModTime), 8),
		strconv.Itoa(t.FileMode),
		strconv.Itoa(t.No),
		strconv.Itoa(t.RemFiles),
		strconv.Itoa(t.RemSize),
	}, " ")
	data = append(data, []byte(str)...)
	data = append(data, 0)
	return data
}

func NewZModemFile(filename string, size int) (*ZModemFile, io.WriteCloser) {
	buf := byteutil.NewBlockReadWriter(int64(size))
	return &ZModemFile{
		Filename: filename,
		Size:     size,
		ModTime:  int(time.Now().Unix()),
		FileMode: int(os.ModePerm),
		No:       0,
		RemSize:  size,
		RemFiles: 1,
		isSkip:   false,
		buf:      buf,
	}, buf
}

func NewZModemLocalFile(path string) (*ZModemFile, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	return &ZModemFile{
		Filename: stat.Name(),
		Size:     int(stat.Size()),
		ModTime:  int(stat.ModTime().Unix()),
		FileMode: int(stat.Mode()),
		No:       0,
		RemSize:  int(stat.Size()),
		RemFiles: 1,
		isSkip:   false,
		buf:      file,
	}, nil
}

func parseZModemFile(data []byte) (f ZModemFile, err error) {
	dataLen := len(data)
	//名称以0结尾，找到这个下标
	nameEndIdx := slices.Index(data, 0)
	if nameEndIdx == -1 {
		err = errors.New("parse error1")
		return f, err
	}

	f.Filename = string(data[:nameEndIdx])

	//剩余信息以字符串空格区分
	str := string(data[nameEndIdx+1 : dataLen-1])
	strArr := strings.Split(str, " ")
	if len(strArr) < 6 {
		err = errors.New("parse error2")
		return f, err
	}

	f.Size, err = strconv.Atoi(strArr[0])
	if err != nil {
		return f, err
	}

	var temp int64
	temp, err = strconv.ParseInt(strArr[1], 8, 64)
	if err != nil {
		return f, err
	}
	f.ModTime = int(temp)
	f.FileMode, err = strconv.Atoi(strArr[2])
	if err != nil {
		return f, err
	}

	f.No, err = strconv.Atoi(strArr[3])
	if err != nil {
		return f, err
	}

	f.RemFiles, err = strconv.Atoi(strArr[4])
	if err != nil {
		return f, err
	}

	f.RemSize, err = strconv.Atoi(strArr[5])
	if err != nil {
		return f, err
	}

	return f, err
}
