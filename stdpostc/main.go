package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gong023/my-slack-process/slack"
)

func main() {
	msgs := flag.String("messages", "", "messages")
	webhook := flag.String("webhook", "", "slack incoming webhook url")
	flag.Parse()
	if *msgs == "" || *webhook == "" {
		log.Fatal("missing parameter")
	}

	f, err := os.Open(*msgs)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range strings.Split(string(b), "\n") {
		if msg == "" {
			continue
		}
		attachments := slack.Attachments{}
		err := json.Unmarshal([]byte(msg), &attachments)
		if err != nil {
			attachments = slack.Attachments{
				Attachments: []slack.Attachment{
					{
						Text: msg,
					},
				},
			}
		}
		if err := post(*webhook, attachments); err != nil {
			log.Fatal(err)
		}
	}
}

func post(webhook string, attachment slack.Attachments) (err error) {
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
