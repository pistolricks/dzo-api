package main

import (
	address2 "github.com/Boostport/address"
	"github.com/pistolricks/go-api-template/internal/extended"
	"net/http"
)

func (app *application) createAddressHandler(w http.ResponseWriter, r *http.Request) {
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

	address, error := extended.ValidateAddress(addr)

	co, coerr := extended.GetAddressOSM(address)

	headers := make(http.Header)

	err = app.writeJSON(w, http.StatusCreated, envelope{"address": address, "osm": co[0], "c-errors": coerr, "errors": error}, headers)
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
