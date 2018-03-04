@echo off
set GOOS=linux
set GOARCH=amd64
go build -o bam-weather
build-lambda-zip -o bam-weather.zip bam-weather
