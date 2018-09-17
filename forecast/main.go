package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
		TempMin float64 `json:"temp_min"`
		TempMax float64 `json:"temp_max"`
	}
)

func main() {
	wtoken := flag.String("wtoken", "", "token for https://openweathermap.org/api")
	city := flag.String("city", "Palo Alto", "city to get forecast")
	flag.Parse()
	if *wtoken == "" {
		log.Fatal("wtoken is required")
	}

	wq := url.Values{}
	wq.Set("APPID", *wtoken)
	wq.Set("q", *city)
	wq.Set("units", "metric")
	wr, err := http.Get("http://api.openweathermap.org/data/2.5/weather?" + wq.Encode())
	if err != nil {
		log.Fatal(err)
	}
	defer wr.Body.Close()
	b, err := ioutil.ReadAll(wr.Body)
	if err != nil {
		log.Fatal(err)
	}
	var wres WRes
	err = json.Unmarshal(b, &wres)
	if err != nil {
		log.Fatal(err)
	}

	desc := wres.Weather[0].Description
	fmt.Printf("(%s) %s %fC/%fC", *city, desc, wres.Main.TempMin, wres.Main.TempMax)
}
