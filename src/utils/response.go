package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/json"

	"log"
	"net/http"
	"strconv"
	"sync"
)

const (
	ERROR_SUCCESS         = 0
	ERROR_UNKNOWN_ERROR   = 500
	ERROR_INVALID_REQUEST = 400
)

var errorCode = map[int]string{
	ERROR_UNKNOWN_ERROR: "Unknown error",
}

var errorResponses map[int]([]byte)
var errorResponseCache map[int]*ResponseData

func init() {
	errorResponses = make(map[int]([]byte))
	errorResponseCache = make(map[int]*ResponseData)
	for code, msg := range errorCode {
		res := &ResponseData{
			Code: code,
			Data: msg,
		}

		jsonStr, err := json.Marshal(res)
		if err != nil {
			log.Panic(err)
		}
		errorResponses[code] = jsonStr
		errorResponseCache[code] = res
		// log.Printf("Code %d json %s", code, jsonStr)
	}
}

var bufferPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

var compressionZippers = sync.Pool{New: func() interface{} {
	gz, err := gzip.NewWriterLevel(nil, 6)
	if err != nil {
		log.Panicln(err)
	}
	return gz
}}

func ResponseError(code int, w http.ResponseWriter) {
	jsonStr, ok := errorResponses[code]

	if !ok {
		log.Panic("Unknown error " + strconv.Itoa(code))
	}
	ResponseJsonBytes(jsonStr, w)
}

func compressBytes(value []byte) []byte {
	gz := compressionZippers.Get().(*gzip.Writer)
	defer compressionZippers.Put(gz)
	buf := new(bytes.Buffer)
	buf.Reset()

	gz.Reset(buf)

	gz.Write(value)
	gz.Close()

	return buf.Bytes()
}

type ResponseData struct {
	Code int    `json:"code"`
	Data string `json:"data"`
}

func ResponseResultJsonByteWithGzip(data []byte, w http.ResponseWriter) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	buf.WriteString(`{"code":0,"data":`)
	buf.Write(data)
	buf.WriteByte('}')
	body := buf.Bytes()
	blob := compressBytes(body)
	ResponseGzippedJsonWithLength(blob, w, len(body))
}

func ResponseResultJsonByte(data []byte, w http.ResponseWriter) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	buf.WriteString(`{"code":0,"data":`)
	buf.Write(data)
	buf.WriteByte('}')
	ResponseJsonBytes(buf.Bytes(), w)
}

func GetResponseResultJsonByte(data []byte) []byte {
	buf := new(bytes.Buffer)
	buf.Grow(len(data) + 20) //Trick: avoid unnessary allocation operations.
	buf.WriteString(`{"code":0,"data":`)
	buf.Write(data)
	buf.WriteByte('}')
	return buf.Bytes()
}

func ResponseWithCodeAndData(code int, data []byte, w http.ResponseWriter) {
	buf := new(bytes.Buffer)
	buf.WriteString(`{"code":` + strconv.Itoa(code) + `,"data":`)
	buf.Write(data)
	buf.WriteByte('}')
	ResponseJsonBytes(buf.Bytes(), w)
}

func ResponseResultString(data string, w http.ResponseWriter) {
	result := &ResponseData{
		Code: ERROR_SUCCESS,
		Data: data,
	}
	jsonStr, err := json.Marshal(result)
	if err != nil {
		log.Panic(err)
	}
	ResponseJsonBytes(jsonStr, w)
}

func DirtyTrickToGenerateJsonObject(data map[string][]byte) []byte {
	estimatedLength := 2
	for key, value := range data {
		estimatedLength += len(value) + 5 + len(key)
	}

	buf := new(bytes.Buffer)
	buf.Grow(estimatedLength)

	buf.WriteByte('{')
	first := true
	for key, value := range data {
		if !first {
			buf.WriteByte(',')
		}
		first = false
		buf.WriteByte('"')
		buf.WriteString(key)
		buf.WriteString(`":`)
		buf.Write(value)
	}
	buf.WriteByte('}')
	return buf.Bytes()
}

func ResponseJsonString(data string, w http.ResponseWriter) {
	ResponseJsonBytes([]byte(data), w)
}

func ResponseJsonBytes(data []byte, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Expected-Size", strconv.Itoa(len(data)))
	// w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

func ResponseGzippedJson(data []byte, w http.ResponseWriter) {
	ResponseGzippedJsonWithLength(data, w, len(data))
}

func ResponseGzippedJsonWithLength(data []byte, w http.ResponseWriter, length int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Expected-Size", strconv.Itoa(length))
	// w.Header().Set("Content-Length", strconv.Itoa(length)) // Do not enable it, will crash client
	w.Write(data)
}

func ResponseResultByte(data []byte, w http.ResponseWriter) {
	result := &ResponseData{
		Code: ERROR_SUCCESS,
		Data: string(data),
	}
	jsonStr, err := json.Marshal(result)
	if err != nil {
		log.Panic(err)
	}
	ResponseJsonBytes(jsonStr, w)
}

func ResponseResultDataStruct(data interface{}, w http.ResponseWriter) {
	jsonStr, err := json.Marshal(data)
	// log.Printf("%s", jsonStr)
	if err != nil {
		log.Panic(err)
	}

	ResponseResultByte(jsonStr, w)
}
