package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Control struct {
	Title            string `xml:"Title"`
	DateTime         string `xml:"DateTime"`
	Status           string `xml:"Status"`
	EditorialOffice  string `xml:"EditorialOffice"`
	PublishingOffice string `xml:"PublishingOffice"`
}

type Head struct {
	Title           string
	ReportDateTime  string
	TargetDateTime  string
	TargetDuration  string
	EventID         string
	InfoType        string
	Serial          string
	InfoKind        string
	InfoKindVersion string
}

type Weather struct {
	Type string `xml:"type,attr"`
	Text string `xml:",chardata"`
}

type Base struct {
	Weather Weather `xml:"http://xml.kishou.go.jp/jmaxml1/elementBasis1/ Weather"`
}

type Temporary struct {
	TimeModifier string
	Weather      Weather `xml:"http://xml.kishou.go.jp/jmaxml1/elementBasis1/ Weather"`
}

type SubArea struct {
	Sentence string
}

type WeatherForecastPart struct {
	ID        string `xml:"refID,attr"`
	Sentence  string
	Base      Base
	Temporary Temporary
	SubArea   SubArea
}

type Temperature struct {
	Temp        string `xml:",chardata"`
	Description string `xml:"description,attr"`
	ID          string `xml:"refID,attr"`
	Type        string `xml:"type,attr"`
}

type TemperaturePart struct {
	Temperature Temperature `xml:"http://xml.kishou.go.jp/jmaxml1/elementBasis1/ Temperature"`
}

type Property struct {
	Type             string
	WeatherForecasts []WeatherForecastPart `xml:"DetailForecast>WeatherForecastPart"`
	TemperaturePart  TemperaturePart       `xml:"TemperaturePart"`
}

type Item struct {
	Kinds []Property `xml:"Kind>Property"`
}

type TimeDefine struct {
	ID   string `xml:"timeId,attr"`
	Name string `xml:"Name"`
}

type MeteorologicalInfo struct {
	Type        string       `xml:"type,attr"`
	TimeDefines []TimeDefine `xml:"TimeSeriesInfo>TimeDefines>TimeDefine"`
	Items       []Item       `xml:"TimeSeriesInfo>Item"`
}

type Body struct {
	MeteorologicalInfos []MeteorologicalInfo `xml:"MeteorologicalInfos"`
}

type Report struct {
	Control Control
	Head    Head
	Body    Body
}

func ModifySentence(s string) string {
	s = strings.Replace(s, "　", " ", -1)
	s = strings.Replace(s, "後", "からの", -1)
	s = strings.Replace(s, "一時", "ちょっとのま", -1)
	s = strings.Replace(s, "を伴う", "もある", -1)
	s = strings.Replace(s, "時々", "たま～に", -1)
	s = strings.Replace(s, "を伴い", "もあるし", -1)
	s = strings.Replace(s, "非常に", "めっちゃ", -1)
	s = strings.Replace(s, "激しく", "ぎょーさん", -1)
	s = strings.Replace(s, "山地", "山のほう", -1)
	s = strings.Replace(s, "未明", "夜おそーに", -1)
	s = strings.Replace(s, "では", "らへんは", -1)
	s = strings.Replace(s, "所により", "どっかでは", -1)
	s = strings.Replace(s, "海上", "海のほう", -1)
	s = strings.Replace(s, "夜遅く", "夜遅くに", -1)
	s = strings.Replace(s, "夜のはじめ頃", "夜に", -1)
	s = strings.Replace(s, "晴れ", "\xE2\x98\x80", -1)
	s = strings.Replace(s, "雨", "\xE2\x98\x94", -1)
	s = strings.Replace(s, "雪", "\xE2\x9B\x84", -1)
	s = strings.Replace(s, "くもり", "\xE2\x98\x81", -1)
	s = strings.Replace(s, "雷", "\xE2\x9A\xA1", -1)
	return s
}

type DayInfo struct {
	Weather WeatherForecastPart
	TempL   string
	TempH   string
}

func getWeatherReport(path string) (*DayInfo, error) {
	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dec := xml.NewDecoder(resp.Body)
	var v Report
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}

	var di DayInfo
	if len(v.Body.MeteorologicalInfos) > 0 {
		info := v.Body.MeteorologicalInfos[0]
		if len(info.Items) > 0 {
			item := info.Items[0]
			if len(item.Kinds) > 0 {
				kind := item.Kinds[0]
				if len(kind.WeatherForecasts) > 0 {
					for _, w := range kind.WeatherForecasts {
						if w.ID == "1" {
							di.Weather = w
							break
						}
					}
				}
			}
		}
	}

	if len(v.Body.MeteorologicalInfos) > 0 {
		for _, info := range v.Body.MeteorologicalInfos {
			if info.Type == "地点予報" {
				for _, def := range info.TimeDefines {
					switch def.Name {
					case "明日朝":
						id, err := strconv.Atoi(def.ID)
						if err != nil {
							return nil, err
						}
						di.TempL = info.Items[0].Kinds[id-1].TemperaturePart.Temperature.Description
					case "今日日中":
						id, err := strconv.Atoi(def.ID)
						if err != nil {
							return nil, err
						}
						di.TempH = info.Items[0].Kinds[id-1].TemperaturePart.Temperature.Description
					}
				}
				break
			}
		}
	}
	return &di, nil
}

func generateForecast(wf WeatherForecastPart, tempL, tempH string) string {
	report := ""
	report += ModifySentence(wf.Base.Weather.Text)
	report += ModifySentence("や。")
	if wf.Temporary.TimeModifier != "" {
		report += ModifySentence(wf.Temporary.TimeModifier)
		if !strings.Contains(wf.Temporary.TimeModifier, "時々") {
			report += ModifySentence("は")
		}
		report += ModifySentence(wf.Temporary.Weather.Text)
		report += ModifySentence("や。")
	}
	if wf.SubArea.Sentence != "" {
		report += ModifySentence("なんか")
		report += ModifySentence(wf.SubArea.Sentence)
		report += ModifySentence("らしいで")
	}

	lowest := "いっちゃん低い温度は " + tempL
	highest := "いっちゃん高い温度は " + tempH + "やで"
	report = fmt.Sprintf("大阪の今日の天気は基本%s\n%s\n%s", report, lowest, highest)
	return report
}

func getDayInfo(day time.Duration) (*DayInfo, error) {
	link, err := getXMLLink(day)
	if err != nil {
		return nil, err
	}

	return getWeatherReport(link)
}

func main() {
	today, err := getDayInfo(time.Duration(0))
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	yesterday, err := getDayInfo(time.Duration(-1))
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	tweet(generateForecast(today.Weather, yesterday.TempL, today.TempH))
}
