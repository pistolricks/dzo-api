package main

import (
	address2 "github.com/Boostport/address"
	geojson "github.com/paulmach/go.geojson"
	"github.com/pistolricks/go-api-template/internal/extended"
	"net/http"
	"strconv"
)

type Position struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

type XYZData struct {
	Tags []string `json:"tags"`
}

func (app *application) addressDetailsHandler(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)

	var input struct {
		PlaceID int `json:"place_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	res, errors := extended.GetDetailsWithPlaceId(input.PlaceID)

	err = app.writeJSON(w, http.StatusCreated, envelope{"query": input, "results": res, "errors": errors}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) addressDetailsByCoordinates(w http.ResponseWriter, r *http.Request) {

	headers := make(http.Header)

	var input struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	lat64, err := strconv.ParseFloat(input.Lat, 64)
	lon64, err := strconv.ParseFloat(input.Lon, 64)

	res, err := extended.GetDetailsWithCoordinates(lat64, lon64)

	pos := Position{lat64, lon64}
	geo := app.fillGeoJSON(strconv.FormatInt(int64(res.OsmID), 10), "place", pos, envelope{"place_id": strconv.FormatInt(int64(res.PlaceID), 10), "type": res.Type, "osm_type": res.OsmType, "display": res.DisplayName, "extratags": res.Extratags, "importance": res.Importance, "address": res.Address, "boundingbox": res.Boundingbox, "viewbox": ""})

	err = app.writeJSON(w, http.StatusCreated, envelope{"query": input, "results": geo, "errors": err}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) addressSearchHandler(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)

	var input struct {
		Search  string `json:"search"`
		Viewbox string `json:"viewbox"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	res, errors := extended.SearchOsm(input.Search, input.Viewbox)

	featureCollection := geojson.NewFeatureCollection()

	for key := range res {

		lat64, _ := strconv.ParseFloat(res[key].Lat, 64)
		lon64, _ := strconv.ParseFloat(res[key].Lon, 64)

		pos := Position{lat64, lon64}

		if res[key].Type != "yes" {
			geo := app.fillGeoJSON(strconv.FormatInt(int64(res[key].OsmID), 10), "place", pos, envelope{
				"place_id":        strconv.FormatInt(int64(res[key].PlaceID), 10),
				"type":            res[key].Type,
				"osm_type":        res[key].OsmType,
				"importance":      res[key].Importance,
				"address":         res[key].Address,
				"extratags":       res[key].Extratags,
				"boundingbox":     res[key].Boundingbox,
				"display_name":    res[key].DisplayName,
				"category":        res[key].Category,
				"address_type":    res[key].AddressType,
				"centroid":        res[key].Centroid,
				"addresstags":     res[key].Addresstags,
				"name":            res[key].Name,
				"parent_place_id": res[key].ParentPlaceID,
				"admin_level":     res[key].AdminLevel,
				"local_name":      res[key].Localname,
				"svg":             res[key].Svg})
			featureCollection.AddFeature(geo)
		}
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"query": input.Search, "collection": featureCollection, "errors": errors}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createAddressHandler(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)

	var input struct {
		Viewbox            string `json:"viewbox"`
		PoiKey             string `json:"poi_key"`
		PoiID              string `json:"poi_id"`
		StreetAddress      string `json:"street_address"`
		Locality           string `json:"locality"`
		AdministrativeArea string `json:"administrative_area"`
		PostCode           string `json:"post_code"`
		Country            string `json:"country"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	res, errors := extended.GetAddressOSM(input.Viewbox, input.PoiKey, input.PoiID, input.StreetAddress, input.Locality, input.AdministrativeArea, input.PostCode, input.Country)

	featureCollection := geojson.NewFeatureCollection()

	for key := range res {

		lat64, _ := strconv.ParseFloat(res[key].Lat, 64)
		lon64, _ := strconv.ParseFloat(res[key].Lon, 64)

		pos := Position{lat64, lon64}

		if res[key].Type != "yes" {
			geo := app.fillGeoJSON(strconv.FormatInt(int64(res[key].OsmID), 10), "place", pos, envelope{
				"place_id":        strconv.FormatInt(int64(res[key].PlaceID), 10),
				"type":            res[key].Type,
				"osm_type":        res[key].OsmType,
				"display_name":    res[key].DisplayName,
				"importance":      res[key].Importance,
				"address":         res[key].Address,
				"extratags":       res[key].Extratags,
				"category":        res[key].Category,
				"boundingbox":     res[key].Boundingbox,
				"svg":             res[key].Svg,
				"address_type":    res[key].AddressType,
				"centroid":        res[key].Centroid,
				"addresstags":     res[key].Addresstags,
				"name":            res[key].Name,
				"parent_place_id": res[key].ParentPlaceID,
				"admin_level":     res[key].AdminLevel,
				"local_name":      res[key].Localname,
			})
			featureCollection.AddFeature(geo)
		}

	}
	err = app.writeJSON(w, http.StatusCreated, envelope{"query": input, "collection": featureCollection, "errors": errors}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showAddressForm(w http.ResponseWriter, r *http.Request) {

	headers := make(http.Header)

	err := app.writeJSON(w, http.StatusOK, envelope{"form": address2.GetCountry("US")}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
