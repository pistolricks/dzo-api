package main

import (
	address2 "github.com/Boostport/address"
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

	err = app.writeJSON(w, http.StatusCreated, envelope{"query": input, "results": res, "errors": err}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) addressSearchHandler(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)

	var input struct {
		Search string `json:"search"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	res, errors := extended.SearchOsm(input.Search)
	/*
		for key, value := range res {

		}
	*/

	err = app.writeJSON(w, http.StatusCreated, envelope{"query": input.Search, "results": res, "errors": errors}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createAddressHandler(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)

	var input struct {
		StreetAddress      []string `json:"street_address"`
		Locality           string   `json:"locality"`
		AdministrativeArea string   `json:"administrative_area"`
		PostCode           string   `json:"post_code"`
		Country            string   `json:"country"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	/*
		addr := &extended.Address{
			StreetAddress:      input.StreetAddress,
			Locality:           input.Locality,
			AdministrativeArea: input.AdministrativeArea,
			PostCode:           input.PostCode,
			Country:            input.Country,
		}

		 address, ev := extended.ValidateAddress(addr)
	*/
	res, errors := extended.GetAddressOSM()

	err = app.writeJSON(w, http.StatusCreated, envelope{"results": res, "errors": errors}, headers)
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
