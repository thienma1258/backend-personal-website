package mrutils

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"log"
	"sync"
)

//Smart compress
const minLenToCompress = 1024
const minCompressionReductionToAccept = 128

var writterZippers = sync.Pool{New: func() interface{} {
	gz, err := gzip.NewWriterLevel(nil, 5)
	if err != nil {
		log.Panicln(err)
	}
	return gz
}}

var maximumCompressionZippers = sync.Pool{New: func() interface{} {
	gz, err := gzip.NewWriterLevel(nil, 9)
	if err != nil {
		log.Panicln(err)
	}
	return gz
}}

var readerZippers = sync.Pool{New: func() interface{} {
	return new(gzip.Reader)
}}

//CompressBytes Smart compress
func CompressBytes(value []byte) []byte {
	if len(value) < minLenToCompress {
		return value
	}

	gz := writterZippers.Get().(*gzip.Writer)
	defer writterZippers.Put(gz)

	buf := new(bytes.Buffer)
	buf.Reset()
	// defer gzipBufPool.Put(buf)

	gz.Reset(buf)

	gz.Write(value)
	gz.Close()
	results := buf.Bytes()

	//No meaning to compress it, returns original stream
	if len(value)-len(results) < minCompressionReductionToAccept {
		return value
	}

	// log.Printf("Compress from %d to %d", len(value), len(results))
	return results
}

//Full compress
func ForceCompressBytes(value []byte, forceMaximum bool) []byte {
	var gz *gzip.Writer

	if forceMaximum || len(value) < 100000 {
		gz = maximumCompressionZippers.Get().(*gzip.Writer)
		defer maximumCompressionZippers.Put(gz)
	} else {
		gz = writterZippers.Get().(*gzip.Writer)
		defer writterZippers.Put(gz)
	}

	buf := new(bytes.Buffer)
	buf.Reset()

	gz.Reset(buf)

	gz.Write(value)
	gz.Close()

	return buf.Bytes()
}

func DecompressToString(value []byte) []byte {
	if len(value) < 20 {
		return value
	}

	//Check gzip header
	if value[0] != 0x1f || value[1] != 0x8b {
		return value
	}

	b := bytes.NewReader(value)

	gz := readerZippers.Get().(*gzip.Reader)
	defer readerZippers.Put(gz)

	gz.Reset(b)

	s, err := ioutil.ReadAll(gz)
	if err != nil {
		log.Printf("Cannot decompress string %s %s", value, err.Error())
		return []byte{}
	}

	defer gz.Close()
	return s
}
