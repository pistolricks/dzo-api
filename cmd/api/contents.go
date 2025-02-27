package main

import (
	"fmt"
	"github.com/pistolricks/go-api-template/internal/extended"
	"github.com/pistolricks/validation"
	"strconv"

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
	folder := app.handleEncodeHashids(user.ID, "Ollivr")

	pathway := filepath.Join("ui/static", folder)
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

	hrf := app.handleRenameFile(fp, hfp)
	if hrf != nil {
		app.badRequestResponse(w, r, hrf)
	}

	content := &extended.Content{
		Name:     hashedFileName,
		Original: originalFileName,
		Hash:     hash,
		Src:      hfp,
		Type:     r.FormValue("type"),
		Size:     nbBytes,
		Folder:   folder,
		UserID:   user.ID,
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

	err = app.writeJSON(w, http.StatusCreated, envelope{"userId": user.ID, "folder": folder, "path": fp, "type": r.FormValue("type"), "size": nbBytes, "hash": hash}, nil)
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
	f := filepath.Join(path, filename)
	dst, err := os.Create(f)
	if err != nil {
		return nil, err
	}

	return dst, nil
}

func (app *application) showOwnerContentsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Hash     string
		Name     string
		Original string
		Src      string
		Type     string
		Size     int64
		Folder   string
		UserID   int64
		extended.Filters
	}

	v := validation.New()

	qs := r.URL.Query()

	input.Original = app.readString(qs, "original", "")
	input.Type = app.readString(qs, "type", "")
	input.Hash = app.readString(qs, "hash", "")
	input.Folder = app.readString(qs, "folder", "")
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "original", "type", "folder", "-id", "-original", "-type", "-folder"}

	if extended.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	contents, metadata, err := app.extended.Contents.GetAll(input.Hash, input.Name, input.Original, input.Src, input.Type, input.Size, input.Folder, input.UserID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	item := new(extended.ContentCollection)

	item.Data = contents
	item.Metadata = metadata

	pos := Position{33.983841, -118.451424}

	err = app.writeGeoJSON(w, http.StatusOK, "profile", envelope{"contents": item}, nil, strconv.FormatInt(int64(1), 10), pos)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
