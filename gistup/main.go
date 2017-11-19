package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gong023/my-slack-process/slack"
)

type (
	GistReq struct {
		Files map[string]File `json:"files"`
	}

	File struct {
		Content string `json:"content"`
		RawURL  string `json:"raw_url"`
	}

	GistRes struct {
		ID    string
		Files map[string]File `json:"files"`
	}
)

func main() {
	gistToken := flag.String("gist_token", "", "gist token")
	baseDir := flag.String("base", "", "depth 1")
	webhook := flag.String("webhook", "", "slack webhook url")
	flag.Parse()
	if *gistToken == "" || *baseDir == "" || *webhook == "" {
		log.Fatal("missing parameter")
	}

	baseFile, err := ioutil.ReadDir(*baseDir)
	if err != nil {
		log.Fatal(err)
	}
	for _, base := range baseFile {
		if base.IsDir() == false {
			continue
		}

		path := *baseDir + "/" + base.Name()
		title, err := ioutil.ReadFile(path + "/title.txt")
		if err != nil {
			log.Fatal(err)
		}
		caption, err := ioutil.ReadFile(path + "/caption.txt")
		if err != nil {
			log.Fatal(err)
		}

		gres, err := createGist(*gistToken, string(title), string(caption))
		if err != nil {
			log.Fatal(err)
		}

		err = upload(gres.ID, path)
		if err != nil {
			log.Fatal(err)
		}

		gres, err = getGist(*gistToken, gres.ID)
		if err != nil {
			log.Fatal(err)
		}

		attachments := slack.Attachments{
			Attachments: []slack.Attachment{
				{
					Title:   string(title),
					Pretext: string(caption),
				},
			},
		}
		for _, file := range gres.Files {
			attachments.Attachments = append(attachments.Attachments, slack.Attachment{
				ImageURL: file.RawURL,
			})
		}
		post(*webhook, attachments)

		if err = exec.Command("rm", "-fr", path).Run(); err != nil {
			log.Fatal(err)
		}
	}
}

func createGist(token, title, caption string) (res GistRes, err error) {
	body := GistReq{
		Files: map[string]File{
			title: {
				Content: caption,
			},
		},
	}
	pg, err := json.Marshal(body)
	if err != nil {
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.github.com/gists", bytes.NewBuffer(pg))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	if err != nil {
		return
	}
	r, err := client.Do(req)
	if err != nil {
		return
	}
	defer r.Body.Close()
	if r.StatusCode >= 300 {
		return res, errors.New(fmt.Sprintf("create gist error (%s)", r.Status))
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &res)
	if err != nil {
		return
	}

	return
}

func upload(gistID, path string) (err error) {
	err = os.Chdir(path)
	if err != nil {
		return
	}

	exec.Command("rm", "-fr", ".git").Run()
	if err = exec.Command("git", "init").Run(); err != nil {
		return
	}
	remote := fmt.Sprintf("git@gist.github.com:%s.git", gistID)
	if err = exec.Command("git", "remote", "add", "origin", remote).Run(); err != nil {
		return
	}
	if err = exec.Command("git", "add", "*.txt").Run(); err != nil {
		return
	}
	if err = exec.Command("git", "add", "*.jpeg").Run(); err != nil {
		return
	}
	if err = exec.Command("git", "commit", "-m", "upload").Run(); err != nil {
		return
	}
	if err = exec.Command("git", "push", "-f", "origin", "master").Run(); err != nil {
		return
	}

	return
}

func getGist(token, gistID string) (res GistRes, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/gists/"+gistID, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	if err != nil {
		return
	}
	r, err := client.Do(req)
	if err != nil {
		return
	}
	defer r.Body.Close()
	if r.StatusCode >= 300 {
		return res, errors.New(fmt.Sprintf("get gist error (%s)", r.Status))
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &res)
	if err != nil {
		return
	}

	return
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
