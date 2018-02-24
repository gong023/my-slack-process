package hls

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang/groupcache/lru"
	"github.com/kz26/m3u8"
)

type Playlist struct {
	URI           string
	totalDuration time.Duration
}

// todo
func Download(client *http.Client, req *http.Request, savePath string) (err error) {
	res, err := client.Do(req)
	if err != nil {
		return
	}
	playlist, _, err := m3u8.DecodeFrom(res.Body, true)
	mpl, ok := playlist.(*m3u8.MasterPlaylist)
	if !ok {
		return errors.New("invalid playlist type")
	}
	r, err := http.NewRequest("GET", mpl.Variants[0].URI, nil)
	if err != nil {
		return
	}

	psChan := make(chan *Playlist, 1024)
	go getPlaylist(client, r, psChan)
	err = downloadSegment(client, savePath, psChan)
	return
}

func getPlaylist(client *http.Client, req *http.Request, psChan chan *Playlist) error {
	cache := lru.New(1024)
	recDuration := time.Duration(0)

	for {
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		playlist, listType, err := m3u8.DecodeFrom(resp.Body, true)
		resp.Body.Close()
		if listType != m3u8.MEDIA {
			return errors.New("invalid media playlist")
		}
		mpl, ok := playlist.(*m3u8.MediaPlaylist)
		if !ok {
			return errors.New("invalid playlist type")
		}

		for _, v := range mpl.Segments {
			if v == nil {
				continue
			}
			var msURI string
			if strings.HasPrefix(v.URI, "http") {
				msURI, err = url.QueryUnescape(v.URI)
				if err != nil {
					return err
				}
			} else {
				msURL, err := req.URL.Parse(v.URI)
				if err != nil {
					return err
				}
				msURI, err = url.QueryUnescape(msURL.String())
				if err != nil {
					return err
				}
			}
			_, ok := cache.Get(msURI)
			if !ok {
				cache.Add(msURI, nil)
				recDuration += time.Duration(int64(v.Duration * 1000000000))
				psChan <- &Playlist{msURI, recDuration}
			}
		}

		if mpl.Closed {
			close(psChan)
			return nil
		}
		time.Sleep(time.Duration(int64(mpl.TargetDuration * 1000000000)))
	}
}

func downloadSegment(client *http.Client, savePath string, psChan chan *Playlist) error {
	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()

	for v := range psChan {
		req, err := http.NewRequest("GET", v.URI, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, resp.Body); err != nil {
			return err
		}
		resp.Body.Close()
		log.Printf("Downloaded %v\n", v.URI)
	}
	return nil
}
