@echo off
set GOOS=linux
set GOARCH=amd64
go-assets-builder --output=assets/bindata.go -p=assets assets/AmeChanPopMaruTTFLight-Regular.ttf assets/rain.png assets/snow.png assets/sun.png assets/thunder.png assets/cloud.png
go build -o bam-weather
build-lambda-zip -o bam-weather.zip bam-weather
