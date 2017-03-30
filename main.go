package main

import (
	"flag"
	"io"
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
	httpAddr    *string
	assetPath   *string
)

//Check/load static assets
func init() {
	httpAddr = flag.String("http", ":80", "listen addr for http server")
	assetPath = flag.String("assets", "assets", "listen addr for http server")
	flag.Parse()
	flist := map[string]string{
		"/15kb.png":  "https://upload.wikimedia.org/wikipedia/en/6/66/Circle_sampling.png",
		"/15kb.jpg":  "http://static.cdnplanet.com/static/rum/15kb-image.jpg",
		"/100kb.jpg": "http://static.cdnplanet.com/static/rum/100kb-image.jpg",
		"/10kb.js":   "https://rum.turbobytes.com/static/rum/rum.js",
		"/160kb.js":  "https://ajax.googleapis.com/ajax/libs/angularjs/1.5.7/angular.min.js",
		"/86kb.js":   "https://ajax.googleapis.com/ajax/libs/jquery/3.1.1/jquery.min.js",
		"/100kb.js":  "https://cdn.jsdelivr.net/angular.bootstrap/2.5.0/ui-bootstrap.min.js",
		"/10mb.mp4":  "https://tdispatch.com/wp-content/uploads/2014/11/tdispatch-10MB-MP4-.mp4?_=2",
		"/150mb.avi": "http://download.blender.org/peach/bigbuckbunny_movies/big_buck_bunny_480p_stereo.avi",
	}
	//Ensure assets directory exists
	os.Mkdir(*assetPath, os.ModePerm)
	for fname, url := range flist {
		fname = *assetPath + fname
		log.Println(fname, url)
		if _, err := os.Stat(fname); os.IsNotExist(err) {
			//Asset missing, download it.
			dl(fname, url)
		}
	}

}

func dl(fname, url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal(resp.Status)
	}
	defer resp.Body.Close()
	f, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		f.Close()
		os.Remove(f.Name())
		log.Fatal(err)
	}
	f.Close()
}

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
	w.length += len(b)
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
	http.ServeFile(w, r, *assetPath+r.URL.Path)
}

func main() {
	hook := logrusly.NewLogglyHook(logglyToken, "https://logs-01.loggly.com/bulk/", logrus.InfoLevel, "mockserver", "origin")
	loggly.Hooks.Add(hook)

	http.HandleFunc("/", handler)

	log.Fatal(http.ListenAndServe(*httpAddr, WriteLog(gziphandler.GzipHandler(http.DefaultServeMux))))
}
