package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gong023/my-slack-process/oauth"
	"github.com/gong023/my-slack-process/slack"
	"os"
)

type (
	StreamRes struct {
		Items []Item
	}

	Item struct {
		ID        string `json:"id"`
		Title     string
		Canonical []Canonical
		Summary   Summary
		Origin    Origin
		Enclosure []Enclosure
	}

	Canonical struct {
		Href string
	}

	Summary struct {
		Content string
	}

	Origin struct {
		Title   string
		HtmlURL string `json:"htmlUrl"`
	}

	Enclosure struct {
		Href string
	}
)

func main() {
	refreshRes := flag.String("refresh_res", "", "refresh token response")
	flag.Parse()
	clientID := os.Getenv("CLI_ID")
	clientSec := os.Getenv("CLI_SEC")
	tags := os.Getenv("TAG")
	if *refreshRes == "" {
		log.Fatal("missing parameter:refresh")
	}
	if clientID == "" {
		log.Fatal("missing parameter:clientID")
	}
	if clientSec == "" {
		log.Fatal("missing parameter:clientSec")
	}
	if tags == "" {
		log.Fatal("missing parameter:tags")
	}

	var r oauth.TokenRes
	err := json.Unmarshal([]byte(*refreshRes), &r)
	if err != nil {
		log.Fatal(err)
	}

	req := oauth.NewRefresh("https://www.inoreader.com/oauth2/token")
	tres, err := req.Refresh(clientID, clientSec, r.RefreshToken)
	if err != nil {
		log.Fatal(err)
	}

	for _, tag := range strings.Split(tags, ",") {
		streamRes := getStream(tres.AccessToken, tag)
		readQuery := url.Values{}
		readQuery.Add("a", "user/-/state/com.google/read")
		for _, item := range streamRes.Items {
			outPutItem(item)
			readQuery.Add("i", item.ID)
		}
		if len(streamRes.Items) <= 0 {
			continue
		}
		markRead(tres.AccessToken, readQuery)
	}
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

func outPutItem(item Item) {
	re := regexp.MustCompile("<.+>")
	summary := re.ReplaceAllString(item.Summary.Content, "")
	summary = strings.TrimSpace(summary)
	summary = strings.Join(strings.Fields(summary), " ")
	if r := []rune(summary); len(r) > 400 {
		summary = string(r[:400])
	}

	attachment := slack.Attachment{
		Fallback:   item.Canonical[0].Href,
		AuthorName: item.Origin.Title,
		Text:       summary,
		Title:      item.Title,
		TitleLink:  item.Canonical[0].Href,
	}

	if len(item.Enclosure) >= 1 {
		attachment.ImageURL = item.Enclosure[0].Href
	}

	if u, err := url.Parse(item.Canonical[0].Href); err == nil {
		attachment.AuthorIcon = u.Scheme + "://" + u.Host + "/favicon.ico"
	}

	b, err := json.Marshal(attachment)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
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
