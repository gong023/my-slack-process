package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type (
	RefreshReq struct {
		URL string
	}

	TokenRes struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
)

const (
	msgTpl = `<html> <body> {{.Msg}} </body> </html>`
)

func (r *RefreshReq) Refresh(clientID, clientSec, refreshToken string) (res TokenRes, err error) {
	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("client_secret", clientSec)
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", refreshToken)
	resp, err := http.PostForm(r.URL, values)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(b))
		return res, errors.New("refresh token error. go https://glassof.garden/oauth")
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &res)
	return
}

func NewRefresh(url string) RefreshReq {
	return RefreshReq{URL: url}
}

func StartHandler(reqURL, pass string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.PostForm.Get("pass") == pass {
			http.Redirect(w, r, reqURL, http.StatusFound)
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
}

func CallbackHandler(vals url.Values, tokenURL, accessPath, refreshPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vals.Add("code", r.URL.Query().Get("code"))
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
		var tres TokenRes
		err = json.Unmarshal(b, &tres)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}
		ioutil.WriteFile(accessPath, []byte(tres.AccessToken), os.ModePerm)
		ioutil.WriteFile(refreshPath, []byte(tres.RefreshToken), os.ModePerm)
	}
}
