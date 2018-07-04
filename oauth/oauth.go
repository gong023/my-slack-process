package oauth

import (
	"encoding/json"
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
		AccessToken  string `json:"access_token"`
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
		return res, fmt.Errorf("%s. go https://cron.gonge.fun/oauth", string(b))
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
