package mrresponse

import (
	"net/http"
	"sync"
)
import "encoding/json"
import "log"
import "bytes"
import "strconv"

// import "fmt"

const (
	ERROR_UNKNOWN_ERROR             = 100
	ERROR_ACCESS_DENIED             = 101
	ERROR_INVALID_USER_ID           = 102
	ERROR_UNKNOWN_SERIES            = 103
	ERROR_LICENSED                  = 104
	ERROR_NO_UPDATE                 = 105
	ERROR_UNKNOWN_CHAPTER           = 106
	ERROR_CHAPTER_REMOVED           = 108
	ERROR_UNKNOWN_QUERY             = 109
	ERROR_INVALID_REQUEST           = 110
	ERROR_SOURCE_NOT_AVAILABLE      = 111
	ERROR_NO_SCRIPT                 = 112
	ERROR_IP_SERVICE_NOT_FOUND      = 113
	ERROR_UNKNOWN_CHARACTER         = 114
	ERROR_UNKNOWN_AUTHOR            = 115
	ERROR_UNKNOWN_GENRE             = 116
	ERROR_UNKNOWN_COLLECTION        = 117
	ERROR_UNKNOWN_FORYOU            = 118
	ERROR_UNKNOWN_FORYOU_SECTION    = 119
	ERROR_REQUIRE_LOGIN             = 120
	ERROR_WITH_MESSAGE              = 121
	ERROR_BETA_VERSION_TERMINATED   = 122
	ERROR_INVALID_BETA_VERSION_USER = 123
	ERROR_INVALID_USER              = 124
	ERROR_EMPTY_USER_DATA           = 125
	ERROR_NO_DATA_URL               = 126
	ERROR_NO_STRIPE_SOURCE_TOKEN    = 127
	ERROR_INVALID_ENROLL            = 128
	ERROR_INVALID_CANCELLATION      = 129
	ERROR_NO_SUBSCRIPTION           = 130
	ERROR_EXPIRED_SUBSCRIPTION      = 131
	ERROR_NO_STRIPE_PLAN            = 132
	ERROR_SUBSCRIPTION_EXIST        = 133
	ERROR_NO_STRIPE_CUSTOMER        = 134
	ERROR_USER_HAVE_NO_EMAIL        = 135
	ERROR_INVALID_SUBSCRIPTION_PLAN = 136
	ERROR_INVALID_BOT_DETECTED      = 137
	ERROR_STRIPE_REQUIRE_ACTION     = 138
	ERROR_SUCCESS                   = 0
)

type ResponseData struct {
	Code int    `json:"code"`
	Data string `json:"data"`
}

var errorCode = map[int]string{
	ERROR_UNKNOWN_QUERY:             "Unknown Query Version",
	ERROR_ACCESS_DENIED:             "Access denied",
	ERROR_INVALID_USER_ID:           "Invalid user id",
	ERROR_INVALID_REQUEST:           "Invalid request",
	ERROR_SOURCE_NOT_AVAILABLE:      "Source is not available",
	ERROR_UNKNOWN_SERIES:            "Unknown series",
	ERROR_UNKNOWN_CHAPTER:           "Unknown chapter",
	ERROR_UNKNOWN_CHARACTER:         "Unknown character",
	ERROR_UNKNOWN_AUTHOR:            "Unknown author",
	ERROR_UNKNOWN_COLLECTION:        "Unknown collection",
	ERROR_UNKNOWN_GENRE:             "Unknown genre",
	ERROR_UNKNOWN_ERROR:             "Unknown error",
	ERROR_UNKNOWN_FORYOU:            "Unknown foryou",
	ERROR_UNKNOWN_FORYOU_SECTION:    "Unknown foryou section",
	ERROR_LICENSED:                  "Manga is licensed",
	ERROR_CHAPTER_REMOVED:           "Chapter is removed",
	ERROR_NO_SCRIPT:                 "No script",
	ERROR_IP_SERVICE_NOT_FOUND:      "Provided IP does not belong to any range of IP in Database",
	ERROR_REQUIRE_LOGIN:             "Login required",
	ERROR_WITH_MESSAGE:              "Error message",
	ERROR_BETA_VERSION_TERMINATED:   "Version terminated",
	ERROR_INVALID_BETA_VERSION_USER: "Invalid beta version user",
	ERROR_INVALID_USER:              "Invalid user",
	ERROR_EMPTY_USER_DATA:           "Empty user data",
	ERROR_NO_DATA_URL:               "User has no data's storage url",
	ERROR_NO_STRIPE_SOURCE_TOKEN:    "Error no stripe source token",
	ERROR_INVALID_ENROLL:            "Invalid user's enrollment",
	ERROR_INVALID_CANCELLATION:      "Invalid user's cancallation",
	ERROR_EXPIRED_SUBSCRIPTION:      "Error subscription's expired",
	ERROR_NO_SUBSCRIPTION:           "Error no valid subscription",
	ERROR_NO_STRIPE_PLAN:            "No stripe user's subscription plan",
	ERROR_SUBSCRIPTION_EXIST:        "Error user's subscription already exist",
	ERROR_NO_STRIPE_CUSTOMER:        "Error user has no stripe customer",
	ERROR_USER_HAVE_NO_EMAIL:        "Error user has no email",
	ERROR_INVALID_SUBSCRIPTION_PLAN: "Invalid subscription plan",
	ERROR_INVALID_BOT_DETECTED:      "Bots Detected",
	ERROR_NO_UPDATE:                 "No update",
}
var bufferPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
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

func GetErrorResponseData(code int) *ResponseData {
	if data, ok := errorResponseCache[code]; ok {
		return data
	}
	return errorResponseCache[ERROR_UNKNOWN_ERROR]
}

func GetErrorResponseBytes(code int) []byte {
	if data, ok := errorResponses[code]; ok {
		return data
	}
	return errorResponses[ERROR_UNKNOWN_ERROR]
}

func ResponseError(code int, w http.ResponseWriter) {
	jsonStr, ok := errorResponses[code]

	if !ok {
		log.Panic("Unknown error " + strconv.Itoa(code))
	}
	ResponseJsonBytes(jsonStr, w)
}

// ResponseErrorWithMessage Response Error With Message
func ResponseErrorWithMessage(message string, w http.ResponseWriter) {
	result := &ResponseData{
		Code: ERROR_WITH_MESSAGE,
		Data: message,
	}
	jsonStr, err := json.Marshal(result)
	// log.Printf("%s", jsonStr)
	if err != nil {
		log.Panic(err)
	}
	ResponseJsonBytes(jsonStr, w)
}

func GetResponseErrorJson(code int) []byte {
	jsonStr, ok := errorResponses[code]

	if !ok {
		log.Panic("Unknown error " + strconv.Itoa(code))
	}
	return jsonStr
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
	buf.WriteString(`{"code":` + strconv.Itoa(code) +`,"data":`)
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

func ResponseResultDataStruct(data interface{}, w http.ResponseWriter) {
	jsonStr, err := json.Marshal(data)
	// log.Printf("%s", jsonStr)
	if err != nil {
		log.Panic(err)
	}

	ResponseResultByte(jsonStr, w)
}
