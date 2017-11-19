package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
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
	client_id := flag.String("client_id", "", "client id")
	device_token := flag.String("device_token", "", "device_token")
	client_sec := flag.String("client_sec", "", "client secret")
	refresh_token := flag.String("refresh_token", "", "refresh token")
	save := flag.String("save", "/tmp", "path to save")
	s := flag.Int64("since", 20, "following illust since X min")
	flag.Parse()
	if *client_id == "" || *device_token == "" || *client_sec == "" || *refresh_token == "" {
		log.Fatal("missing parameter")
	}

	token, err := getToken(*client_id, *client_sec, *device_token, *refresh_token)
	if err != nil {
		log.Fatal(err)
	}

	illusts, err := getFollowingIllusts(token)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	since := time.Now().Add(-1 * time.Duration(*s) * time.Minute)
	for _, illust := range illusts.Illusts {
		create, err := time.Parse(time.RFC3339, illust.CreateDate)
		if err != nil {
			log.Fatal(err)
		}
		if create.Before(since) {
			continue
		}

		h := md5.New()
		io.WriteString(h, illust.Caption)
		path := *save + "/" + fmt.Sprintf("%x", h.Sum(nil)) + "/"
		os.MkdirAll(path, os.ModePerm)
		ioutil.WriteFile(path+"/title.txt", []byte(illust.Title), os.ModePerm)
		ioutil.WriteFile(path+"/caption.txt", []byte(illust.Caption), os.ModePerm)

		if len(illust.MetaPages) <= 0 {
			wg.Add(1)
			go func(save, imageURL string) {
				defer wg.Done()
				b, _ := getImage(imageURL)
				ioutil.WriteFile(save, b, os.ModePerm)
			}(path+"0.jpeg", illust.ImageURLs.Medium)
			continue
		}

		for i, metaPage := range illust.MetaPages {
			wg.Add(1)
			go func(save, imageURL string) {
				defer wg.Done()
				b, _ := getImage(imageURL)
				ioutil.WriteFile(save, b, os.ModePerm)
			}(path+strconv.Itoa(i)+".jpeg", metaPage.Medium)
		}
	}
	wg.Wait()
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

func getImage(url string) (b []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Add("Referer", "https://app-api.pixiv.net/")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	return
}

func commonHeader(h http.Header) {
	h.Add("Accept-Language", "ja-jp")
	h.Add("App-Version", "7.0.5")
	h.Add("User-Agent", "PixivIOSApp/7.0.5 (iOS 11.1; iPad4,4)")
	h.Add("App-OS-Version", "11.1")
	h.Add("X-Client-Time", time.Now().Format(time.RFC3339))
}
