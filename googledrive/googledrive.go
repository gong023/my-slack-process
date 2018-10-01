package googledrive

import (
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Create(localPath, driveFolder string) (link string, err error) {
	b, err := ioutil.ReadFile("/etc/serviceaccount.json")
	if err != nil {
		return
	}

	config, err := google.JWTConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return
	}
	client := config.Client(oauth2.NoContext)

	srv, err := drive.New(client)
	if err != nil {
		return
	}

	f, err := os.Open(localPath)
	if err != nil {
		return
	}
	_, fn := filepath.Split(f.Name())

	df := &drive.File{
		Parents: []string{driveFolder},
		Name:    fn,
	}
	ff, err := srv.Files.Create(df).Media(f).Do()
	if err != nil {
		return
	}
	return fmt.Sprintf("https://drive.google.com/file/d/%s/view", ff.Id), nil
}
