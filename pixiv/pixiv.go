package pixiv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"os"
)

type (
	TokenRes struct {
		Response TokenData `json:"response"`
	}

	TokenData struct {
		AccessToken string `json:"access_token"`
		User        User   `json:"user"`
	}

	User struct {
		ID string `json:"id"`
	}

	FollowingIllusts struct {
		Illusts []Illust `json:"illusts"`
	}

	RankingIllusts struct {
		Illusts []Illust `json:"illusts"`
	}

	Illust struct {
		ID         int `json:"id"`
		Title      string
		Caption    string
		CreateDate string `json:"create_date"`
		ImageURLs  `json:"image_urls"`
		MetaPages  []MetaPage `json:"meta_pages"`
	}

	MetaPage struct {
		ImageURLs `json:"image_urls"`
	}

	ImageURLs struct {
		SquareMedium string `json:"square_medium"`
		Medium       string `json:"medium"`
		Large        string `json:"large"`
		Original     string `json:"original"`
	}

	SaveChan struct {
		Title     string
		Caption   string
		ImageURLs []string
	}
)

var cli = &http.Client{
	Timeout: 20 * time.Minute,
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, e error) {
			return net.DialTimeout(network, addr, 15*time.Minute)
		},
		TLSHandshakeTimeout:   10 * time.Minute,
		ResponseHeaderTimeout: 10 * time.Minute,
		ExpectContinueTimeout: 10 * time.Minute,
	},
}

func GetToken() (token TokenData, err error) {
	clientID := os.Getenv("CLI_ID")
	deviceToken := os.Getenv("DEVICE_TOKEN")
	clientSec := os.Getenv("CLI_SEC")
	refreshToken := os.Getenv("REF_TOKEN")
	if clientID == "" {
		return token, errors.New("CLI_ID is not given")
	}
	if deviceToken == "" {
		return token, errors.New("DEVICE_TOKEN is not given")
	}
	if clientSec == "" {
		return token, errors.New("CLI_SEC is not given")
	}
	if refreshToken == "" {
		return token, errors.New("REF_TOKEN is not given")
	}

	data := url.Values{}
	data.Add("client_id", clientID)
	data.Add("client_secret", clientSec)
	data.Add("device_token", deviceToken)
	data.Add("refresh_token", refreshToken)
	data.Add("grant_type", "refresh_token")
	data.Add("get_secure_url", "true")
	body := bytes.NewReader([]byte(data.Encode()))
	req, err := http.NewRequest("POST", "https://oauth.secure.pixiv.net/auth/token", body)
	commonHeader(req.Header)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return
	}

	res, err := cli.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	if res.StatusCode >= 300 {
		return token, errors.New("token error [code]:" + res.Status)
	}

	var tokenRes TokenRes
	err = json.Unmarshal(b, &tokenRes)
	if err != nil {
		return
	}

	return tokenRes.Response, nil
}

func GetFollowingIllusts(token TokenData) (illusts FollowingIllusts, err error) {
	q := url.Values{}
	q.Add("restrict", "all")
	url := "https://app-api.pixiv.net/v2/illust/follow?" + q.Encode()
	req, err := http.NewRequest("GET", url, nil)
	commonHeader(req.Header)
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	if err != nil {
		return
	}

	res, err := cli.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	if res.StatusCode >= 300 {
		return illusts, fmt.Errorf("following error [%s]:(%s)", res.Status, string(b))
	}

	err = json.Unmarshal(b, &illusts)
	if err != nil {
		return
	}

	return
}

func GetDailyRankingIllusts(token TokenData) (illusts RankingIllusts, err error) {
	q := url.Values{}
	q.Add("mode", "day_manga") // or week_manga, month_manga
	q.Add("filter", "for_ios")
	u := "https://app-api.pixiv.net/v1/illust/ranking?" + q.Encode()
	req, err := http.NewRequest("GET", u, nil)
	commonHeader(req.Header)
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	if err != nil {
		return
	}

	res, err := cli.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	if res.StatusCode >= 300 {
		return illusts, fmt.Errorf("following error [%s]:(%s)", res.Status, string(b))
	}

	err = json.Unmarshal(b, &illusts)
	if err != nil {
		return
	}

	return
}

func ImagePath(origin string) string {
	originURL, _ := url.Parse(origin)
	u := url.Values{}
	u.Add("q", originURL.RequestURI())
	return u.Encode()
}

func commonHeader(h http.Header) {
	h.Add("Accept-Language", "ja-jp")
	h.Add("App-Version", "7.6.2")
	h.Add("App-Os-Version", "12.2")
	h.Add("User-Agent", "PixivIOSApp/7.6.2 (iOS 12.2; iPad7,3)")
	h.Add("X-Client-Time", time.Now().Format(time.RFC3339))
}
