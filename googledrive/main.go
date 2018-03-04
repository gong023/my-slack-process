package main

import (
	"flag"
	"io/ioutil"
	"log"

	"fmt"

	"github.com/gong023/my-slack-process/oauth"
)

type ()

func main() {
	clientID := flag.String("client_id", "", "")
	clientSec := flag.String("client_sec", "", "")
	refreshPath := flag.String("refresh_path", "", "file refresh token stored")
	flag.Parse()
	if *refreshPath == "" || *clientID == "" || *clientSec == "" {
		log.Fatal("missing parameter")
	}

	refreshToken, err := ioutil.ReadFile(*refreshPath)
	if err != nil {
		log.Fatal(err)
	}

	req := oauth.NewRefresh("https://www.googleapis.com/oauth2/v4/token")
	tres, err := req.Refresh(*clientID, *clientSec, string(refreshToken))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(tres.AccessToken)
}
