package main

import (
	"fmt"
	"github.com/pistolricks/go-api-template/internal/extended"
	"github.com/pistolricks/validation"

	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (app *application) uploadImageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	err := r.ParseMultipartForm(10)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			app.badRequestResponse(w, r, err)
		}
	}(file)

	fmt.Println("Past readJSON")
	user := app.contextGetUser(r)

	id := app.handleEncodeHashids(user.ID, "Ollivr-Contents")

	pathway := filepath.Join("ui/static", id)
	fileExt := filepath.Ext(handler.Filename)
	originalFileName := strings.TrimSuffix(filepath.Base(handler.Filename), filepath.Ext(handler.Filename))
	now := time.Now()
	filename := strings.ReplaceAll(strings.ToLower(originalFileName), " ", "-") + "-" + fmt.Sprintf("%v", now.Unix()) + fileExt

	dst, err := app.createFile(w, r, pathway, filename)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer func(dst *os.File) {
		err := dst.Close()
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}
	}(dst)

	nbBytes, _ := io.Copy(dst, file)
	fp := filepath.Join(pathway, filename)
	hash := extended.HashImage(fp)

	hashedFileName := hash + fileExt
	hfp := filepath.Join(pathway, hashedFileName)

	herr := app.handleRenameFile(fp, hfp)
	if herr != nil {
		app.badRequestResponse(w, r, herr)
	}

	content := &extended.Content{
		Name:     hashedFileName,
		Original: originalFileName,
		Hash:     hash,
		Src:      hfp,
		Type:     r.FormValue("type"),
		Size:     nbBytes,
		UserID:   id,
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

	err = app.writeJSON(w, http.StatusCreated, envelope{"userId": id, "path": fp, "type": r.FormValue("type"), "size": nbBytes, "hash": hash}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createFile(w http.ResponseWriter, r *http.Request, path string, filename string) (*os.File, error) {
	// Create an static directory if it doesn’t exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0755)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}
	}

	// Build the file path and create it
	dst, err := os.Create(filepath.Join(path, filename))
	if err != nil {
		return nil, err
	}

	return dst, nil
}

func (app *application) listContentsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Hash     string
		Name     string
		Original string
		Src      string
		Type     string
		Size     int64
		UserID   string
		extended.Filters
	}

	v := validation.New()

	qs := r.URL.Query()

	input.Original = app.readString(qs, "original", "")
	input.Type = app.readString(qs, "type", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "original", "type", "-id", "-original", "-type"}

	if extended.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	contents, metadata, err := app.extended.Contents.GetAll(input.Hash, input.Name, input.Original, input.Src, input.Type, input.Size, input.UserID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"contents": contents, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
