package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type RegularLXml struct {
	Entries []Entry `xml:"entry"`
}

type Entry struct {
	Title   string `xml:"title"`
	Link    Link   `xml:"link"`
	Author  string `xml:"author>name"`
	Updated string `xml:"updated"`
}

type Link struct {
	URL string `xml:"href,attr"`
}

func getXMLLink2(sday, eday time.Time) (string, error) {
	fetchURL := "http://www.data.jma.go.jp/developer/xml/feed/regular_l.xml"

	log.Println("Fetch URL:", fetchURL)

	resp, err := http.Get(fetchURL)
	if err != nil {
		return "", fmt.Errorf("getXMLLink2:fetch error:%v", err)
	}
	defer resp.Body.Close()

	dec := xml.NewDecoder(resp.Body)
	var r RegularLXml
	if err := dec.Decode(&r); err != nil {
		return "", fmt.Errorf("getXMLLink2:decode error:%v", err)
	}

	for _, entry := range r.Entries {
		if entry.Title == "府県天気予報" && entry.Author == "大阪管区気象台" {
			tt, err := time.Parse("2006-01-02T15:04:05Z", entry.Updated)
			if err != nil {
				return "", fmt.Errorf("getXMLLink2:parse error:%v", err)
			}
			if tt.Before(eday) && tt.After(sday) {
				return entry.Link.URL, nil
			}
		}
	}
	return "", errors.New("getXMLLink2:link was not found")
}
