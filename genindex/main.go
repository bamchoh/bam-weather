package genindex

import (
	"fmt"
	"html/template"
	"io"
	"time"
)

func today() string {
	wdays := []string{"日", "月", "火", "水", "木", "金", "土"}
	now := time.Now()
	return fmt.Sprintf("%d月%d日(%s)", now.Month(), now.Day(), wdays[now.Weekday()])
}

func Generate(f io.Writer) error {
	const html = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta property="og:title" content="大阪の天気 {{ .Today }}" />
    <meta property="og:type" content="article" />
    <meta property="og:url" content="https://s3-ap-northeast-1.amazonaws.com/bam-weather/index.html" />
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
		Today:  today(),
		Serial: time.Now().Unix(),
	})
	return err
}
