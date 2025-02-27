package main

import (
	"errors"
	"fmt"
	"github.com/pistolricks/go-api-template/internal/extended"
	"github.com/pistolricks/validation"
	"net/http"
	"strconv"
)

func (app *application) createProfileHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Profile string         `json:"profile"`
		Name    string         `json:"name"`
		Data    extended.Attrs `json:"data"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	hash := app.handleHashString(input.Profile + "_" + input.Name)

	item := new(extended.Profile)
	item.Data = extended.Attrs{
		"properties": input.Data,
	}

	profile := &extended.Profile{
		Profile: input.Profile,
		Name:    input.Name,
		Hash:    strconv.Itoa(int(hash)),
		Data:    item.Data,
	}

	err = app.extended.Profiles.Insert(profile)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/profiles/%d", profile.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"profile": profile}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showProfileHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Hash string `json:"hash"`
	}

	profile, err := app.extended.Profiles.Get(input.Hash)
	if err != nil {
		switch {
		case errors.Is(err, extended.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	pos := Position{33.983841, -118.451424}

	err = app.writeGeoJSON(w, http.StatusOK, "profile", envelope{"profile": profile}, nil, profile.ID, pos)

	// err = app.writeJSON(w, http.StatusOK, envelope{"profile": profile}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteProfileHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Hash string `json:"hash"`
	}

	err := app.extended.Profiles.Delete(input.Hash)
	if err != nil {
		switch {
		case errors.Is(err, extended.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "profile successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listProfilesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Profile string
		Name    string
		Hash    string
		Data    extended.Attrs
		extended.Filters
	}

	v := validation.New()

	qs := r.URL.Query()

	input.Profile = app.readString(qs, "profile", "")
	input.Name = app.readString(qs, "name", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	if extended.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	profiles, metadata, err := app.extended.Profiles.GetAll(input.Profile, input.Name, input.Hash, input.Data, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"profiles": profiles, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
