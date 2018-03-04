package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type (
	RefreshReq struct {
		URL string
	}

	TokenRes struct {
		AccessToken  string `access_token:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
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
	var tres TokenRes
	err = json.Unmarshal(b, &tres)
	if err != nil {
		return
	}
	return
}

func NewRefresh(url string) RefreshReq {
	return RefreshReq{URL: url}
}
