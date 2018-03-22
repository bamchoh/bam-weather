package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bamchoh/bam-weather/genindex"
	"github.com/bamchoh/bam-weather/genpng"
	"github.com/bamchoh/bam-weather/mys3"
	"github.com/pkg/errors"
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

type WeatherInfo struct {
	TimeModifier string
	Weather      Weather `xml:"http://xml.kishou.go.jp/jmaxml1/elementBasis1/ Weather"`
}

type SubArea struct {
	Sentence string
}

type WeatherForecastPart struct {
	ID        string `xml:"refID,attr"`
	Sentence  string
	Base      WeatherInfo
	Temporary []WeatherInfo
	Becoming  []WeatherInfo
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
	s = strings.Replace(s, "未明", "夜おそぉに", -1)
	s = strings.Replace(s, "では", "らへんは", -1)
	s = strings.Replace(s, "所により", "どっかでは", -1)
	s = strings.Replace(s, "海上", "海のほう", -1)
	s = strings.Replace(s, "夜遅く", "夜おそぉ", -1)
	s = strings.Replace(s, "夜のはじめ頃", "会社から退社する頃", -1)
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

func (t WeatherInfo) Exists(searchText []string) bool {
	for _, text := range searchText {
		if strings.Contains(t.TimeModifier, text) {
			return true
		}
	}
	return false
}

func generateForecast(wf WeatherForecastPart, tempL, tempH string) string {
	var ws []WeatherInfo

	ws = append(ws, wf.Base)
	ws = append(ws, wf.Temporary...)
	ws = append(ws, wf.Becoming...)

	searchText := []string{"時々", "後"}
	report := ""
	for _, w := range ws {
		if w.TimeModifier != "" {
			if !w.Exists(searchText) {
				report += ModifySentence("や。")
			}
			report += ModifySentence(w.TimeModifier)
			if !w.Exists(searchText) {
				report += ModifySentence("は")
			}
		}
		report += ModifySentence(w.Weather.Text)
	}
	report += ModifySentence("や。")

	if wf.SubArea.Sentence != "" {
		report += ModifySentence("なんか")
		report += ModifySentence(wf.SubArea.Sentence)
		report += ModifySentence("らしいで")
	}

	lowest := "いっちゃん低い温度は " + tempL
	highest := "いっちゃん高い温度は " + tempH + "やで"
	tag := "#bam_weather"

	report = fmt.Sprintf("大阪の今日(%s)の天気は基本%s\n%s\n%s\n%s", time.Now().Format("1月2日"), report, lowest, highest, tag)
	return report
}

func getDayInfo(day time.Time) (*DayInfo, error) {
	link, err := getXMLLink(day)
	if err != nil {
		return nil, err
	}

	return getWeatherReport(link)
}

func genWeatherInfo(today, yesterday *DayInfo) genpng.WeatherInfo {
	bases := strings.Split(today.Weather.Base.Weather.Text, " ")

	info := genpng.WeatherInfo{
		First: bases[0],
		Low:   strings.Replace(yesterday.TempL, "度", "", -1),
		High:  strings.Replace(today.TempH, "度", "", -1),
	}

	switch {
	case len(bases) > 2:
		info.Second = bases[1]
		info.Third = bases[2]
	case len(today.Weather.Becoming) > 0:
		bec := today.Weather.Becoming[0]
		mod := bec.TimeModifier
		switch mod {
		case "後", "時々":
			info.Second = mod
		default:
			info.Second = "後"
		}
		info.Third = today.Weather.Becoming[0].Weather.Text
	}

	return info
}

func run() error {
	var err error
	logFile := os.Stdout
	if err != nil {
		return errors.Wrap(err, "failed to open log file")
	}

	log.SetOutput(logFile)

	tt := time.Now()
	// For debug
	// loc, err := time.LoadLocation("Local")
	// if err != nil {
	// 	panic(err)
	// }
	// tt := time.Date(2018, 3, 1, 7, 0, 0, 0, loc)
	today, err := getDayInfo(tt)
	if err != nil {
		return errors.Wrap(err, "failed to get today info")
	}
	yesterday, err := getDayInfo(tt.Add(-24 * time.Hour))
	if err != nil {
		return errors.Wrap(err, "failed to get yesterday info")
	}
	log.Println(today.Weather)

	info := genWeatherInfo(today, yesterday)

	bucket := "bam-weather"
	region := "ap-northeast-1"

	var buffer *bytes.Buffer
	buffer = bytes.NewBuffer(make([]byte, 0))
	err = genpng.Generate(info, buffer)
	if err != nil {
		return err
	}

	err = mys3.Upload(bucket, region, "weather.png", "binary/octet-stream", buffer)
	if err != nil {
		return err
	}

	buffer = bytes.NewBuffer(make([]byte, 0))
	err = genindex.Generate(buffer)
	if err != nil {
		return err
	}

	err = mys3.Upload(bucket, region, "index.html", "text/html", buffer)
	if err != nil {
		return err
	}

	text := generateForecast(today.Weather, yesterday.TempL, today.TempH)
	text += fmt.Sprintf("\nhttps://s3-ap-northeast-1.amazonaws.com/bam-weather/index.html?%d", time.Now().Unix())
	log.Println("Text:", text)
	tweet(text)

	return nil
}

func main() {
	lambda.Start(run)
}
