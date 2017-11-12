package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
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
	flag.Parse()
	if *refreshPath == "" || *clientID == "" || *clientSec == "" || *tags == "" {
		log.Fatal("missing parameter")
	}
	refreshToken, err := ioutil.ReadFile(*refreshPath)
	if err != nil {
		log.Fatal(err)
	}

	tres := getToken(*clientID, *clientSec, string(refreshToken))
	for _, tag := range strings.Split(*tags, ",") {
		streamRes := getStream(tres.AccessToken, tag)
		readQuery := url.Values{}
		readQuery.Add("a", "user/-/state/com.google/read")
		for _, item := range streamRes.Items {
			fmt.Println(item.Canonical[0].Href)
			readQuery.Add("i", item.ID)
		}
		if len(streamRes.Items) <= 0 {
			continue
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
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Fatal("refresh token error. go https://glassof.garden/oauth")
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var tres TokenRes
	err = json.Unmarshal(b, &tres)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != 200 {
		log.Fatal("invalid stream api response: " + string(b))
	}
	var streamRes StreamRes
	err = json.Unmarshal(b, &streamRes)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != 200 {
		log.Fatal("mark read api response: " + string(b))
	}
}
