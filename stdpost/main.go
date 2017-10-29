package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/gong023/my-slack-process/util"
)

func main() {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err.Error())
	}
	if len(bytes) == 0 {
		os.Exit(0)
	}

	webhook := flag.String("webhook", "", "slack incoming webhook url")
	flag.Parse()
	if *webhook == "" {
		log.Fatal("webhook is required")
	}

	util.PostText(*webhook, string(bytes))
}
