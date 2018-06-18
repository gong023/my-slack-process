package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"
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

func main() {
	email := os.Getenv("EMAIL")
	password := os.Getenv("PASS")

	s := flag.Duration("since", 5*time.Hour, "get since")
	localSave := flag.String("local", os.TempDir(), "path to save tmp file in local")
	flag.Parse()
	if email == "" || password == "" {
		log.Fatal("missing parameter")
	}
	since := time.Now().Add(*(s) * -1)

	auth, err := login(email, password)
	if err != nil {
		log.Fatal(err)
	}

	favs, err := getFavorites(auth)
	if err != nil {
		log.Fatal(err)
	}

	jloc, _ := time.LoadLocation("Asia/Tokyo")

	var eg errgroup.Group
	for _, fav := range favs {
		fav := fav
		eg.Go(func() error {
			episode := fav.Program.Episode
			updated, err := time.ParseInLocation("2006/01/02 15:04:05", episode.UpdatedAt, jloc)
			if err != nil {
				return err
			}
			if updated.Before(since) {
				return err
			}

			localFiles := make(map[string]PlayCheck)

			playCheck, err := getPlayCheck(auth, episode.Video)
			if err != nil {
				return err
			}
			filePath := fmt.Sprintf("%s/[%s]%s.aac", *localSave, fav.Program.LatestEpisodeName, fav.Program.Name)
			localFiles[filePath] = playCheck

			if fav.Program.Episode.AdditionalVideo.ID != 0 {
				playCheck, err := getPlayCheck(auth, fav.Program.Episode.AdditionalVideo)
				if err != nil {
					return err
				}
				filePath := fmt.Sprintf("%s/[%s]%s-omake.aac", *localSave, fav.Program.LatestEpisodeName, fav.Program.Name)
				localFiles[filePath] = playCheck
			}

			var veg errgroup.Group
			for fileName, playCheck := range localFiles {
				fileName, playCheck := fileName, playCheck
				veg.Go(func() error {
					return ffmpeg(playCheck.PlaylistURL, fileName)
				})
			}

			return veg.Wait()
		})
	}

	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}

func login(email, password string) (auth Auth, err error) {
	data := url.Values{}
	data.Add("email", email)
	data.Add("password", password)
	body := bytes.NewReader([]byte(data.Encode()))
	req, err := http.NewRequest("POST", host+"/api/v1/users/auth", body)
	if err != nil {
		return
	}
	commonHeader(req.Header, auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
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
		return auth, errors.New(fmt.Sprintf("login error [%s]:(%s)", res.Status, string(b)))
	}

	err = json.Unmarshal(b, &auth)
	return
}

func getFavorites(auth Auth) (favs []FavRes, err error) {
	req, err := http.NewRequest("GET", host+"/api/v1/user_favorites", nil)
	if err != nil {
		return
	}
	commonHeader(req.Header, auth)

	client := &http.Client{}
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
		return favs, errors.New(fmt.Sprintf("user_favorites error [%s]:(%s)", res.Status, string(b)))
	}
	err = json.Unmarshal(b, &favs)
	return
}

func getPlayCheck(auth Auth, video Video) (playCheck PlayCheck, err error) {
	videoID := strconv.Itoa(video.ID)
	req, err := http.NewRequest("GET", host+"/api/v1/videos/play_check?video_id="+videoID, nil)
	if err != nil {
		return
	}
	commonHeader(req.Header, auth)

	client := &http.Client{}
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
		return playCheck, errors.New(fmt.Sprintf("play_check error [%s]:(%s)", res.Status, string(b)))
	}
	err = json.Unmarshal(b, &playCheck)
	return
}

func ffmpeg(m3u8URL, saveFile string) (err error) {
	args := []string{"-y", "-vn", "-i", m3u8URL, "-acodec", "copy", "-bsf:a", "aac_adtstoasc", saveFile}
	return exec.Command("ffmpeg", args...).Run()
}

func commonHeader(h http.Header, auth Auth) {
	h.Add("User-Agent", ua)
	h.Add("X-BushiroadMusic-App-Version", appVer)
	h.Add("X-BushiroadMusic-PackageIdentifier", appBundle)
	h.Add("X-BushiroadMusic-Os", "ipad")
	h.Add("X-Hibiki-User-Id", strconv.Itoa(auth.ID))
	h.Add("user-id", strconv.Itoa(auth.ID))
	h.Add("X-Hibiki-Access-Token", auth.AccessToken)
}
