package main

import (
	"errors"
	"fmt"
	"github.com/pistolricks/go-api-template/internal/extended"
	"github.com/pistolricks/validation"
	"net/http"
)

func (app *application) createVendorHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string           `json:"title"`
		Year    int32            `json:"year"`
		Runtime extended.Runtime `json:"runtime"`
		Genres  []string         `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	vendor := &extended.Vendor{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validation.New()

	if extended.ValidateVendor(v, vendor); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.extended.Vendors.Insert(vendor)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/vendors/%d", vendor.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"vendor": vendor}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showVendorHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	vendor, err := app.extended.Vendors.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, extended.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"vendor": vendor}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateVendorHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	vendor, err := app.extended.Vendors.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, extended.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title   *string           `json:"title"`
		Year    *int32            `json:"year"`
		Runtime *extended.Runtime `json:"runtime"`
		Genres  []string          `json:"genres"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		vendor.Title = *input.Title
	}

	if input.Year != nil {
		vendor.Year = *input.Year
	}
	if input.Runtime != nil {
		vendor.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		vendor.Genres = input.Genres
	}

	v := validation.New()

	if extended.ValidateVendor(v, vendor); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.extended.Vendors.Update(vendor)
	if err != nil {
		switch {
		case errors.Is(err, extended.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"vendor": vendor}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteVendorHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.extended.Vendors.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, extended.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "vendor successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listVendorsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		extended.Filters
	}

	v := validation.New()

	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	if extended.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	vendors, metadata, err := app.extended.Vendors.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"vendors": vendors, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
