package main

import (
	"log"
	"net/http"
)

var pixel = []byte{71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 0, 0, 0, 255, 255, 255, 33, 249, 4, 1, 0, 0, 0, 0, 44, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 1, 68, 0, 59}

func handler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	for k, v := range r.Form {
		for _, val := range v {
			if val != "" {
				w.Header().Add(k, val)
			}
		}
	}
	w.Header().Set("content-type", "image/png")
	w.Write(pixel)
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":80", nil))
}
