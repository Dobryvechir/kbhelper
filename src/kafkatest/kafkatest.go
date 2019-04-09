package main

import (
      "fmt" 
      "net/http"
      "log"
      "time" 
)

func generalHandler(w http.ResponseWriter, r *http.Request) {
       log.Printf("%s %s",r.Method,r.URL.Path)
       fmt.Fprintf(w, "%s", r.URL.Path)
}

func main() {
     http.HandleFunc("/", generalHandler)
     s := &http.Server{
	Addr:           ":9092",
	ReadTimeout:    10 * time.Second,
	WriteTimeout:   10 * time.Second,
	MaxHeaderBytes: 1 << 20,
     }
     log.Fatal(s.ListenAndServe())
}