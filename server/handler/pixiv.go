package handler

import (
	"io/ioutil"
	"net/http"
)

func ImageProxy(w http.ResponseWriter, r *http.Request) {
	imageURL := "https://i.pximg.net/" + r.URL.Query().Get("q")
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	req.Header.Add("Referer", "https://app-api.pixiv.net/")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	defer res.Body.Close()

	var b []byte
	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if _, err := w.Write(b); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
