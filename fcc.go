// An API to the FCC's ECFS system.
package fcc

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

type conf struct {
	ApiKey string `json:"api_key"`
}

type Error struct {
	error
	fatal bool
}

func (e *Error) Fatal() bool {
	if e == nil {
		return false
	}
	return e.fatal
}

func newErr(e error, fatal bool) *Error {
	if e == nil {
		return nil
	}

	return &Error{
		error: e,
		fatal: fatal,
	}
}

func strErr(s string, fatal bool, args ...interface{}) *Error {
	return &Error{
		error: fmt.Errorf(s, args...),
		fatal: fatal,
	}
}

var (
	ecfs_root, _ = url.Parse("https://publicapi.fcc.gov/ecfs/")
	c            *conf
	client              = http.DefaultClient
)

func init() {
	f, err := os.Open("conf.json")
	if err != nil {
		log.Fatal(err)
	}

	if err := json.NewDecoder(f).Decode(&c); err != nil {
		log.Fatal(err)
	}
}
