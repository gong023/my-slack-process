package googledrive

import (
	"bytes"
	"context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"io/ioutil"
	"net/http"
)

type Client struct {
	cli *http.Client
}

func CreateFromSvcCredential(ctx context.Context, path string) (cli *Client, err error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	config, err := google.JWTConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return
	}

	return &Client{cli: config.Client(ctx)}, nil
}

func (c *Client) Upload(name string, data []byte) error {
	svc, err := drive.New(c.cli)
	if err != nil {
		return err
	}
	f := &drive.File{Name: name}
	b := bytes.NewBuffer(data)
	svc.Files.GenerateIds()
	_, err = svc.Files.Create(f).Media(b).Fields("test").Do()
	return err
}
