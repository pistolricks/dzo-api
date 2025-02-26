package main

import (
	"fmt"
	sm "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	"github.com/pistolricks/go-api-template/internal/extended"
	"github.com/pistolricks/validation"
	"image/color"
	"net/http"
	"path/filepath"
	"strconv"
)

func (app *application) positionMapHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title    string `json:"title"`
		Filename string `json:"filename"`
		Lat      string `json:"lat"`
		Lng      string `json:"lng"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	lat64, err := strconv.ParseFloat(input.Lat, 64)
	lng64, err := strconv.ParseFloat(input.Lng, 64)

	geo := &extended.Geo{
		Title:    input.Title,
		Filename: input.Filename,
		Lat:      lat64,
		Lng:      lng64,
	}
	/*
		app.background(func() {
			extended.PositionMap(geo)
		})
	*/
	user := app.contextGetUser(r)
	folder := app.handleEncodeHashids(user.ID, "Ollivr")

	pathway := filepath.Join("ui/static", folder)

	// var filename = filepath.Base(geo.Filename)
	var filename = geo.Filename
	ctx := sm.NewContext()
	ctx.SetSize(1200, 1200)
	ctx.SetZoom(14)

	ctx.OverrideAttribution(geo.Title)
	ctx.AddObject(
		sm.NewMarker(
			s2.LatLngFromDegrees(geo.Lat, geo.Lng),
			color.RGBA{0xff, 0, 0, 0xff},
			16.0,
		),
	)

	img, err := ctx.Render()
	if err != nil {
		panic(err)
	}

	fileTitle := fmt.Sprintf("%s.png", filename)
	f := filepath.Join(pathway, fileTitle)
	if err := gg.SavePNG(f, img); err != nil {
		panic(err)
	}

	hash := extended.HashImage(f)
	hashedFileName := fmt.Sprintf("%s.png", hash)

	hfp := filepath.Join(pathway, hashedFileName)

	herr := app.handleRenameFile(f, hfp)
	if herr != nil {
		app.badRequestResponse(w, r, herr)
	}

	content := &extended.Content{
		Name:     hashedFileName,
		Original: geo.Filename,
		Hash:     hash,
		Src:      hfp,
		Type:     "image/png",
		Size:     600,
		Folder:   folder,
		UserID:   user.ID,
	}

	v := validation.New()

	if extended.ValidateContent(v, content); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.extended.Contents.Insert(content)
	if err != nil {
		app.duplicateErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"userId": user.ID, "folder": folder, "path": pathway, "type": r.FormValue("type"), "size": 600, "hash": hash}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
