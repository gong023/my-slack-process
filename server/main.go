package main

import (
	"log"
	"net/http"

	"github.com/gong023/my-slack-process/server/handler"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handler.Index)
	mux.HandleFunc("/pi", handler.ImageProxy)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
