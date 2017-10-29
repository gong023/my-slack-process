package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gong023/my-slack-process/util"
)

type (
	WRes struct {
		Weather []Weather
		Main    Main
	}

	Weather struct {
		Main        string
		Description string
	}

	Main struct {
		TempMin int `json:"temp_min"`
		TempMax int `json:"temp_max"`
	}

	SlackMessage struct {
		Text string `json:"text"`
	}
)

func main() {
	wtoken := flag.String("wtoken", "", "token for https://openweathermap.org/api")
	city := flag.String("city", "San Francisco", "city to get forecast")
	webhook := flag.String("webhook", "", "slack incoming webhook url")
	flag.Parse()
	if *wtoken == "" {
		util.LogErr(errors.New("wtoken is required"))
	}
	if *webhook == "" {
		util.LogErr(errors.New("webhook is required"))
	}

	wq := url.Values{}
	wq.Set("APPID", *wtoken)
	wq.Set("q", *city)
	wq.Set("units", "metric")
	wr, err := http.Get("http://api.openweathermap.org/data/2.5/weather?" + wq.Encode())
	if err != nil {
		util.LogErr(err)
	}
	defer wr.Body.Close()
	b, err := ioutil.ReadAll(wr.Body)
	if err != nil {
		util.LogErr(err)
	}
	var wres WRes
	err = json.Unmarshal(b, &wres)
	if err != nil {
		util.LogErr(err)
	}

	desc := wres.Weather[0].Description
	min := strconv.Itoa(wres.Main.TempMin)
	max := strconv.Itoa(wres.Main.TempMax)
	m := "(" + *city + ") " + desc + " " + min + "C/" + max + "C"

	util.PostText(*webhook, m)
}
