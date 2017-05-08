package fcc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
)

type (
	Filer struct {
		Name string `json:"name"`
	}

	FilingInfo struct {
		Address
		Filer
		Email string
		Text  string
	}

	Address struct {
		FirstLine  string `json:"address_line_1"`
		SecondLine string `json:"address_line_2,omitempty"`
		City       string `json:"city"`
		ZipCode    string `json:"zip_code"`
		Zip4       string `json:"zip4"`
	}

	ECFSFiling struct {
		Proceedings []*Proceeding `json:"proceedings"`

		Filers []Filer `json:"filers"`

		Address `json:"addressentity"`

		Email   string `json:"contact_email"`
		Text    string `json:"text_data"`
		Express int    `json:"express_comment"`
	}
)

func (f FilingInfo) submit(proceeding *Proceeding) error {
	filing := ECFSFiling{
		Proceedings: []*Proceeding{
			proceeding,
		},
		Filers: []Filer{
			f.Filer,
		},

		Address: f.Address,
		Email:   f.Email,
		Text:    f.Text,
		Express: 1,
	}

	fileVals := url.Values{}
	fileVals.Add("api_key", c.ApiKey)

	loc, _ := ECFS_ROOT.Parse("filings")
	loc.RawQuery = fileVals.Encode()

	var (
		b   bytes.Buffer
		wg  sync.WaitGroup
		err error
	)

	wg.Add(1)

	go func() {
		err = json.NewEncoder(b).Encode(&filing)
		wg.Done()
	}()

	resp, httpErr := http.Post(loc.String(), "application/json", b)
	wg.Wait()

	if err != nil {
		return err
	}

}
