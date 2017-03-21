package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/nytimes/gziphandler"
	"github.com/sebest/logrusly"
)

var (
	logglyToken = os.Getenv("LOGGLY_TOKEN")
	loggly      = logrus.New()
)

//Originally stollen from https://github.com/ajays20078/go-http-logger/blob/master/httpLogger.go
//Adapted to loggly
type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	w.length = len(b)
	return w.ResponseWriter.Write(b)
}

// WriteLog Logs the Http Status for a request into fileHandler and returns a httphandler function which is a wrapper to log the requests.
func WriteLog(handle http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		start := time.Now()
		writer := statusWriter{w, 0, 0}
		handle.ServeHTTP(&writer, request)
		end := time.Now()
		latency := end.Sub(start)
		statusCode := writer.status
		length := writer.length

		msg := make(map[string]interface{})
		msg["reqHdr"] = request.Header
		msg["respHdr"] = w.Header()
		msg["start"] = start
		msg["end"] = end
		msg["host"] = request.Host
		msg["remoteaddr"] = request.RemoteAddr
		msg["path"] = request.URL.Path
		msg["rawquery"] = request.URL.RawQuery
		msg["proto"] = request.Proto
		msg["statuscode"] = statusCode
		msg["length"] = length
		msg["latency"] = latency
		loggly.WithFields(msg).Info()
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	//Static headers, which can be overridden by QS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.ParseForm()
	for k, v := range r.Form {
		for _, val := range v {
			if val != "" {
				w.Header().Add(k, val)
			}
		}
	}
	//w.Header().Set("content-type", "image/png")
	w.Header().Set("X-TB-time", time.Now().String())
	//w.Write(pixel)
	http.ServeFile(w, r, "assets"+r.URL.Path)
}

func main() {
	hook := logrusly.NewLogglyHook(logglyToken, "https://logs-01.loggly.com/bulk/", logrus.InfoLevel, "mockserver", "origin")
	loggly.Hooks.Add(hook)

	http.HandleFunc("/", handler)

	log.Fatal(http.ListenAndServe(":80", WriteLog(gziphandler.GzipHandler(http.DefaultServeMux))))
}
