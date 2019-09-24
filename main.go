package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

var connectionCount int

type app struct{}

var delay time.Duration
var addr string

func (a *app) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	connectionCount++
	i := connectionCount
	log.Printf("%d: Got connection sleeping for 100", i)
	time.Sleep(delay)
	log.Printf("%d: Finished", i)
}

func main() {

	flag.DurationVar(&delay, "delay", 100*time.Second, "Delay to wait for")
	flag.StringVar(&addr, "addr", ":8090", "Port to bind to")
	flag.Parse()

	a := &app{}
	srv := &http.Server{
		Addr:           addr,
		Handler:        a,
		ReadTimeout:    200 * time.Second,
		WriteTimeout:   200 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("Binding black hole web server to %s", addr)
	log.Fatal(srv.ListenAndServe())
}
