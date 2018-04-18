package genindex

import (
	"fmt"
	"html/template"
	"io"
	"time"
)

func dayString(day time.Time) string {
	wdays := []string{"日", "月", "火", "水", "木", "金", "土"}
	return fmt.Sprintf("%d月%d日(%s)", day.Month(), day.Day(), wdays[day.Weekday()])
}

func Generate(f io.Writer, day time.Time, serial int64) error {
	const html = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta property="og:title" content="大阪の天気 {{ .Today }}" />
    <meta property="og:type" content="article" />
    <meta property="og:url" content="https://s3-ap-northeast-1.amazonaws.com/bam-weather/index.html?{{ .Serial }}" />
    <meta property="og:image" content="https://s3-ap-northeast-1.amazonaws.com/bam-weather/weather.png?{{ .Serial }}" />
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:site" content="@bamchoh">
    <title>大阪の天気 {{ .Today }}</title>
  </head>
  <body>
    <img src="https://s3-ap-northeast-1.amazonaws.com/bam-weather/weather.png" />
  </body>
</html>
`

	t := template.Must(template.New("html").Parse(html))

	err := t.Execute(f, struct {
		Today  string
		Serial int64
	}{
		Today:  dayString(day),
		Serial: serial,
	})
	return err
}
