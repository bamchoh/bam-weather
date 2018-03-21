package genpng

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/golang/freetype"
	"github.com/nfnt/resize"

	"github.com/bamchoh/bam-weather/assets"
)

func resizeImg(filename string, size uint) (*image.Image, error) {
	src, err := assets.Assets.Open(filename)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		return nil, err
	}

	img = resize.Resize(0, size, img, resize.Lanczos3)
	return &img, nil
}

func generateTemp(m draw.Image, x, y int, low, high string) (err error) {
	var size float64 = 36
	file, err := assets.Assets.Open("/assets/AmeChanPopMaruTTFLight-Regular.ttf")
	if err != nil {
		log.Println(err)
		return
	}

	// Read the font data.
	var fontBytes []byte
	fontBytes, err = ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		return
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetClip(m.Bounds())
	c.SetDst(m)
	c.SetHinting(font.HintingNone)
	c.SetFontSize(size)
	var next fixed.Point26_6

	pt := freetype.Pt(x, y+int(c.PointToFixed(size)>>6))
	c.SetSrc(image.NewUniform(color.RGBA{255, 255, 255, 255}))
	next, err = c.DrawString("H:", pt)
	if err != nil {
		return
	}

	pt.X = next.X
	c.SetSrc(image.NewUniform(color.RGBA{255, 0, 0, 255}))
	next, err = c.DrawString(fmt.Sprintf("%s°", high), pt)
	if err != nil {
		return
	}

	pt.X = next.X + fixed.I(10)
	c.SetSrc(image.NewUniform(color.RGBA{255, 255, 255, 255}))
	next, err = c.DrawString("L:", pt)
	if err != nil {
		return
	}

	pt.X = next.X
	c.SetSrc(image.NewUniform(color.RGBA{0, 0, 255, 255}))
	next, err = c.DrawString(fmt.Sprintf("%s°", low), pt)
	if err != nil {
		return
	}

	return
}

func generateWeather(wType string, m draw.Image, x, y int) (next fixed.Point26_6, err error) {
	var filename = ""
	switch wType {
	case "雨":
		filename = "/assets/rain.png"
	case "晴れ":
		filename = "/assets/sun.png"
	case "雪":
		filename = "/assets/snow.png"
	case "くもり":
		filename = "/assets/cloud.png"
	case "雷":
		filename = "/assets/thunder.png"
	default:
		err = fmt.Errorf("weather type (%v) is not supported", wType)
		return
	}
	img, err := resizeImg(filename, 100)
	if err != nil {
		return
	}

	dp := image.Pt(x, y)
	r := image.Rectangle{dp, dp.Add((*img).Bounds().Size())}
	draw.Draw(m, r, *img, image.ZP, draw.Over)
	next = fixed.P(r.Max.X, r.Max.Y)
	return
}

func generateConnection(text string, rgba color.RGBA, m draw.Image, x, y int) (next fixed.Point26_6, err error) {
	var size float64 = 24
	file, err := assets.Assets.Open("/assets/AmeChanPopMaruTTFLight-Regular.ttf")
	if err != nil {
		log.Println(err)
		return
	}

	// Read the font data.
	var fontBytes []byte
	fontBytes, err = ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		return
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetClip(m.Bounds())
	c.SetDst(m)
	c.SetHinting(font.HintingNone)
	c.SetFontSize(size)

	pt := freetype.Pt(x, y+int(c.PointToFixed(size)>>6))
	c.SetSrc(image.NewUniform(rgba))
	return c.DrawString(text, pt)
}

type WeatherInfo struct {
	First  string
	Second string
	Third  string
	Low    string
	High   string
}

func generateWeatherImage(info WeatherInfo, m draw.Image, x, y int) (err error) {
	var next fixed.Point26_6
	next, err = generateWeather(info.First, m, x, y)

	if err != nil {
		return err
	}

	if info.Second != "" {
		rgba := color.RGBA{255, 255, 255, 255}
		next, err = generateConnection(info.Second, rgba, m, next.X.Ceil(), y+50)
		if err != nil {
			return err
		}

		next, err = generateWeather(info.Third, m, next.X.Ceil(), y)
		if err != nil {
			return err
		}
	}

	return generateTemp(m, x, next.Y.Ceil(), info.Low, info.High)
}

func Generate(info WeatherInfo, buffer io.Writer) error {
	var err error
	w := 300
	h := 175
	x := 0
	y := 0
	m := image.NewRGBA(image.Rect(x, y, w, h))

	bg := image.NewUniform(color.RGBA{0, 200, 255, 255})
	draw.Draw(m, m.Bounds(), bg, image.ZP, draw.Src)

	err = generateWeatherImage(info, m, x+25, y+20)
	if err != nil {
		return err
	}

	return png.Encode(buffer, m)
}
