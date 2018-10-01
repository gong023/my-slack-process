package main

import (
	"flag"
	"fmt"
	"github.com/gong023/my-slack-process/googledrive"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	email := os.Getenv("EMAIL")
	password := os.Getenv("PASS")
	driveDirID := os.Getenv("DRIVE_DIR_ID")

	s := flag.Duration("since", 5*time.Hour, "get since")
	localSave := flag.String("local", os.TempDir(), "path to save tmp file in local")
	flag.Parse()
	if email == "" || password == "" {
		log.Fatal("missing parameter")
	}
	since := time.Now().Add(*(s) * -1)

	session := &Session{}
	err := session.Start(email, password)
	if err != nil {
		log.Fatal(err)
	}

	favs, err := session.Favorites()
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

			playCheck, err := session.PlayCheck(episode.Video)
			if err != nil {
				return err
			}
			filePath := fmt.Sprintf("%s/[%s]%s.aac", *localSave, fav.Program.LatestEpisodeName, fav.Program.Name)
			localFiles[filePath] = playCheck

			if fav.Program.Episode.AdditionalVideo.ID != 0 {
				playCheck, err := session.PlayCheck(fav.Program.Episode.AdditionalVideo)
				if err != nil {
					return err
				}
				filePath := fmt.Sprintf("%s/[%s]%s-omake.aac", *localSave, fav.Program.LatestEpisodeName, fav.Program.Name)
				localFiles[filePath] = playCheck
			}

			var veg errgroup.Group
			for filePath, playCheck := range localFiles {
				filePath, playCheck := filePath, playCheck
				veg.Go(func() error {
					link, err := runBackUp(playCheck.PlaylistURL, filePath, driveDirID)
					if err != nil {
						return err
					}
					_, f := filepath.Split(filePath)
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

func runBackUp(m3u8URL, savePath, driveDirID string) (string, error) {
	if err := ffmpegStream(m3u8URL, savePath); err != nil {
		return "", err
	}
	if err := ffmpegToMP3(savePath); err != nil {
		return "", err
	}
	return googledrive.Create(savePath+".mp3", driveDirID)
}

func ffmpegStream(m3u8URL, saveFile string) (err error) {
	args := []string{"-y", "-vn", "-i", m3u8URL, "-acodec", "copy", "-bsf:a", "aac_adtstoasc", saveFile}
	return exec.Command("ffmpeg", args...).Run()
}

func ffmpegToMP3(filePath string) error {
	args := []string{"-y", "-i", filePath, "-acodec", "libmp3lame", "-ac", "2", "-ab", "160", filePath + ".mp3"}
	return exec.Command("ffmpeg", args...).Run()
}
