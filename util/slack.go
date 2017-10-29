package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

type SlackMessage struct {
	Text string `json:"text"`
}

func PostText(webhook string, text string) (err error) {
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

func LogErr(err error) {
	os.Stderr.WriteString(err.Error() + "\n")
	os.Exit(1)
}
