package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/gong023/my-slack-process/slack"
)

var (
	cli    *anaconda.TwitterApi
	ignore []int64
)

func main() {
	accessToken := os.Getenv("TWITTER_ACCESS")
	if accessToken == "" {
		log.Fatal("accessToken is empty")
	}
	accessSecret := os.Getenv("TWITTER_ASEC")
	if accessSecret == "" {
		log.Fatal("accessToken secret is empty")
	}
	consumerKey := os.Getenv("TWITTER_CKEY")
	if consumerKey == "" {
		log.Fatal("consumerKey is empty")
	}
	consumerSecret := os.Getenv("TWITTER_CSEC")
	if consumerSecret == "" {
		log.Fatal("consumerKey secret is empty")
	}
	webhook := os.Getenv("WEBHOOK")
	if webhook == "" {
		log.Fatal("webhook is empty")
	}
	cli = anaconda.NewTwitterApiWithCredentials(accessToken, accessSecret, consumerKey, consumerSecret)

	mediaMin := flag.Int("media_min", 300, "")
	mediaMax := flag.Int("media_max", 3000, "")
	tolerance := flag.Int("tolerance", 3, "for minor")
	activeTweet := flag.Duration("active_tweet", -3*24*time.Hour, "")
	interval := flag.Duration("interval", 15*time.Minute, "")
	flag.Parse()

	out, err := process(*mediaMin, *mediaMax, *tolerance, *activeTweet)
	if err != nil {
		out = []slack.Attachment{
			{
				Text: err.Error(),
			},
		}
	}
	if err := postSlack(webhook, slack.Attachments{Attachments: out}); err != nil {
		log.Println(err)
	}

	tick := time.Tick(*interval)
	for {
		select {
		case <-tick:
			out, err := process(*mediaMin, *mediaMax, *tolerance, *activeTweet)
			if err != nil {
				out = []slack.Attachment{
					{
						Text: err.Error(),
					},
				}
			}
			if err := postSlack(webhook, slack.Attachments{Attachments: out}); err != nil {
				log.Println(err)
			}
		}
	}
}

func process(mediaMin, mediaMax, tolerance int, activeTweet time.Duration) (attachments []slack.Attachment, err error) {
	val := url.Values{}
	val.Add("screen_name", "gong023")
	ret := <-cli.GetFriendsListAll(val)
	if ret.Error != nil {
		return attachments, ret.Error
	}
	favs, _ := cli.GetFavorites(val)
	for _, fav := range favs {
		ignore = append(ignore, fav.Id)
	}
	for _, friend := range ret.Friends {
		active, err := isActiveUser(activeTweet, friend.ScreenName)
		if err != nil {
			return attachments, err
		}
		if !active {
			continue
		}

		val = url.Values{}
		val.Add("screen_name", friend.ScreenName)
		favs, err := cli.GetFavorites(val)
		if err != nil {
			return attachments, err
		}
		for _, fav := range favs {
			if !isTarget(tolerance, mediaMin, mediaMax, fav) {
				continue
			}
			attachments = append(attachments, slack.Attachment{
				Fallback:   fmt.Sprintf("https://twitter.com/%s/status/%d", fav.User.ScreenName, fav.Id),
				Pretext:    fmt.Sprintf("https://twitter.com/%s/status/%d", fav.User.ScreenName, fav.Id),
				AuthorName: fmt.Sprintf("%s likes %s. count:%d", friend.ScreenName, fav.User.ScreenName, fav.FavoriteCount),
				Text:       fav.FullText,
			})
			for _, media := range fav.Entities.Media {
				attachments = append(attachments, slack.Attachment{
					Title:     media.Url,
					TitleLink: media.Url,
					ImageURL:  media.Media_url_https,
				})
			}
		}
	}
	return attachments, nil
}

func isActiveUser(activeTweet time.Duration, screenName string) (bool, error) {
	val := url.Values{}
	val.Add("screen_name", screenName)

	tweets, err := cli.GetUserTimeline(val)
	if err != nil || len(tweets) == 0 {
		return false, nil
	}
	latest, err := time.Parse(time.RubyDate, tweets[0].CreatedAt)
	if err != nil {
		return false, err
	}

	criteria := time.Now().Add(activeTweet)
	return latest.After(criteria), nil
}

func isTarget(tolerance, mediaMin, mediaMax int, favTweet anaconda.Tweet) bool {
	for _, id := range ignore {
		if id == favTweet.Id {
			return false
		}
	}
	c := favTweet.FavoriteCount
	//if c < tolerance {
	//	ignore = append(ignore, favTweet.Id)
	//	return true
	//}
	if len(favTweet.Entities.Media) != 0 && mediaMin < c && c < mediaMax {
		ignore = append(ignore, favTweet.Id)
		return true
	}
	return false
}

func postSlack(webhook string, attachment slack.Attachments) (err error) {
	sm, err := json.Marshal(attachment)

	if err != nil {
		return
	}

	res, err := http.Post(webhook, "application/json", bytes.NewBuffer(sm))

	if err != nil {
		return
	}

	if res.StatusCode >= 300 {
		return errors.New("fail to post slack:" + res.Status)
	}

	return
}
