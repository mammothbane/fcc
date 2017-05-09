// An API to the FCC's ECFS system.
package fcc

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
)

type conf struct {
	ApiKey string `json:"api_key"`
}

var (
	ecfs_root, _  = url.Parse("https://publicapi.fcc.gov/ecfs/")
	c            *conf
	client       = http.DefaultClient
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
