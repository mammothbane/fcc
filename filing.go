package fcc

import (
	"bytes"
	"encoding/json"
	"net/url"
	"path"
	"time"
)

type (
	// FilingInfo is the basic struct for this package. Instantiate one and call Submit() to send a public filing
	// to the FCC. Call BuildECFS instead to build a request with more to tweak (not recommended).
	FilingInfo struct {
		Address
		Named
		Email string
		Text  string
	}

	Named struct {
		Name string `json:"name"`
	}

	Address struct {
		AddressFirstLine  string `json:"address_line_1"`
		AddressSecondLine string `json:"address_line_2,omitempty"`
		City              string `json:"city"`
		State             string `json:"state"`
		ZipCode           string `json:"zip_code"`
		Zip4              string `json:"zip4"`
	}

	IntlAddress struct {
		Text string `json:"addresstext"`
	}

	SubmissionType struct {
		Id           uint   `json:"id"`
		Abbreviation string `json:"abbreviation"`
		Description  string `json:"description"`
		Short        string `json:"short"`
	}

	FilingStatus struct {
		Id          uint   `json:"id"`
		Description string `json:"description"`
	}

	ViewingStatus struct {
		Id          uint   `json:"id,string"`
		Description string `json:"description"`
	}

	Document struct {
		Source      string `json:"src"`
		Filename    string `json:"filename"`
		Description string `json:"description"`
	}

	// You probably shouldn't be building one of these directly unless you know what you want to change.
	// Prefer FilingInfo.Submit() or FilingInfo.BuildECFS() instead.
	ECFSFiling struct {
		Proceedings []*Proc `json:"proceedings"`

		Filers   []*Named `json:"filers"`
		Authors  []*Named `json:"authors"`
		Bureaus  []*Named `json:"bureaus"`
		Lawfirms []*Named `json:"lawfirms"`

		Address     *Address     `json:"addressentity"`
		IntlAddress *IntlAddress `json:"internationaladdressentity"`

		Email   string `json:"contact_email"`
		Text    string `json:"text_data"`
		Express int    `json:"express_comment"`

		// no idea what these do
		FileNumber     uint            `json:"file_number,omitempty,string"`
		ReportNumber   uint            `json:"report_number,omitempty,string"`
		Entity         string          `json:"entity,omitempty"`
		SubmissionType *SubmissionType `json:"submission_type,omitempty"`
		ViewingStatus  *ViewingStatus  `json:"viewingstatus,omitempty"`
		FilingStatus   *FilingStatus   `json:"filingstatus,omitempty"`
		DateSubmitted  *time.Time      `json:"date_submission,omitempty"`
		Attachments    []interface{}   `json:"attachments,omitempty"`
		Documents      []*Document     `json:"documents"`
	}

	// FilingConfirmations are returned by the FCC API on submission of a filing.
	FilingConfirmation struct {
		Confirmation string `json:"confirm"`
		Received     string `json:"received"`
		Status       string `json:"status"`
	}
)

func Status(confirmationId string) (*ECFSFiling, *Error) {
	fileVals := url.Values{}
	fileVals.Add("api_key", c.ApiKey)

	loc, err := ecfs_root.Parse(path.Join("filings" + confirmationId))
	if err != nil {
		return nil, newErr(err, true)
	}
	loc.RawQuery = fileVals.Encode()

	resp, err := client.Get(loc.String())
	defer resp.Body.Close()
	if err != nil {
		return nil, newErr(err, true)
	}

	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return nil, strErr("Response code %v checking filing %v", false, resp.StatusCode, confirmationId)
	}

	var ecfs ECFSFiling
	err = json.NewDecoder(resp.Body).Decode(&ecfs)
	return &ecfs, newErr(err, false)
}

// Build an ECFSFiling for submission to the FCC.
func (f FilingInfo) BuildECFS(proceedings ...*Proc) (*ECFSFiling, *Error) {
	if proceedings == nil {
		return nil, strErr("Tried to build a filing with no proceedings.", true)
	}

	filing := ECFSFiling{
		Proceedings: proceedings,
		Filers: []*Named{
			&f.Named,
		},

		Address: &f.Address,
		Email:   f.Email,
		Text:    f.Text,
		Express: 1,

		IntlAddress: &IntlAddress{},
		Authors:     []*Named{},
		Lawfirms:    []*Named{},
		Bureaus:     []*Named{},
		Documents:   []*Document{},
	}

	return &filing, nil
}

// Submits an ECFSFiling to the FCC, returning the confirmation if successful.
func (e *ECFSFiling) Submit() (*FilingConfirmation, *Error) {
	for i, p := range e.Proceedings {
		e.Proceedings[i] = p.strip()
	}

	fileVals := url.Values{}
	fileVals.Add("api_key", c.ApiKey)

	loc, _ := ecfs_root.Parse("filings")
	loc.RawQuery = fileVals.Encode()

	var b bytes.Buffer

	type wrapper struct {
		Filing ECFSFiling `json:"filing"`
	}

	if err := json.NewEncoder(&b).Encode(e); err != nil {
		return nil, newErr(err, true)
	}

	resp, err := client.Post(loc.String(), "application/json", &b)
	defer resp.Body.Close()

	if err != nil {
		return nil, newErr(err, true)
	}

	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return nil, strErr("Response code %v on submission of filing %v.", false, resp.StatusCode, e)
	}

	var conf FilingConfirmation
	err = json.NewDecoder(resp.Body).Decode(&conf)

	return &conf, newErr(err, false)
}

// Submit the FilingInfo to the FCC with the given proceedings.
// Combines FilingInfo.BuildECFS() and ECFSFiling.Submit() into one call.
func (f FilingInfo) Submit(proceedings ...*Proc) (*FilingConfirmation, *Error) {
	ecfs, err := f.BuildECFS(proceedings...)
	if err != nil {
		return nil, err
	}

	return ecfs.Submit()
}
