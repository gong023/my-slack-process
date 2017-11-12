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
)

type (
	SlackMessage struct {
		Text string `json:"text"`
	}
)

func main() {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	if len(bytes) == 0 {
		os.Exit(0)
	}

	webhook := flag.String("webhook", "", "slack incoming webhook url")
	flag.Parse()
	if *webhook == "" {
		log.Fatal("webhook is required")
	}

	postText(*webhook, string(bytes))
}

func postText(webhook string, text string) (err error) {
	sm, err := json.Marshal(SlackMessage{Text: text})

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
