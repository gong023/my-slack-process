package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gong023/my-slack-process/googledrive"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type (
	cli struct {
		token, deviceID, deviceName string
		client                      http.Client
	}

	Favorites struct {
		ProgramIDs []int `json:"program_ids"`
	}

	Program struct {
		Episodes []struct {
			Bonus        bool   `json:"bonus"`
			Current      bool   `json:"current"`
			Description  string `json:"description"`
			EpisodeFiles []struct {
				ID       int    `json:"id"`
				MediaURL string `json:"media_url"`
				Target   string `json:"target"`
			} `json:"episode_files"`
			EpisodeImages []struct {
				Display      bool   `json:"display"`
				DisplayOrder int    `json:"display_order"`
				ID           int    `json:"id"`
				URL          string `json:"url"`
			} `json:"episode_images"`
			Free        bool        `json:"free"`
			ID          int         `json:"id"`
			MediaType   string      `json:"media_type"`
			OnPlaylist  bool        `json:"on_playlist"`
			Premium     bool        `json:"premium"`
			Sticky      bool        `json:"sticky"`
			TagImageURL interface{} `json:"tag_image_url"`
			Title       string      `json:"title"`
			UpdatedOn   string      `json:"updated_on"`
		} `json:"episodes"`
		Favorite      bool     `json:"favorite"`
		Free          bool     `json:"free"`
		GenreList     []string `json:"genre_list"`
		ID            int      `json:"id"`
		MediaCategory string   `json:"media_category"`
		MediumList    []string `json:"medium_list"`
		New           bool     `json:"new"`
		PerformerList []string `json:"performer_list"`
		Title         string   `json:"title"`
		TitleList     []string `json:"title_list"`
	}
)

const host = "https://app.onsen.ag"

func main() {
	token := strings.TrimSuffix(os.Getenv("TOKEN"), "\n")
	if token == "" {
		log.Fatal("token is not given")
	}
	deviceID := strings.TrimSuffix(os.Getenv("DEVICE_ID"), "\n")
	if deviceID == "" {
		log.Fatal("deviceID is not given")
	}
	deviceName := strings.TrimSuffix(os.Getenv("DEVICE_NAME"), "\n")
	if deviceName == "" {
		log.Fatal("deviceName is not given")
	}
	driveDirID := strings.TrimSuffix(os.Getenv("DRIVE_DIR_ID"), "\n")
	if driveDirID == "" {
		log.Fatal("driveDirID is not given")
	}

	localSave := flag.String("local", os.TempDir(), "path to save tmp file in local")
	s := flag.Duration("since", 5*time.Hour, "get since")
	dry := flag.Bool("dry", false, "ffmpeg dry")
	flag.Parse()
	if *localSave == "" {
		log.Fatal("localSave is not given")
	}
	since := time.Now().Add(*(s) * -1)
	jloc, _ := time.LoadLocation("Asia/Tokyo")

	cli := newCli(token, deviceID, deviceName)
	favs, err := cli.favorites()
	if err != nil {
		log.Fatal(err)
	}

	var eg errgroup.Group
	for _, programID := range favs.ProgramIDs {
		programID := programID
		eg.Go(func() error {
			program, err := cli.program(programID)
			if err != nil {
				log.Fatal(err)
			}

			var veg errgroup.Group
			for _, episode := range program.Episodes {
				episode := episode
				updated, err := time.ParseInLocation("2006-01-02T15:04:05.000-07:00", episode.UpdatedOn, jloc)
				if err != nil {
					log.Fatal(err)
				}
				if updated.Before(since) {
					continue
				}

				veg.Go(func() error {
					file := fmt.Sprintf("%s/[%s]%s.aac", *localSave, episode.Title, program.Title)
					if *dry {
						fmt.Println(file)
						return nil
					}

					link, err := runBackUp(episode.EpisodeFiles[0].MediaURL, file, driveDirID)
					if err != nil {
						log.Fatal(err)
					}
					_, f := filepath.Split(file)
					fmt.Printf("%s %s\n", f, link)
					return nil
				})
			}
			return veg.Wait()
		})

	}

	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}

func newCli(token, deviceID, deviceName string) *cli {
	return &cli{
		token:      token,
		deviceID:   deviceID,
		deviceName: deviceName,
		client: http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *cli) favorites() (favs Favorites, err error) {
	req, err := http.NewRequest("GET", host+"/api/me/favorites/ids", nil)
	if err != nil {
		return
	}
	err = c.do(req, &favs)
	return
}

func (c *cli) program(id int) (program Program, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/me/programs/%d", host, id), nil)
	if err != nil {
		return
	}
	err = c.do(req, &program)
	return
}

func (c *cli) do(req *http.Request, res interface{}) error {
	req.Header.Add("x-device-identifier", c.deviceID)
	req.Header.Add("x-app-version", "25")
	req.Header.Add("x-device-os", "ios")
	req.Header.Add("x-device-name", c.deviceName)
	req.Header.Add("user-agent", "iOS/Onsen/2.6.1")
	req.Header.Add("content-type", "json")
	req.Header.Add("authorization", "Bearer "+c.token)

	r, err := c.client.Do(req)
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

func runBackUp(m3u8URL, savePath, driveDirID string) (string, error) {
	if err := ffmpegStream(m3u8URL, savePath); err != nil {
		return "", err
	}
	if err := ffmpegToMP3(savePath); err != nil {
		return "", err
	}
	return googledrive.Create(savePath+".mp3", driveDirID)
}

func ffmpegStream(m3u8URL, saveFile string) error {
	args := []string{"-y", "-vn", "-i", m3u8URL, "-acodec", "copy", "-bsf:a", "aac_adtstoasc", saveFile}
	return exec.Command("ffmpeg", args...).Run()
}

func ffmpegToMP3(filePath string) error {
	args := []string{"-y", "-i", filePath, "-acodec", "libmp3lame", "-ac", "2", "-ab", "160", filePath + ".mp3"}
	return exec.Command("ffmpeg", args...).Run()
}
