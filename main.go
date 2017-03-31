package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/nytimes/gziphandler"
	"github.com/sebest/logrusly"
)

var (
	logglyToken   = os.Getenv("LOGGLY_TOKEN")
	loggly        = logrus.New()
	httpAddr      *string
	assetPath     *string
	fetchExit     *bool
	noFetchAssets *bool
	//Since we are deciding to gzip based on the properties of the request, we can't use content-type.
	//Using file extensions...
	gzipTypes = []string{
		".html",
		".css",
		".js",
		".json",
	}
)

//Check/load static assets
func init() {
	httpAddr = flag.String("http", ":80", "listen addr for http server")
	assetPath = flag.String("assets", "assets", "Path to asset directory")
	fetchExit = flag.Bool("fetchonly", false, "Fetch demo assets and exit, i.e. no server")
	noFetchAssets = flag.Bool("nofetch", false, "Do not fetch demo assets")
	flag.Parse()
	if !*noFetchAssets {
		flist := map[string]string{
			"/15kb.png":  "https://en.wikipedia.org/wiki/File:BlueBellsOfScotland.PNG",
			"/15kb.jpg":  "https://en.wikipedia.org/wiki/File:Joseph_Dudley.jpg",
			"/100kb.jpg": "https://en.wikipedia.org/wiki/File:Procnias_tricarunculata.jpg",
			"/13kb.js":   "https://ajax.googleapis.com/ajax/libs/webfont/1.6.26/webfont.js",
			"/160kb.js":  "https://ajax.googleapis.com/ajax/libs/angularjs/1.6.1/angular.min.js",
			"/86kb.js":   "https://ajax.googleapis.com/ajax/libs/jquery/3.1.1/jquery.min.js",
			"/100kb.js":  "https://cdn.jsdelivr.net/angular.bootstrap/2.5.0/ui-bootstrap.min.js",
			"/11mb.mp4":  "https://archive.org/download/IFeelLove/IFeelLoveWeb_512kb.mp4",                                      //CC-Attribution-Noncommercial-Share Alike 3.0 United States https://archive.org/details/IFeelLove
			"/108mb.avi": "https://archive.org/download/UCONN_2014_Parade_in_Hartford_CT/UCONN_2014_Parade_in_Hartford_CT.mp4", //CC-Attribution-ShareAlike https://archive.org/details/ACMwest.orgACM_West_opening
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
	if *fetchExit {
		os.Exit(0)
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

func stampHeaders(w http.ResponseWriter, r *http.Request) {
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
}

func handler(w http.ResponseWriter, r *http.Request) {
	stampHeaders(w, r)
	http.ServeFile(w, r, *assetPath+r.URL.Path)
}

//genErr generates error
//No trailing slash
//Error codes between 400 - 599 are acceptable
func genErr(w http.ResponseWriter, r *http.Request) {
	stampHeaders(w, r)
	components := strings.Split(r.URL.Path, "/")
	codeStr := components[len(components)-1] //Get the portion after the last "/"
	code, err := strconv.Atoi(codeStr)
	if err != nil {
		http.Error(w, err.Error()+"\nPerhaps you have a trailing slash", http.StatusBadRequest)
		return
	}
	if code < 400 || code > 599 {
		http.Error(w, "Code must be between 400 - 600", http.StatusBadRequest)
		return
	}
	http.Error(w, "Error: "+codeStr, code)
}

//gzipfilterhandler is proxy to gziphandler.GzipHandler to only compress certain types
type gzipfilterhandler struct {
	gziphandler http.Handler
	original    http.Handler
}

func newgzipfilterhandler(h http.Handler) *gzipfilterhandler {
	return &gzipfilterhandler{gziphandler.GzipHandler(h), h}
}

func isGzipable(path string) bool {
	for _, typ := range gzipTypes {
		if strings.HasSuffix(path, typ) {
			return true
		}
	}
	return false
}

func (gz *gzipfilterhandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//Select based on file extension if we want to pass to gziphandler.GzipHandler or serve directly
	if isGzipable(r.URL.Path) {
		gz.gziphandler.ServeHTTP(w, r)
	} else {
		gz.original.ServeHTTP(w, r)
	}
}

func main() {
	hook := logrusly.NewLogglyHook(logglyToken, "https://logs-01.loggly.com/bulk/", logrus.InfoLevel, "mockserver", "origin")
	loggly.Hooks.Add(hook)

	http.HandleFunc("/", handler)
	http.HandleFunc("/err/", genErr)
	loggly.Infof("Starting server on %v", *httpAddr)
	log.Fatal(http.ListenAndServe(*httpAddr, WriteLog(newgzipfilterhandler(http.DefaultServeMux))))
}
