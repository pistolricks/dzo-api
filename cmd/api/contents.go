package main

import (
	"fmt"
	"github.com/pistolricks/go-api-template/internal/extended"
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

	fileExt := filepath.Ext(handler.Filename)
	originalFileName := strings.TrimSuffix(filepath.Base(handler.Filename), filepath.Ext(handler.Filename))
	now := time.Now()
	filename := strings.ReplaceAll(strings.ToLower(originalFileName), " ", "-") + "-" + fmt.Sprintf("%v", now.Unix()) + fileExt

	dst, err := app.createFile(w, r, filename)
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
	hash := extended.HashImage(filepath.Join("uploads", filename))

	err = app.writeJSON(w, http.StatusCreated, envelope{"path": filepath.Join("uploads", filename), "type": r.FormValue("type"), "size": nbBytes, "hash": hash}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createFile(w http.ResponseWriter, r *http.Request, filename string) (*os.File, error) {
	// Create an uploads directory if it doesn’t exist
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		err := os.Mkdir("uploads", 0755)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}
	}

	// Build the file path and create it
	dst, err := os.Create(filepath.Join("uploads", filename))
	if err != nil {
		return nil, err
	}

	return dst, nil
}
