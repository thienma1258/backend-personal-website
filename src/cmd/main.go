package main

import (
	"dongpham/src/config"
	"dongpham/src/rest"
	"dongpham/src/utils"
	"dongpham/version"
	"flag"
	"fmt"
	"github.com/gorilla/mux"

	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"
)

func middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("%s: %s", err, debug.Stack())
				log.Printf("Recovered in service %+v", err)
				utils.ResponseError(utils.ERROR_UNKNOWN_ERROR, w)
			}
		}()

		startTime := time.Now().UnixNano()
		h.ServeHTTP(w, r)
		if config.Verbose {
			code := r.Header.Get("CF-Ipcountry")
			if len(code) > 0 {
				log.Printf("%s %s \t%dms \t%sb \t: %s\t--> %s", code, utils.GetRemoteIp(r), (time.Now().UnixNano()-startTime)/1000000, w.Header().Get("Expected-Size"), r.Method, "https://"+r.Host+r.URL.Path+"?"+r.URL.RawQuery)
			} else {
				log.Printf("%s \t%dms \t%sb \t: %s\t--> %s", utils.GetRemoteIp(r), (time.Now().UnixNano()-startTime)/1000000, w.Header().Get("Expected-Size"), r.Method, "https://"+r.Host+r.URL.Path+"?"+r.URL.RawQuery)
			}
		}
	})
}

func initHTTPServer(httpPort int) {

	router := mux.NewRouter().StrictSlash(true)
	rest.RegisterRoutes(router)
	router.HandleFunc("/ping", ping)

	srv := &http.Server{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		Addr:         ":" + strconv.Itoa(httpPort),
		Handler:      middleware(router),
	}

	hostname, _ := os.Hostname()
	log.Printf("%s : (%s) Starting HTTP server at %d. isMaster=%v, verbose=%v", hostname, version.Version, httpPort, config.IsMaster, config.Verbose)
	log.Printf("%s : (%s) Starting HTTP server at %d. isMaster=%v, verbose=%v",
		hostname, version.Version, httpPort, config.IsMaster, config.Verbose)
	log.Fatal(srv.ListenAndServe())
}

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func SetupCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		os.Exit(0)
	}()
}

// ping use to test if the server is still alive
func ping(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("PONG"))
	if err != nil {
		log.Printf("Ping %v", err)
	}

}

func main() {
	httpPort := flag.Int("port", config.HTTPPort, "which http port that server will be listening")
	// init randome seed
	rand.Seed(time.Now().UTC().UnixNano())
	requireLoop := false
	if *httpPort > 0 {
		requireLoop = true
		go initHTTPServer(*httpPort)
	}

	if requireLoop {
		SetupCloseHandler()
		select {}
	}

}
