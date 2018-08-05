package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gong023/my-slack-process/slack"
	"os"
	"strconv"
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

	Illust struct {
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

func main() {
	clientID := os.Getenv("CLI_ID")
	deviceToken := os.Getenv("DEVICE_TOKEN")
	clientSec := os.Getenv("CLI_SEC")
	refreshToken := os.Getenv("REF_TOKEN")
	host := os.Getenv("PROXY_HOST")
	if clientID == "" {
		log.Fatal("CLI_ID is not given")
	}
	if deviceToken == "" {
		log.Fatal("DEVICE_TOKEN is not given")
	}
	if clientSec == "" {
		log.Fatal("CLI_SEC is not given")
	}
	if refreshToken == "" {
		log.Fatal("REF_TOKEN is not given")
	}
	if host == "" {
		log.Fatal("PROXY_HOST is not given")
	}

	s := 20
	if os.Getenv("SINCE") != "" {
		is, err := strconv.Atoi(os.Getenv("SINCE"))
		if err != nil {
			log.Fatal(err)
		}
		s = is
	}

	token, err := getToken(clientID, clientSec, deviceToken, refreshToken)
	if err != nil {
		log.Fatal(err)
	}

	illusts, err := getFollowingIllusts(token)
	if err != nil {
		log.Fatal(err)
	}

	since := time.Now().Add(-1 * time.Duration(s) * time.Minute)
	for _, illust := range illusts.Illusts {
		create, err := time.Parse(time.RFC3339, illust.CreateDate)
		if err != nil {
			log.Fatal(err)
		}
		if create.Before(since) {
			continue
		}

		attachments := slack.Attachments{
			Attachments: []slack.Attachment{
				{
					Title:   illust.Title,
					Pretext: illust.Caption,
				},
			},
		}
		if len(illust.MetaPages) <= 0 {
			attachments.Attachments = append(attachments.Attachments, slack.Attachment{
				ImageURL: host + "?" + imagePath(illust.ImageURLs.Medium),
			})
		} else {
			for _, metaPage := range illust.MetaPages {
				attachments.Attachments = append(attachments.Attachments, slack.Attachment{
					ImageURL: host + "?" + imagePath(metaPage.Medium),
				})
			}
		}

		b, err := json.Marshal(attachments)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	}
}

func getToken(client_id, client_sec, device_token, refresh_token string) (token TokenData, err error) {
	data := url.Values{}
	data.Add("client_id", client_id)
	data.Add("client_secret", client_sec)
	data.Add("device_token", device_token)
	data.Add("refresh_token", refresh_token)
	data.Add("grant_type", "refresh_token")
	data.Add("get_secure_url", "true")
	body := bytes.NewReader([]byte(data.Encode()))
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://oauth.secure.pixiv.net/auth/token", body)
	commonHeader(req.Header)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return
	}

	res, err := client.Do(req)
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

func getFollowingIllusts(token TokenData) (illusts FollowingIllusts, err error) {
	q := url.Values{}
	q.Add("restrict", "all")
	url := "https://app-api.pixiv.net/v2/illust/follow?" + q.Encode()
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	commonHeader(req.Header)
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	if err != nil {
		return
	}

	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	if res.StatusCode >= 300 {
		return illusts, errors.New(fmt.Sprintf("following error [%s]:(%s)", res.Status, string(b)))
	}

	err = json.Unmarshal(b, &illusts)
	if err != nil {
		return
	}

	return
}

func imagePath(origin string) string {
	originURL, _ := url.Parse(origin)
	u := url.Values{}
	u.Add("q", originURL.RequestURI())
	return u.Encode()
}

func commonHeader(h http.Header) {
	h.Add("Accept-Language", "ja-jp")
	h.Add("App-Version", "7.0.5")
	h.Add("User-Agent", "PixivIOSApp/7.0.5 (iOS 11.1; iPad4,4)")
	h.Add("App-OS-Version", "11.1")
	h.Add("X-Client-Time", time.Now().Format(time.RFC3339))
}
