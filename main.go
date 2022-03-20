package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

var connectionCount int

type app struct{}

var delay time.Duration
var addr string
var proxyAddr string

func (a *app) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	connectionCount++
	i := connectionCount
	log.Printf("%d: Got connection sleeping for %v", i, delay)
	time.Sleep(delay)

	if a.proxyCallToUpstream(w, req) {
		return
	}

	log.Printf("%d: Finished", i)
}

func (a *app) proxyCallToUpstream(w http.ResponseWriter, req *http.Request) bool {
	client := newHTTPClient()
	parse, _ := url.Parse(proxyAddr)
	request := http.Request{
		Method: req.Method,
		URL:    parse,
		Body:   req.Body,
	}
	upstreamResponse, err := client.Do(&request)
	if err != nil {
		log.Printf("error doing request: %v", err)
		w.WriteHeader(500)
		return true
	}
	w.WriteHeader(upstreamResponse.StatusCode)
	body, err := ioutil.ReadAll(upstreamResponse.Body)
	if err != nil {
		log.Printf("failed to read body from upstream: %v", err)
		w.WriteHeader(500)
		return true
	}
	if _, err := w.Write(body); err != nil {
		log.Printf("failed to write body to client: %v", err)
	}
	return false
}

func newHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   120 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   120 * time.Second,
			ResponseHeaderTimeout: 120 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: 3 * time.Minute,
	}
	return client
}

func main() {

	flag.DurationVar(&delay, "delay", 100*time.Second, "Delay to wait for")
	flag.StringVar(&addr, "addr", ":8090", "Port to bind to")
	flag.StringVar(&proxyAddr, "proxy", "", "Where to send requests to")
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
