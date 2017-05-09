package fcc

import (
	"encoding/json"
	"net/url"
	"time"
)

type Proc struct {
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	Id                 uint       `json:"id_proceeding,string"`
	DisplayDescription string     `json:"description_display,omitempty"`
	Index              string     `json:"_index,omitempty"`
	BureauCode         string     `json:"bureau.code,omitempty"`
	BureauName         string     `json:"bureau.name,omitempty"`
	Filer              string     `json:"filed_by,omitempty"`
	ApplicantName      string     `json:"applicant_name,omitempty"`
	Created            *time.Time `json:"date_proceeding_created,omitempty"`
	Closed             *time.Time `json:"date_closed,omitempty"`
}

func (p *Proc) strip() *Proc {
	return &Proc{
		Name:        p.Name,
		Description: p.Description,
		Id:          p.Id,
	}
}

func Proceeding(name string) (*Proc, *Error) {
	procVals := url.Values{}
	procVals.Add("api_key", c.ApiKey)
	procVals.Add("name", name)

	loc, _ := ecfs_root.Parse("proceedings")
	loc.RawQuery = procVals.Encode()

	resp, err := client.Get(loc.String())
	defer resp.Body.Close()
	if err != nil {
		return nil, newErr(err, true)
	}

	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return nil, strErr("Response code %v on retrieval of proceedings %v", false, resp.StatusCode, name)
	}

	var out struct {
		Proceedings []Proc `json:"proceedings"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, newErr(err, false)
	}

	if len(out.Proceedings) != 1 {
		return nil, strErr("Got the wrong number (%v) of proceedings back for name '%s'", true, len(out.Proceedings), name)
	}

	return &out.Proceedings[0], nil
}
