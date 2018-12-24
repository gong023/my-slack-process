package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const (
	ua        = "Mozilla/5.0 (iPad; CPU OS 11_1 like Mac OS X) AppleWebKit/604.3.5 (KHTML, like Gecko) Mobile/15B93 BushiroadMusic(com.bushiroad.HibikiRadio)/2.1.6"
	host      = "https://vcms-api.hibiki-radio.jp"
	appVer    = "20106"
	appBundle = "com.bushiroad.HibikiRadio"
)

type (
	Auth struct {
		ID          int    `json:"id"`
		AccessToken string `json:"access_token"`
	}

	FavRes struct {
		Program Program
	}

	Program struct {
		Name              string  `json:"name"`
		ID                int     `json:"id"`
		LatestEpisodeName string  `json:"latest_episode_name"`
		Episode           Episode `json:"episode"`
	}

	Episode struct {
		ID              int    `json:"id"`
		Video           Video  `json:"video"`
		AdditionalVideo Video  `json:"additional_video"`
		UpdatedAt       string `json:"updated_at"`
	}

	Video struct {
		ID int `json:"id"`
	}

	PlayCheck struct {
		Token       string
		PlaylistURL string `json:"playlist_url"`
	}
)

type Session struct {
	auth Auth
}

func (s *Session) Start(email, password string) error {
	data := url.Values{}
	data.Add("email", email)
	data.Add("password", password)
	body := bytes.NewReader([]byte(data.Encode()))
	req, err := http.NewRequest("POST", host+"/api/v1/users/auth", body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	auth := Auth{}
	err = s.do(req, &auth)
	if err != nil {
		return err
	}
	s.auth = auth

	return nil
}

func (s *Session) Favorites() (fav []FavRes, err error) {
	req, err := http.NewRequest("GET", host+"/api/v1/user_favorites", nil)
	if err != nil {
		return
	}
	err = s.do(req, &fav)
	return
}

func (s *Session) PlayCheck(video Video) (playCheck PlayCheck, err error) {
	videoID := strconv.Itoa(video.ID)
	req, err := http.NewRequest("GET", host+"/api/v1/videos/play_check?video_id="+videoID, nil)
	if err != nil {
		return
	}
	err = s.do(req, &playCheck)
	return
}

func (s *Session) do(req *http.Request, res interface{}) error {
	req.Header.Add("User-Agent", ua)
	req.Header.Add("X-BushiroadMusic-App-Version", appVer)
	req.Header.Add("X-BushiroadMusic-PackageIdentifier", appBundle)
	req.Header.Add("X-BushiroadMusic-Os", "ipad")
	req.Header.Add("X-Hibiki-User-Id", strconv.Itoa(s.auth.ID))
	req.Header.Add("user-id", strconv.Itoa(s.auth.ID))
	req.Header.Add("X-Hibiki-Access-Token", s.auth.AccessToken)

	client := &http.Client{}
	r, err := client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if r.StatusCode >= 300 {
		return fmt.Errorf("%s error [%s]:(%s)", req.URL.Path, r.Status, string(b))
	}

	return json.Unmarshal(b, res)
}
