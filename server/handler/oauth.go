package handler

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"

	"cloud.google.com/go/storage"
	"github.com/gong023/my-slack-process/server/config"
)

const (
	msgTpl = `<html> <body> {{.Msg}} </body> </html>`
)

func OauthIndex(w http.ResponseWriter, r *http.Request) {
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

func InoStart(w http.ResponseWriter, r *http.Request) {
	c := config.New()
	v := url.Values{}
	v.Add("client_id", c.InoCliID)
	v.Add("redirect_uri", c.Host+"/oauth/inoreader/callback")
	v.Add("response_type", "code")
	v.Add("scope", "read write")
	v.Add("state", "abcde")
	url := "https://www.inoreader.com/oauth2/auth?" + v.Encode()

	start(w, r, url)
}

func InoCallback(w http.ResponseWriter, r *http.Request) {
	c := config.New()
	v := url.Values{}
	v.Add("client_id", c.InoCliID)
	v.Add("client_secret", c.InoCliSec)
	v.Add("redirect_uri", c.Host+"/oauth/inoreader/callback")
	v.Add("grant_type", "authorization_code")
	u := "https://www.inoreader.com/oauth2/token"

	callback(w, r, v, u, "ino")
}

func start(w http.ResponseWriter, r *http.Request, url string) {
	r.ParseForm()
	if r.PostForm.Get("pass") == config.New().Pass {
		http.Redirect(w, r, url, http.StatusFound)
	}

	t, err := template.New("start").Parse(msgTpl)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	data := struct {
		Msg string
	}{
		Msg: r.PostForm.Get("pass"),
	}
	err = t.Execute(w, data)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
}

func callback(w http.ResponseWriter, r *http.Request, vals url.Values, tokenURL, field string) {
	vals.Add("code", r.URL.Query().Get("code"))
	fmt.Fprintf(w, vals.Encode())
	fmt.Fprintf(w, tokenURL)
	res, err := http.PostForm(tokenURL, vals)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	ctx := r.Context()
	c := config.New()
	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	bucket := client.Bucket(c.OauthBucket)
	ow := bucket.Object(field).NewWriter(ctx)
	defer ow.Close()
	if _, err := ow.Write(b); err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
}
