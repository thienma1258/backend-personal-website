package mrslack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"mangarockhd.com/mrlibs/mrutils"
)

var webHookHTTPClient = http.Client{
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 100 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
	Timeout: time.Duration(50 * time.Second),
}

var singleWebhookHTTPClient = http.Client{
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 100 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
	Timeout: time.Duration(50 * time.Second),
}

var webHookLock = sync.RWMutex{}
var webHookMessages = make(map[string]*WebHookData)

// Attachments slack post message attachments
type Attachments struct {
	Title string `json:"title"`
	Body  string `json:"text"`
	Color string `json:"color"`
}

// WebHookData slack post message
type WebHookData struct {
	Data []Attachments `json:"attachments"`
}

var defaultTitle string
var defaultWebhookURL string

// PostToWebhookSingleMessage push single message to slack app
func PostToWebhookSingleMessage(webhookURL, title, body, color string) {
	if len(webhookURL) <= 0 {
		return
	}
	_title := title
	if _title == "" {
		_title = defaultTitle
	}
	message := WebHookData{Data: []Attachments{Attachments{Title: _title, Body: body, Color: color}}}
	jsonBytes, err := json.Marshal(&message)
	if err != nil {
		mrutils.Log("Encoding data error=%v", err)
		return
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		mrutils.Log("NewRequest error=%v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := singleWebhookHTTPClient.Do(req)
	if err != nil {
		mrutils.Log("Send request error=%v", err)
		return
	}
	resp.Body.Close()
	return
}

//SetDefaultTitle set default title for slack message
func SetDefaultTitle(title string) {
	hostname, _ := os.Hostname()
	defaultTitle = hostname + " - " + title
}

//SetDefaultWebhookURL set default title for slack message
func SetDefaultWebhookURL(webhookURL string) {
	defaultWebhookURL = webhookURL
}

// PostToWebHook post messages to slack channel
func PostToWebHook(webHookURL, title, body, color string) {
	if len(webHookURL) <= 0 {
		return
	}

	webHookLock.RLock()
	message := webHookMessages[webHookURL]
	webHookLock.RUnlock()
	if message == nil {
		message = new(WebHookData)
		webHookLock.Lock()
		webHookMessages[webHookURL] = message
		webHookLock.Unlock()
	}

	webHookLock.Lock()
	message.Data = append(message.Data, Attachments{Title: title, Body: body, Color: color})
	webHookLock.Unlock()
}

func findLocationFromStacktrace(sign string) string {
	stackTrace := string(debug.Stack())
	start := strings.Index(stackTrace, sign)
	count := 0
	i := start + 10
	length := len(stackTrace)
	for ; i < length && count < 5; i++ {
		if stackTrace[i] == '\n' {
			count++
			if count == 3 {
				start = strings.Index(stackTrace[i:], "/src/") + i + 5
			}
		}
	}
	end := strings.Index(stackTrace[start:], " ") + start + 1
	pos := ""
	if start > 0 && end > 0 && start < length && end < length {
		pos = stackTrace[start:end]
	}
	return pos
}

// ErrorWebhookURL Post message with priority error
func ErrorWebhookURL(webHookURL, format string, v ...interface{}) {
	message := fmt.Sprintf(findLocationFromStacktrace("mrslack.ErrorWebhookURL(0x")+format, v...) + "\n" + string(debug.Stack())
	log.Println(message)
	PostToWebHook(webHookURL, defaultTitle, message, "#f5c6cb")
}

// WarningWebhookURL Post message with priority warning
func WarningWebhookURL(webHookURL, format string, v ...interface{}) {
	message := fmt.Sprintf(findLocationFromStacktrace("mrslack.WarningWebhookURL(0x")+format, v...)
	log.Println(message)
	PostToWebHook(webHookURL, defaultTitle, message, "#ffeeba")
}

// InfoWebhookURL Post message with priority info
func InfoWebhookURL(webHookURL, format string, v ...interface{}) {
	message := fmt.Sprintf(findLocationFromStacktrace("mrslack.InfoWebhookURL(0x")+format, v...)
	log.Println(message)
	PostToWebHook(webHookURL, defaultTitle, message, "#bee5eb")
}

// SuccessWebhookURL Post message with priority Success
func SuccessWebhookURL(webHookURL, format string, v ...interface{}) {
	message := fmt.Sprintf(findLocationFromStacktrace("mrslack.SuccessWebhookURL(0x")+format, v...)
	log.Println(message)
	PostToWebHook(webHookURL, defaultTitle, message, "#c3e6cb")
}

// PanicWebhookURL Post message with priority Success
func PanicWebhookURL(webHookURL, format string, v ...interface{}) {
	message := fmt.Sprintf(findLocationFromStacktrace("mrslack.PanicWebhookURL(0x")+format, v...) + "\n" + string(debug.Stack())
	log.Println(message)
	go PostToWebhookSingleMessage(webHookURL, defaultTitle, "<!channel> "+message, "#721c24")
}

// Error Post message with priority error
func Error(format string, v ...interface{}) {
	message := fmt.Sprintf(findLocationFromStacktrace("mrslack.Error(0x")+format, v...) + "\n" + string(debug.Stack())
	log.Println(message)
	PostToWebHook(defaultWebhookURL, defaultTitle, message, "#f5c6cb")
}

// Warning Post message with priority warning
func Warning(format string, v ...interface{}) {
	message := fmt.Sprintf(findLocationFromStacktrace("mrslack.Warning(0x")+format, v...)
	log.Println(message)
	PostToWebHook(defaultWebhookURL, defaultTitle, message, "#ffeeba")
}

// Info Post message with priority info
func Info(format string, v ...interface{}) {
	message := fmt.Sprintf(findLocationFromStacktrace("mrslack.Info(0x")+format, v...)
	log.Println(message)
	PostToWebHook(defaultWebhookURL, defaultTitle, message, "#bee5eb")
}

// Success Post message with priority Success
func Success(format string, v ...interface{}) {
	message := fmt.Sprintf(findLocationFromStacktrace("mrslack.Success(0x")+format, v...)
	log.Println(message)
	PostToWebHook(defaultWebhookURL, defaultTitle, message, "#c3e6cb")
}

// Panic Post message with priority Panic
func Panic(format string, v ...interface{}) {
	message := fmt.Sprintf(findLocationFromStacktrace("mrslack.Panic(0x")+format, v...) + "\n" + string(debug.Stack())
	log.Println(message)
	go PostToWebhookSingleMessage(defaultWebhookURL, defaultTitle, "<!channel> "+message, "#721c24")
}

func initSendingInterval() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		webHookLock.RLock()
		countMessage := len(webHookMessages)
		webHookLock.RUnlock()
		if countMessage > 0 {
			webHookLock.Lock()
			messages := webHookMessages
			webHookMessages = make(map[string]*WebHookData)
			webHookLock.Unlock()
			for webhookURL, message := range messages {
				jsonBytes, err := json.Marshal(&message)
				if err != nil {
					mrutils.Log("Encoding data error=%v", err)
					continue
				}

				req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonBytes))
				if err != nil {
					mrutils.Log("NewRequest error=%v", err)
					return
				}
				req.Header.Set("Content-Type", "application/json")

				resp, err := webHookHTTPClient.Do(req)
				if err != nil {
					mrutils.Log("Send request error=%v", err)
					return
				}

				resp.Body.Close()
			}
		}
	}
}

func init() {
	go initSendingInterval()
}
