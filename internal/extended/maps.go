package extended

import (
	"fmt"
	sm "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	"image/color"
)

type Geo struct {
	Title    string  `json:"title"`
	Filename string  `json:"filename"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
}

func PositionMap(geo *Geo) {

	title := geo.Title
	filename := geo.Filename
	lat := geo.Lat
	lng := geo.Lng

	if title == "" {
		panic("missing title")
	}
	if filename == "" {
		panic("missing filename")
	}
	if lat == 0 {
		panic("missing lat")
	}

	if lng == 0 {
		panic("missing lng")
	}

	ctx := sm.NewContext()
	ctx.SetSize(600, 400)
	ctx.SetZoom(14)

	ctx.OverrideAttribution(title)
	ctx.AddObject(
		sm.NewMarker(
			s2.LatLngFromDegrees(lat, lng),
			color.RGBA{0xff, 0, 0, 0xff},
			16.0,
		),
	)

	img, err := ctx.Render()
	if err != nil {
		panic(err)
	}
	fileTitle := fmt.Sprintf("%s.png", filename)
	if err := gg.SavePNG(fileTitle, img); err != nil {
		panic(err)
	}
}

func areaMap(geo *Geo) {

	title := geo.Title
	filename := geo.Filename
	lat := geo.Lat
	lng := geo.Lng

	if title == "" {
		panic("missing title")
	}
	if filename == "" {
		panic("missing filename")
	}
	if lat == 0 {
		panic("missing lat")
	}

	if lng == 0 {
		panic("missing lng")
	}

	ctx := sm.NewContext()
	ctx.SetSize(600, 400)

	ctx.OverrideAttribution(title)
	ctx.SetCenter(s2.LatLngFromDegrees(lat, lng))
	img, err := ctx.Render()
	if err != nil {
		panic(err)
	}
	fileTitle := fmt.Sprintf("%s.png", filename)
	if err := gg.SavePNG(fileTitle, img); err != nil {
		panic(err)
	}
}

func MultilineMap() {
	ctx := sm.NewContext()
	ctx.SetSize(1920, 1080)

	newyork := sm.NewMarker(s2.LatLngFromDegrees(40.641766, -73.780968), color.RGBA{255, 0, 0, 255}, 16.0)
	hongkong := sm.NewMarker(s2.LatLngFromDegrees(22.308046, 113.918480), color.RGBA{0, 0, 255, 255}, 16.0)
	ctx.AddObject(newyork)
	ctx.AddObject(hongkong)
	path := make([]s2.LatLng, 0, 2)
	path = append(path, newyork.Position)
	path = append(path, hongkong.Position)
	ctx.AddObject(sm.NewPath(path, color.RGBA{0, 255, 0, 255}, 4.0))

	img, err := ctx.Render()
	if err != nil {
		panic(err)
	}

	if err := gg.SavePNG("idl.png", img); err != nil {
		panic(err)
	}
}
