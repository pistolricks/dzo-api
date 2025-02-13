package main

import (
	address2 "github.com/Boostport/address"
	"github.com/pistolricks/go-api-template/internal/extended"
	"net/http"
)

type Position struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

type XYZData struct {
	Tags []string `json:"tags"`
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

	addr := &extended.Address{
		StreetAddress:      input.StreetAddress,
		Locality:           input.Locality,
		AdministrativeArea: input.AdministrativeArea,
		PostCode:           input.PostCode,
		Country:            input.Country,
	}

	address, ev := extended.ValidateAddress(addr)

	co, coerr := extended.GetAddressOSM(address)

	err = app.writeJSON(w, http.StatusCreated, envelope{"address": address, "results": co, "results-errors": coerr, "errors": ev}, headers)
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
