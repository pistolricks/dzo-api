package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	geojson "github.com/paulmach/go.geojson"
	"github.com/pistolricks/validation"
	"github.com/speps/go-hashids/v2"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

type envelope map[string]any

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")

	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func (app *application) fillGeoJSON(id string, profileName string, position Position, data envelope) *geojson.Feature {

	feature := geojson.NewPointFeature([]float64{position.Longitude, position.Latitude})
	feature.SetProperty(profileName, data)
	feature.ID = id
	return feature
}

func NewGeoJSON(position Position, data envelope, tags []string) ([]byte, error) {
	featureCollection := geojson.NewFeatureCollection()
	feature := geojson.NewPointFeature([]float64{position.Longitude, position.Latitude})
	feature.SetProperty("@ns:com:here:xyz", XYZData{Tags: tags})
	featureCollection.AddFeature(feature)
	return featureCollection.MarshalJSON()
}

func (app *application) writeGeoJSON(w http.ResponseWriter, status int, profileName string, data envelope, headers http.Header, id string, position Position) error {

	feature := geojson.NewPointFeature([]float64{position.Longitude, position.Latitude})
	feature.SetProperty(profileName, data)
	feature.ID = id

	js, err := feature.MarshalJSON()

	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %q", fieldName)
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d MB", maxBytesError.Limit)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {

	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	return s
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {

	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}
	return strings.Split(csv, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validation.Validator) int {

	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer")
		return defaultValue
	}
	return i
}

func (app *application) background(fn func()) {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()

		fn()
	}()
}

func HashID(id int64) string {
	idString := strconv.FormatInt(id, 10)
	hasher := sha256.New()

	_, err := hasher.Write([]byte(idString))
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	hashedID := hex.EncodeToString(hasher.Sum(nil))

	return hashedID
}

func (app *application) handleFileRequest(w http.ResponseWriter, r *http.Request) {
	fileBytes, err := os.ReadFile("read.txt")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(fileBytes)
	return
}

func (app *application) handleRenameFile(o string, n string) error {
	OriginalPath := o
	NewPath := n
	e := os.Rename(OriginalPath, NewPath)
	if e != nil {
		log.Fatal(e)
	}
	return nil
}

func (app *application) handleEncodeHashids(id int64, salt string) string {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = 10
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode([]int{int(id)})
	return e
}

func (app *application) handleDecodeHashids(id string, salt string) int64 {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = 10
	h, _ := hashids.NewWithData(hd)
	d, _ := h.DecodeWithError(id)

	return int64(d[0])
}

func (app *application) handleHashString(text string) uint32 {
	algorithm := fnv.New32a()
	algorithm.Write([]byte(text))
	return algorithm.Sum32()
}
