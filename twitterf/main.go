package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/gong023/my-slack-process/slack"
	"log"
	"net/url"
	"time"
	"os"
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
	cli = anaconda.NewTwitterApiWithCredentials(accessToken, accessSecret, consumerKey, consumerSecret)

	mediaMin := flag.Int("media_min", 2000, "")
	mediaMax := flag.Int("media_max", 3000, "")
	tolerance := flag.Int("tolerance", 3, "for minor")
	activeTweet := flag.Duration("active_tweet", -3*24*time.Hour, "")
	flag.Parse()

	val := url.Values{}
	val.Add("screen_name", "gong023")
	ret := <-cli.GetFriendsListAll(val)
	if ret.Error != nil {
		log.Fatal(ret.Error)
	}
	favs, _ := cli.GetFavorites(val)
	for _, fav := range favs {
		ignore = append(ignore, fav.Id)
	}
	for _, friend := range ret.Friends {
		active, err := isActiveUser(*activeTweet, friend.ScreenName)
		if err != nil {
			log.Fatal(err)
		}
		if !active {
			continue
		}

		val = url.Values{}
		val.Add("screen_name", friend.ScreenName)
		favs, err := cli.GetFavorites(val)
		if err != nil {
			log.Fatal(err)
		}
		for _, fav := range favs {
			if !isTarget(*tolerance, *mediaMin, *mediaMax, fav) {
				continue
			}
			attachments := slack.Attachments{
				Attachments: []slack.Attachment{
					{
						Fallback:   fmt.Sprintf("https://twitter.com/%s/status/%d", fav.User.ScreenName, fav.Id),
						Pretext:    fmt.Sprintf("https://twitter.com/%s/status/%d", fav.User.ScreenName, fav.Id),
						AuthorName: fmt.Sprintf("%s likes %s. count:%d", friend.ScreenName, fav.User.ScreenName, fav.FavoriteCount),
						Text:       fav.FullText,
					},
				},
			}
			for _, media := range fav.Entities.Media {
				attachments.Attachments = append(attachments.Attachments, slack.Attachment{
					Title:     media.Url,
					TitleLink: media.Url,
					ImageURL:  media.Media_url_https,
				})
			}

			if b, err := json.Marshal(attachments); err == nil {
				fmt.Println(string(b))
			}
		}
	}
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
