package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"
)

// APIData is API Data
type APIData struct {
	DateTime string   `json:"datetime"`
	Headline []string `json:"headline"`
	Link     string   `json:"link"`
	Title    string   `json:"title"`
}

// JmardbAPI is JmardbAPI
type JmardbAPI struct {
	Data []APIData
}

func getXMLLink(sday, eday time.Time) (string, error) {
	ssday := sday.Format("2006-01-02 15:04:05")
	seday := eday.Format("2006-01-02 15:04:05")
	v := url.Values{}
	v.Set("title", "府県天気予報")
	v.Add("areacode_mete", "270000")
	v.Add("datetime", ssday)
	v.Add("datetime", seday)
	apiURL := `http://api.aitc.jp/jmardb-api/search`
	fetchURL := apiURL + "?" + v.Encode()

	log.Println("Fetch URL:", fetchURL)

	resp, err := http.Get(fetchURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var d JmardbAPI
	if err := dec.Decode(&d); err != nil {
		return "", err
	}
	return d.Data[len(d.Data)-1].Link, nil
}
