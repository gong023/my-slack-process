package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gong023/my-slack-process/server/config"
	"github.com/gong023/my-slack-process/server/google"
	"github.com/gong023/my-slack-process/server/inoreader"
	"github.com/gong023/my-slack-process/server/pixiv"
	"golang.org/x/crypto/acme/autocert"
)

func oauthIndex(w http.ResponseWriter, r *http.Request) {
	html := `
        <html>
        <body>
        <form action="/oauth/inoreader/start" method="post">
			inoreader:
        	<input type="password" name="pass" />
        	<input type="submit" />
        </form>
        <form action="/oauth/google/start" method="post">
			google:
        	<input type="password" name="pass" />
        	<input type="submit" />
        </form>
        </body>
        </html>
        `
	fmt.Fprintf(w, html)
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "it works")
}

func main() {
	c := config.New()
	http.HandleFunc("/", index)
	http.HandleFunc("/oauth", oauthIndex)
	http.HandleFunc("/oauth/inoreader/start", inoreader.Start(c))
	http.HandleFunc("/oauth/inoreader/callback", inoreader.Callback(c))
	http.HandleFunc("/oauth/google/start", google.Start(c))
	http.HandleFunc("/oauth/google/callback", google.Callback(c))
	http.HandleFunc("/pi", pixiv.ImageProxy)

	srv := &http.Server{
		Addr:         "1443",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	if c.Prod == "1" {
		u, err := url.Parse(c.Host)
		if err != nil {
			log.Fatal(err)
		}
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(u.Host),
			Cache:      autocert.DirCache(c.Cert),
		}

		srv.TLSConfig = &tls.Config{
			GetCertificate: certManager.GetCertificate,
		}
	}

	closed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("srv shutdown:%v", err)
		}
		close(closed)
	}()

	if c.Prod == "1" {
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}
	<-closed
}
