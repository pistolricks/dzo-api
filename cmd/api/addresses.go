package main

import (
	"github.com/pistolricks/go-api-template/internal/extended"
	"net/http"
)

func (app *application) createAdressHandler(w http.ResponseWriter, r *http.Request) {
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

	address := extended.ValidateAddress(addr)

	headers := make(http.Header)

	err = app.writeJSON(w, http.StatusCreated, envelope{"address": address}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
