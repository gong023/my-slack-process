package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gong023/my-slack-process/util"
)

type (
	TokenRes struct {
		AccessToken string `json:"access_token"`
	}
	StreamRes struct {
		Items []Item
	}

	Item struct {
		ID        string `json:"id"`
		Title     string
		Canonical []Canonical
		Summary   Summary
		Origin    Origin
	}

	Canonical struct {
		Href string
	}

	Summary struct {
		Content string
	}

	Origin struct {
		Title string
	}
)

func main() {
	refreshPath := flag.String("refresh_path", "", "refresh token file path")
	clientID := flag.String("client_id", "", "inoreader client id")
	clientSec := flag.String("client_sec", "", "inoreader client secret")
	tags := flag.String("tags", "", "target innoreader tags. separate by comma(,)")
	webhook := flag.String("webhook", "", "slack incoming webhook url")
	flag.Parse()
	if *refreshPath == "" || *clientID == "" || *clientSec == "" || *tags == "" || *webhook == "" {
		util.LogErr(errors.New("missing parameter"))
	}
	refreshToken, err := ioutil.ReadFile(*refreshPath)
	if err != nil {
		util.LogErr(err)
	}

	tres := getToken(*clientID, *clientSec, string(refreshToken))
	for _, tag := range strings.Split(*tags, ",") {
		streamRes := getStream(tres.AccessToken, tag)
		attachments := []util.Attachment{}
		readQuery := url.Values{}
		readQuery.Add("a", "user/-/state/com.google/read")
		for _, item := range streamRes.Items {
			attachment := util.Attachment{
				Pretext: item.Canonical[0].Href,
				Text:    item.Title + "\n" + item.Origin.Title,
			}
			attachments = append(attachments, attachment)
			readQuery.Add("i", item.ID)
		}
		if len(attachments) <= 0 {
			continue
		}
		err = util.PostAttachMents(*webhook, util.SlackAttachments{
			Attachments: attachments,
		})
		if err != nil {
			util.LogErr(err)
		}
		markRead(tres.AccessToken, readQuery)
	}
}

func getToken(clientID, clientSec, refreshToken string) TokenRes {
	vals := url.Values{}
	vals.Set("client_id", clientID)
	vals.Set("client_secret", clientSec)
	vals.Set("grant_type", "refresh_token")
	vals.Set("refresh_token", refreshToken)
	resp, err := http.PostForm("https://www.inoreader.com/oauth2/token", vals)
	if err != nil {
		util.LogErr(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		util.LogErr(errors.New("refresh token error. go https://glassof.garden/oauth"))
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		util.LogErr(err)
	}
	var tres TokenRes
	err = json.Unmarshal(b, &tres)
	if err != nil {
		util.LogErr(err)
	}

	return tres
}

func getStream(token, tag string) StreamRes {
	client := &http.Client{}
	q := url.Values{}
	q.Set("xt", "user/-/state/com.google/read") // exclude target
	q.Set("n", "50")
	url := "https://www.inoreader.com/reader/api/0/stream/contents/user/-/label/" + tag + "?" + q.Encode()
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		util.LogErr(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		util.LogErr(err)
	}
	if res.StatusCode != 200 {
		util.LogErr(errors.New("invalid stream api response: " + string(b)))
	}
	var streamRes StreamRes
	err = json.Unmarshal(b, &streamRes)
	if err != nil {
		util.LogErr(err)
	}

	return streamRes
}

func markRead(token string, q url.Values) {
	client := &http.Client{}
	url := "https://www.inoreader.com/reader/api/0/edit-tag?" + q.Encode()
	req, err := http.NewRequest("POST", url, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		util.LogErr(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		util.LogErr(err)
	}
	if res.StatusCode != 200 {
		util.LogErr(errors.New("mark read api response: " + string(b)))
	}
}
