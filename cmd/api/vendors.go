package main

import (
	"errors"
	"fmt"
	"github.com/pistolricks/go-api-template/internal/extended"
	"github.com/pistolricks/validation"
	"net/http"
	"strconv"
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
	user := app.contextGetUser(r)
	err = app.extended.Vendors.AddOwner(user.ID, vendor.ID)
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

	uv, err := app.extended.Owners.GetVendorUserIds(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	user, err := app.extended.Users.GetUser(uv.UserID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	owner := app.contextGetUser(r)

	item := new(extended.VendorCollection)
	output := append(item.Data, vendor)
	metadata := extended.Metadata{
		CurrentPage:  1,
		PageSize:     1,
		FirstPage:    1,
		LastPage:     1,
		TotalRecords: 1,
	}

	item.Data = output
	item.Metadata = metadata

	var profileName = ""
	env := envelope{}

	if user.ID == owner.ID {
		profileName = "profile"
		env = envelope{"vendors": item}
	} else {
		profileName = "vendors"
		env = envelope{"data": item.Data, "metadata": item.Metadata}

	}

	pos := Position{33.983841, -118.451424}

	err = app.writeGeoJSON(w, http.StatusOK, profileName, env, nil, strconv.FormatInt(vendor.ID, 10), pos)

	// err = app.writeJSON(w, http.StatusOK, envelope{"vendor": vendor}, nil)
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

	pos := Position{33.983841, -118.451424}

	err = app.writeGeoJSON(w, http.StatusOK, "vendors", envelope{"data": vendors, "metadata": metadata}, nil, strconv.FormatInt(int64(100), 10), pos)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
