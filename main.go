package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bamchoh/bam-weather/genindex"
	"github.com/bamchoh/bam-weather/genpng"
	"github.com/bamchoh/bam-weather/mys3"
	"github.com/pkg/errors"
)

var (
	indexURL = "https://s3-ap-northeast-1.amazonaws.com/bam-weather/index.html"
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

func (t WeatherInfo) Exists(searchText []string) bool {
	for _, text := range searchText {
		if strings.Contains(t.TimeModifier, text) {
			return true
		}
	}
	return false
}

func genWeatherInfo(day *DayInfo, templ, temph string) genpng.WeatherInfo {
	bases := strings.Split(day.Weather.Base.Weather.Text, " ")

	info := genpng.WeatherInfo{
		First: bases[0],
		Low:   strings.Replace(templ, "度", "", -1),
		High:  strings.Replace(temph, "度", "", -1),
	}

	switch {
	case len(bases) > 2:
		info.Second = bases[1]
		info.Third = bases[2]
	case len(day.Weather.Becoming) > 0:
		bec := day.Weather.Becoming[0]
		mod := bec.TimeModifier
		switch mod {
		case "後", "時々":
			info.Second = mod
		default:
			info.Second = "後"
		}
		info.Third = day.Weather.Becoming[0].Weather.Text
	}

	return info
}

func generateForecast(wf WeatherForecastPart, tempL, tempH, when string) string {
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

	report = fmt.Sprintf("大阪の%sの天気は基本%s\n%s\n%s\n%s", when, report, lowest, highest, tag)
	return report
}

type WeatherGenerator interface {
	Init() error
	Text() string
	WeatherInfo() genpng.WeatherInfo
	Day() time.Time
}

type SpecificTime struct {
	Specify bool `json:"specify"`
	Day     int  `json:"day"`
	Hour    int  `json:"hour"`
}

func run(event SpecificTime) error {
	var err error
	logFile := os.Stdout
	if err != nil {
		err = errors.Wrap(err, "failed to open log file")
		log.Println(err)
		return err
	}

	log.SetOutput(logFile)

	tt := time.Now()
	if event.Specify {
		loc, err := time.LoadLocation("Local")
		if err != nil {
			log.Println(err)
			return err
		}
		tt = time.Date(
			tt.Year(),
			tt.Month(),
			event.Day,
			event.Hour,
			0,
			0,
			0,
			loc)
	}

	var gen WeatherGenerator
	if tt.Hour() >= 18 {
		log.Println("Tomorrow")
		gen = &TomorrowWeatherGenerator{
			BaseTime: tt,
		}
	} else {
		log.Println("Today")
		gen = &TodayWeatherGenerator{
			BaseTime: tt,
		}
	}

	err = gen.Init()
	if err != nil {
		log.Println(err)
		return err
	}
	info := gen.WeatherInfo()

	bucket := "bam-weather"
	region := "ap-northeast-1"

	var buffer *bytes.Buffer
	buffer = bytes.NewBuffer(make([]byte, 0))
	err = genpng.Generate(info, buffer)
	if err != nil {
		log.Println(err)
		return err
	}

	err = mys3.Upload(bucket, region, "weather.png", "binary/octet-stream", buffer)
	if err != nil {
		log.Println(err)
		return err
	}

	buffer = bytes.NewBuffer(make([]byte, 0))
	err = genindex.Generate(buffer, gen.Day(), tt.Unix())
	if err != nil {
		log.Println(err)
		return err
	}

	err = mys3.Upload(bucket, region, "index.html", "text/html", buffer)
	if err != nil {
		log.Println(err)
		return err
	}

	text := gen.Text()
	log.Println("Text:", text)
	tweet(text)

	return nil
}

func main() {
	lambda.Start(run)
}
