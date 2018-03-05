package main

import (
	"log"
	"net/http"
	"net/http/httputil"
)

func main() {
	rp := &httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL.Scheme = "https"
			request.URL.Host = ":443"
		},
	}
	srv := http.Server{
		Addr:    ":80",
		Handler: rp,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
