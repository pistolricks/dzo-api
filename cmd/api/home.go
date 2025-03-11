package main

import (
	address2 "github.com/Boostport/address"
	"net/http"
)

func (app *application) showHome(w http.ResponseWriter, r *http.Request) {

	headers := make(http.Header)

	err := app.writeJSON(w, http.StatusOK, envelope{"form": address2.GetCountry("US")}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
