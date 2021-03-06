package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gong023/my-slack-process/pixiv"
	"github.com/gong023/my-slack-process/slack"
	"log"
	"os"
	"time"
)

func main() {
	host := os.Getenv("PROXY_HOST")
	if host == "" {
		log.Fatal("PROXY_HOST is not given")
	}

	s := flag.Duration("since", 20*time.Minute, "get since")
	flag.Parse()

	token, err := pixiv.GetToken()
	if err != nil {
		log.Fatal(err)
	}

	illusts, err := pixiv.GetFollowingIllusts(token)
	if err != nil {
		log.Fatal(err)
	}

	since := time.Now().Add(*(s) * -1)
	for _, illust := range illusts.Illusts {
		create, err := time.Parse(time.RFC3339, illust.CreateDate)
		if err != nil {
			log.Fatal(err)
		}
		if create.Before(since) {
			continue
		}

		attachments := slack.Attachments{
			Attachments: []slack.Attachment{
				{
					Title:   illust.Title,
					Pretext: illust.Caption,
					Text:    fmt.Sprintf("https://www.pixiv.net/member_illust.php?mode=medium&illust_id=%d", illust.ID),
				},
			},
		}
		if len(illust.MetaPages) <= 0 {
			attachments.Attachments = append(attachments.Attachments, slack.Attachment{
				ImageURL: host + "?" + pixiv.ImagePath(illust.ImageURLs.Large),
			})
		} else {
			for _, metaPage := range illust.MetaPages {
				attachments.Attachments = append(attachments.Attachments, slack.Attachment{
					ImageURL: host + "?" + pixiv.ImagePath(metaPage.Large),
				})
			}
		}

		b, err := json.Marshal(attachments)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	}
}
