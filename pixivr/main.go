package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gong023/my-slack-process/pixiv"
	"github.com/gong023/my-slack-process/slack"
)

func main() {
	host := os.Getenv("PROXY_HOST")
	if host == "" {
		log.Fatal("PROXY_HOST is not given")
	}

	limit := flag.Int("limit", 3, "rankings up to x")
	flag.Parse()

	token, err := pixiv.GetToken()
	if err != nil {
		log.Fatal(err)
	}

	illusts, err := pixiv.GetDailyRankingIllusts(token)
	if err != nil {
		log.Fatal(err)
	}

	for i, illust := range illusts.Illusts {
		if i > *limit {
			break
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
