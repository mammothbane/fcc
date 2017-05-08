package fcc

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type Proceeding struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Id          string `json:"id_proceeding"`
}

func proceeding(name string) (*Proceeding, error) {
	procVals := url.Values{}
	procVals.Add("api_key", c.ApiKey)
	procVals.Add("q", name)

	loc, _ := ECFS_ROOT.Parse("proceedings")
	loc.RawQuery = procVals.Encode()

	resp, err := client.Get(loc.String())
	if err != nil {
		return nil, err
	}

	var out struct {
		Proceedings []Proceeding `json:"proceedings"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		resp.Body.Close()
		return nil, err
	}

	if len(out.Proceedings) != 1 {
		return nil, fmt.Errorf("Invalid proceeding count: %v", len(out.Proceedings))
	}

	return &out.Proceedings[0], err
}
