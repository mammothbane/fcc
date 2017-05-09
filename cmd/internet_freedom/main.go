package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/schema"
	"github.com/mammothbane/fcc"
	"gopkg.in/go-playground/validator.v9"
)

var (
	assets     = http.FileServer(http.Dir("assets"))
	dec        = schema.NewDecoder()
	proceeding *fcc.Proc
	validate   *validator.Validate = validator.New()
)

type formContent struct {
	Name    string `form:"name" validate:"required"`
	Email   string `form:"email" validate:"required"`
	Address string `form:"address" validate:"required"`
	City    string `form:"city" validate:"required"`
	Zip     string `form:"zip" validate:"required"`
	State   string `form:"state" validate:"required"`
	Comment string `form:"comment" validate:"required"`
}

func (f formContent) toFilingInfo() *fcc.FilingInfo {
	return &fcc.FilingInfo{
		Named: fcc.Named{Name: f.Name},
		Address: fcc.Address{
			AddressFirstLine: f.Address,
			City:             f.City,
			ZipCode:          f.Zip,
			State:            f.State,
		},
		Text:  f.Comment,
		Email: f.Email,
	}
}

const PROCEEDING = "17-108"

func decodeInfo(r *http.Request) (*fcc.FilingInfo, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	var content formContent
	if err := dec.Decode(&content, r.Form); err != nil {
		return nil, err
	}

	if err := validate.Struct(content); err != nil {
		return nil, err
	}

	return content.toFilingInfo(), nil
}

func handleSubmit(w http.ResponseWriter, r *http.Request) {
	log.Println("got submit")

	bail := func(err error, failcode int, msg string, args ...interface{}) bool {
		if err != nil {
			log.Printf(msg, append(args, err)...)
			w.WriteHeader(failcode)
			return true
		}
		return false
	}

	info, err := decodeInfo(r)
	if bail(err, http.StatusBadRequest, "Unable to decode filing information: %v") {
		return
	}

	bck := &backoff{
		factor:  2,
		current: 10 * time.Millisecond,
		max:     10 * time.Second,
	}

	var conf *fcc.FilingConfirmation

	log.Println("submitting filing info...")
	err = bck.do(func() *fcc.Error {
		inner, err := info.Submit(proceeding)
		if err == nil {
			conf = inner
		} else if !err.Fatal() {
			log.Printf("Error: '%v'. Retrying.", err)
		}

		return err
	})

	if bail(err, http.StatusInternalServerError, "Failed submitting filing info: %v") {
		return
	}

	log.Println("Successfully submitted! Confirmation:", conf.Confirmation)
}

func main() {

	bck := &backoff{
		factor:  2,
		current: 10 * time.Millisecond,
		max:     10 * time.Second,
	}

	log.Println("retrieving proceeding...")
	err := bck.do(func() *fcc.Error {
		inner, err := fcc.Proceeding(PROCEEDING)
		if err == nil {
			proceeding = inner
		} else if !err.Fatal() {
			log.Printf("Error: '%v'. Retrying.", err)
		}

		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("got proceeding", PROCEEDING)

	http.Handle("/", assets)
	http.HandleFunc("/submit", handleSubmit)

	log.Println("starting server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type backoff struct {
	factor  float64
	current time.Duration
	max     time.Duration
}

func (b *backoff) do(op func() *fcc.Error) *fcc.Error {
	for {
		err := op()
		if err == nil || err.Fatal() {
			return err
		}

		b.current = time.Duration(float64(b.current) * b.factor)
		if b.current > b.max {
			return err
		}

		time.Sleep(b.current)
	}
}
