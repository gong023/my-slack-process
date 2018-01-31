package main

import (
	"errors"
	"flag"
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

var client = &http.Client{}
var verbose bool

func main() {
	// https://d22puzix29w08m.cloudfront.net/video/2810bb676e325412913e2d431479e8e4/7d35e7e2f52a9613/ts_audio.m3u8
	m3u8_str := flag.String("url", "", "m3u8 url")
	save_path := flag.String("save_path", "", "")
	//v := flag.Bool("v", false, "verbose")
	//user_agent := flag.String("user_agent", "AppleCoreMedia/1.0.0.15B93 (iPad; U; CPU OS 11_1 like Mac OS X; ja_jp", "")
	flag.Parse()
	if *m3u8_str == "" || *save_path == "" {
		log.Fatal("missing parameter")
	}
	verbose = true
	m3u8_url, err := url.Parse(*m3u8_str)
	if err != nil {
		log.Fatal(err)
	}

	psChan := make(chan *Playlist, 1024)
	go getPlaylist(m3u8_url, psChan)
	if err := downloadSegment(*save_path, psChan); err != nil {
		log.Fatal(err)
	}
}

func getPlaylist(m3u8_url *url.URL, psChan chan *Playlist) error {
	cache := lru.New(1024)
	recDuration := time.Duration(0)

	for {
		req, err := http.NewRequest("GET", m3u8_url.String(), nil)
		if err != nil {
			return err
		}
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
				msURL, err := m3u8_url.Parse(v.URI)
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

func downloadSegment(save_path string, psChan chan *Playlist) error {
	out, err := os.Create(save_path)
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
		if verbose {
			log.Printf("Downloaded %v\n", v.URI)
		}
	}
	return nil
}
