package compress

import (
	"bytes"
	"compress/flate"
	"io"

	log "github.com/schollz/logger"
)

func CompressWithOption(src []byte, level int) []byte {
	compressdData := new(bytes.Buffer)
	compress(src, compressdData, level)
	return compressdData.Bytes()
}

func Compress(src []byte) []byte {
	compressdData := new(bytes.Buffer)
	compress(src, compressdData, -2)
	return compressdData.Bytes()
}

func Decompress(src []byte) []byte {
	compressdData := bytes.NewBuffer(src)
	deCompressdData := new(bytes.Buffer)
	decompress(compressdData, deCompressdData)
	return deCompressdData.Bytes()
}

func compress(src []byte, dest io.Writer, level int) {
	compressor, err := flate.NewWriter(dest, level)
	if err != nil {
		log.Debugf("error level data:%v", err)
		return
	}
	if _, err := compressor.Write(src); err != nil {
		log.Debugf("error wirte data:%v", err)
	}
	compressor.Close()
}

func decompress(src io.Reader, dest io.Writer) {
	decompressor := flate.NewReader(src)
	if _, err := io.Copy(dest, decompressor); err != nil {
		log.Debugf("error copying data:%v", err)
	}
	decompressor.Close()
}
