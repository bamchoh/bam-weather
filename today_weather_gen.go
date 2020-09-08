package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/bamchoh/bam-weather/genpng"
	"github.com/pkg/errors"
)

type TodayWeatherGenerator struct {
	BaseTime    time.Time
	text        string
	weatherInfo genpng.WeatherInfo
}

func (gen *TodayWeatherGenerator) Init() error {
	tt := gen.BaseTime
	today, err := gen.getDayInfo(tt.Add(-6*time.Hour), tt)
	if err != nil {
		err = errors.Wrap(err, "failed to get today info")
		log.Println(err)
		return err
	}

	tt2 := tt.Add(-24 * time.Hour)
	yesterday, err := gen.getDayInfo(tt2.Add(-6*time.Hour), tt2)
	if err != nil {
		err = errors.Wrap(err, "failed to get yesterday info")
		log.Println(err)
		return err
	}

	log.Println(today.Weather)

	when := fmt.Sprintf("今日(%s)", gen.Day().Format("1月2日"))
	gen.text = generateForecast(today.Weather, yesterday.TempL, today.TempH, when)
	gen.text += fmt.Sprintf("\n%v?%d", indexURL, time.Now().Unix())

	gen.weatherInfo = genWeatherInfo(today, yesterday.TempL, today.TempH)
	return nil
}

func (gen *TodayWeatherGenerator) getDayInfo(sday, eday time.Time) (*DayInfo, error) {
	link, err := getXMLLink2(sday, eday)
	if err != nil {
		return nil, err
	}

	return gen.getWeatherReport(link)
}

func (gen *TodayWeatherGenerator) getWeatherReport(path string) (*DayInfo, error) {
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

func (gen *TodayWeatherGenerator) Text() string {
	return gen.text
}

func (gen *TodayWeatherGenerator) WeatherInfo() genpng.WeatherInfo {
	return gen.weatherInfo
}

func (gen *TodayWeatherGenerator) Day() time.Time {
	return gen.BaseTime
}
