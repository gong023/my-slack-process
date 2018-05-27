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
	mux.HandleFunc("/oauth", handler.OauthIndex)
	mux.HandleFunc("/oauth/inoreader/start", handler.InoStart)
	mux.HandleFunc("/oauth/inoreader/callback", handler.InoCallback)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
