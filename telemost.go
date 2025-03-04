package telemost

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"slices"

	"github.com/essentialkaos/ek/v13/req"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	ROOM_LEVEL_PUBLIC  = "PUBLIC"
	ROOM_LEVEL_ORG     = "ORGANIZATION"
	ROOM_LEVEL_ADMINS  = "ADMINS"
	ROOM_LEVEL_UNKNOWN = "UNKNOWN"
)

const (
	ACCESS_LEVEL_PUBLIC  = "PUBLIC"
	ACCESS_LEVEL_ORG     = "ORGANIZATION"
	ACCESS_LEVEL_UNKNOWN = "UNKNOWN"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Client is Yandex.Telemost API client
type Client struct {
	engine *req.Engine
	token  string
}

// Conference contains basic info about conference
type Conference struct {
	WaitingRoomLevel string      `json:"waiting_room_level,omitempty"`
	LiveStream       *LiveStream `json:"live_stream,omitempty"`
	CoHosts          Hosts       `json:"cohosts,omitempty"`
}

// ConferenceInfo contains information about an existing conference
type ConferenceInfo struct {
	Conference
	ID             string `json:"id"`
	JoinURL        string `json:"join_url"`
	SIPURIMeeting  string `json:"sip_uri_meeting"`
	SIPURITelemost string `json:"sip_uri_telemost"`
	SIPID          string `json:"sip_id"`
}

// LiveStream contains info about conference stream
type LiveStream struct {
	WatchURL    string `json:"watch_url"`
	AccessLevel string `json:"access_level"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Host contains info about host
type Host struct {
	Email string `json:"email"`
}

// Hosts is a slice with hosts
type Hosts []*Host

// ////////////////////////////////////////////////////////////////////////////////// //

// apiError contains info
type apiError struct {
	Code        string `json:"error"`
	Description string `json:"description"`
}

// ////////////////////////////////////////////////////////////////////////////////// //

// API is URL of Yandex.Telemost API
var API = "https://cloud-api.yandex.net/v1/telemost-api/conferences"

var (
	ErrEmptyToken    = fmt.Errorf("Token is empty")
	ErrEmptyID       = fmt.Errorf("Conference ID is empty")
	ErrEmptyCohosts  = fmt.Errorf("Cohosts slice is empty")
	ErrNilClient     = fmt.Errorf("Client is nil")
	ErrNilConference = fmt.Errorf("Conference is nil")
)

// ////////////////////////////////////////////////////////////////////////////////// //

// NewClient creates new client instance
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, ErrEmptyToken
	}

	c := &Client{engine: &req.Engine{}, token: token}
	c.SetUserAgent("", "")

	return c, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// SetUserAgent sets client user agent
func (c *Client) SetUserAgent(app, version string) {
	if c == nil || c.engine == nil {
		return
	}

	if app == "" || version == "" {
		c.engine.SetUserAgent("EK|Telemost.go", "1")
	} else {
		c.engine.SetUserAgent(app, version, "EK|Telemost.go/1")
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Create creates new conference or broadcast
//
// https://yandex.ru/dev/telemost/doc/ru/conference-create
func (c *Client) Create(conf *Conference) (*ConferenceInfo, error) {
	switch {
	case c == nil || c.engine == nil:
		return nil, ErrNilClient
	case conf == nil:
		return nil, ErrNilConference
	}

	err := validateConference(conf)

	if err != nil {
		return nil, err
	}

	info := &ConferenceInfo{}
	err = c.sendRequest(req.POST, "", info, conf, nil)

	if err != nil {
		return nil, err
	}

	return info, nil
}

// Get fetches info about conference or broadcast
//
// https://yandex.ru/dev/telemost/doc/ru/conference-read
func (c *Client) Get(id string) (*ConferenceInfo, error) {
	switch {
	case c == nil || c.engine == nil:
		return nil, ErrNilClient
	case id == "":
		return nil, ErrEmptyID
	}

	info := &ConferenceInfo{}
	err := c.sendRequest(req.GET, "/"+id, info, nil, nil)

	if err != nil {
		return nil, err
	}

	return info, nil
}

// Update updates conference or broadcast
//
// https://yandex.ru/dev/telemost/doc/ru/conference-update
func (c *Client) Update(id string, conf *Conference) (*ConferenceInfo, error) {
	switch {
	case c == nil || c.engine == nil:
		return nil, ErrNilClient
	case id == "":
		return nil, ErrEmptyID
	case conf == nil:
		return nil, ErrNilConference
	}

	err := validateConference(conf)

	if err != nil {
		return nil, err
	}

	info := &ConferenceInfo{}
	err = c.sendRequest(req.PATCH, "/"+id, info, conf, nil)

	if err != nil {
		return nil, err
	}

	return info, nil
}

// Delete cancels conference or broadcast
func (c *Client) Delete(id string) error {
	switch {
	case c == nil || c.engine == nil:
		return ErrNilClient
	case id == "":
		return ErrEmptyID
	}

	return c.sendRequest(req.DELETE, "/"+id, nil, nil, nil)
}

// GetCohosts fetches slice with all cohosts
//
// https://yandex.ru/dev/telemost/doc/ru/cohosts-read
func (c *Client) GetCohosts(id string) (Hosts, error) {
	switch {
	case c == nil || c.engine == nil:
		return nil, ErrNilClient
	case id == "":
		return nil, ErrEmptyID
	}

	resp := &struct {
		Cohosts Hosts `json:"cohosts"`
	}{}

	err := c.sendRequest(
		req.GET, "/"+id+"/cohosts", resp, nil,
		req.Query{"offset": 0, "limit": 256},
	)

	if err != nil {
		return nil, err
	}

	return resp.Cohosts, nil
}

// AddCohosts appends given hosts to conference cohosts
//
// https://yandex.ru/dev/telemost/doc/ru/cohosts-add
func (c *Client) AddCohosts(id string, cohosts []string) error {
	switch {
	case c == nil || c.engine == nil:
		return ErrNilClient
	case id == "":
		return ErrEmptyID
	case len(cohosts) == 0:
		return ErrEmptyCohosts
	}

	payload := &struct {
		Cohosts Hosts `json:"cohosts"`
	}{
		Cohosts: convertHosts(cohosts),
	}

	return c.sendRequest(req.PATCH, "/"+id+"/cohosts", nil, payload, nil)
}

// UpdateCohosts updates conference cohosts
//
// https://yandex.ru/dev/telemost/doc/ru/cohosts-update
func (c *Client) UpdateCohosts(id string, cohosts []string) error {
	switch {
	case c == nil || c.engine == nil:
		return ErrNilClient
	case id == "":
		return ErrEmptyID
	case len(cohosts) == 0:
		return ErrEmptyCohosts
	}

	payload := &struct {
		Cohosts Hosts `json:"cohosts"`
	}{
		Cohosts: convertHosts(cohosts),
	}

	return c.sendRequest(req.PUT, "/"+id+"/cohosts", nil, payload, nil)
}

// DeleteCohosts removes given hosts from chosts of conference
//
// https://yandex.ru/dev/telemost/doc/ru/cohosts-del
func (c *Client) DeleteCohosts(id string, cohosts []string) error {
	switch {
	case c == nil || c.engine == nil:
		return ErrNilClient
	case id == "":
		return ErrEmptyID
	case len(cohosts) == 0:
		return ErrEmptyCohosts
	}

	return c.sendRequest(
		req.DELETE, "/"+id+"/cohosts", nil, nil,
		req.Query{"cohost_emails": cohosts},
	)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Flatten converts slice of hosts to slice with emails
func (h Hosts) Flatten() []string {
	var result []string

	for _, hh := range h {
		result = append(result, hh.Email)
	}

	return result
}

// ////////////////////////////////////////////////////////////////////////////////// //

// sendRequest sends request to API
func (c *Client) sendRequest(method, endpoint string, response, payload any, query req.Query) error {
	r := req.Request{
		Method:  method,
		URL:     API + endpoint,
		Query:   query,
		Accept:  req.CONTENT_TYPE_JSON,
		Headers: req.Headers{"Authorization": "OAuth " + c.token},
	}

	if payload != nil {
		r.ContentType = req.CONTENT_TYPE_JSON
		r.Body = payload
	}

	resp, err := c.engine.Do(r)

	if err != nil {
		return fmt.Errorf("Can't send request to API: %w", err)
	}

	if resp.StatusCode > 299 {
		apiErr := &apiError{}
		err = resp.JSON(apiErr)

		if err == nil {
			return fmt.Errorf("API returned error: %s (%s)", apiErr.Description, apiErr.Code)
		} else {
			return fmt.Errorf("API returned non-ok status code %d", resp.StatusCode)
		}
	}

	if response != nil {
		err = resp.JSON(response)

		if err != nil {
			return fmt.Errorf("Can't decode API response: %w", err)
		}
	}

	return nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// validateConference validates conference settings
func validateConference(conf *Conference) error {
	switch {
	case conf.WaitingRoomLevel != "" &&
		!slices.Contains([]string{ROOM_LEVEL_PUBLIC, ROOM_LEVEL_ORG, ROOM_LEVEL_ADMINS, ROOM_LEVEL_UNKNOWN}, conf.WaitingRoomLevel):
		return fmt.Errorf("Unknown waiting room level %q", conf.WaitingRoomLevel)

	case conf.LiveStream != nil && conf.LiveStream.AccessLevel != "" &&
		!slices.Contains([]string{ACCESS_LEVEL_PUBLIC, ACCESS_LEVEL_ORG, ACCESS_LEVEL_UNKNOWN}, conf.LiveStream.AccessLevel):
		return fmt.Errorf("Unknown live stream access level %q", conf.LiveStream.AccessLevel)

	case conf.LiveStream != nil && len(conf.LiveStream.Title) > 1024:
		return fmt.Errorf("Live stream title exceeds maximum length (%d > 1024)", len(conf.LiveStream.Title))

	case conf.LiveStream != nil && len(conf.LiveStream.Description) > 2048:
		return fmt.Errorf("Live stream description exceeds maximum length (%d > 2048)", len(conf.LiveStream.Description))

	case len(conf.CoHosts) > 30:
		return fmt.Errorf("Too many cohosts (%d > 30)", len(conf.CoHosts))
	}

	return nil
}

// convertHosts converts slice with emails to hosts
func convertHosts(hosts []string) Hosts {
	var result Hosts

	for _, h := range hosts {
		result = append(result, &Host{h})
	}

	return result
}
