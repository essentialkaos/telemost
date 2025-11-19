package telemost

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"net/http"
	"strings"
	"testing"
	"time"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const TEST_PORT = "56123"

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type TelemostSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&TelemostSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *TelemostSuite) SetUpSuite(c *C) {
	API = "http://127.0.0.1:" + TEST_PORT

	mux := http.NewServeMux()
	server := &http.Server{Addr: ":" + TEST_PORT, Handler: mux}

	mux.HandleFunc("POST /", handlerCreateConference)
	mux.HandleFunc("GET /12345678901234", handlerGetConference)
	mux.HandleFunc("PATCH /12345678901234", handlerUpdateConference)
	mux.HandleFunc("DELETE /12345678901234", handlerDeleteConference)
	mux.HandleFunc("GET /12345678901234/cohosts", handlerGetCohosts)
	mux.HandleFunc("PATCH /12345678901234/cohosts", handlerAddCohosts)
	mux.HandleFunc("PUT /12345678901234/cohosts", handlerUpdateCohosts)
	mux.HandleFunc("DELETE /12345678901234/cohosts", handlerDeleteCohosts)

	go server.ListenAndServe()

	time.Sleep(time.Second)
}

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *TelemostSuite) TestConferenceValidation(c *C) {
	conf := &Conference{WaitingRoomLevel: "TEST"}
	c.Assert(validateConference(conf).Error(), Equals, `Unknown waiting room level "TEST"`)

	conf = &Conference{LiveStream: &LiveStream{AccessLevel: "TEST"}}
	c.Assert(validateConference(conf).Error(), Equals, `Unknown live stream access level "TEST"`)

	conf = &Conference{LiveStream: &LiveStream{AccessLevel: "PUBLIC", Title: strings.Repeat("TEST1234", 180)}}
	c.Assert(validateConference(conf).Error(), Equals, `Live stream title exceeds maximum length (1440 > 1024)`)

	conf = &Conference{LiveStream: &LiveStream{AccessLevel: "PUBLIC", Description: strings.Repeat("TEST1234", 300)}}
	c.Assert(validateConference(conf).Error(), Equals, `Live stream description exceeds maximum length (2400 > 2048)`)

	var hosts Hosts

	for range 50 {
		hosts = append(hosts, &Host{`test@test.com`})
	}

	conf = &Conference{CoHosts: hosts}
	c.Assert(validateConference(conf).Error(), Equals, `Too many cohosts (50 > 30)`)
}

func (s *TelemostSuite) TestGet(c *C) {
	api, _ := NewClient("Test1234")
	info, err := api.Get("12345678901234")

	api.SetUserAgent("Test", "1.2.3")

	c.Assert(err, IsNil)
	c.Assert(info, NotNil)

	c.Assert(info.ID, Equals, "12345678901234")
	c.Assert(info.WaitingRoomLevel, Equals, "ORGANIZATION")
	c.Assert(info.LiveStream.WatchURL, Equals, "https://telemost.yandex.ru/live/123456789abcdef0123456789abcdef0")
	c.Assert(info.LiveStream.AccessLevel, Equals, "PUBLIC")
	c.Assert(info.LiveStream.Title, Equals, "Example conference created via API")
	c.Assert(info.LiveStream.Description, Equals, "Some description of example conference created via API")
	c.Assert(info.JoinURL, Equals, "https://telemost.yandex.ru/j/12345678901234")
	c.Assert(info.SIPURIMeeting, Equals, "12345678901234567890@sip.t.ya.ru")
	c.Assert(info.SIPURITelemost, Equals, "j@sip.t.ya.ru")
	c.Assert(info.SIPID, Equals, "12345678901234567890")
}

func (s *TelemostSuite) TestCreate(c *C) {
	conf := &Conference{}

	api, _ := NewClient("Test1234")
	info, err := api.Create(conf.WithCohosts("user@domain.com"))

	c.Assert(err, IsNil)
	c.Assert(info, NotNil)

	c.Assert(info.ID, Equals, "12345678901234")
	c.Assert(info.JoinURL, Equals, "https://telemost.yandex.ru/j/12345678901234")
	c.Assert(info.LiveStream.WatchURL, Equals, "https://telemost.yandex.ru/live/123456789abcdef0123456789abcdef0")

	_, err = api.Create(&Conference{WaitingRoomLevel: "TEST"})
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, `Unknown waiting room level "TEST"`)

	api, _ = NewClient("http-error")
	_, err = api.Create(&Conference{})

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, `API returned error: Conference not found. (ConferenceNotFound)`)
}

func (s *TelemostSuite) TestUpdate(c *C) {
	api, _ := NewClient("Test1234")
	info, err := api.Update("12345678901234", &Conference{LiveStream: &LiveStream{
		AccessLevel: "PUBLIC",
		Title:       "Example conference created via API",
		Description: "Some description of example conference created via API",
	}})

	c.Assert(err, IsNil)
	c.Assert(info, NotNil)

	c.Assert(info.LiveStream.AccessLevel, Equals, "PUBLIC")
	c.Assert(info.LiveStream.Title, Equals, "Example conference created via API")
	c.Assert(info.LiveStream.Description, Equals, "Some description of example conference created via API")

	_, err = api.Update("12345678901234", &Conference{WaitingRoomLevel: "TEST"})
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, `Unknown waiting room level "TEST"`)

	api, _ = NewClient("http-error")
	_, err = api.Update("12345678901234", &Conference{})

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, `API returned error: Conference not found. (ConferenceNotFound)`)

}

func (s *TelemostSuite) TestDelete(c *C) {
	api, _ := NewClient("Test1234")
	err := api.Delete("12345678901234")

	c.Assert(err, IsNil)
}

func (s *TelemostSuite) TestGetCohosts(c *C) {
	api, _ := NewClient("Test1234")
	cohosts, err := api.GetCohosts("12345678901234")

	c.Assert(err, IsNil)
	c.Assert(cohosts, NotNil)
	c.Assert(cohosts, HasLen, 2)

	c.Assert(cohosts.Flatten(), DeepEquals, []string{"user1@yandex.ru", "user2@org-domain.ru"})

	api, _ = NewClient("http-error")
	_, err = api.GetCohosts("12345678901234")

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, `API returned error: Conference not found. (ConferenceNotFound)`)
}

func (s *TelemostSuite) TestAddCohosts(c *C) {
	api, _ := NewClient("Test1234")
	err := api.AddCohosts("12345678901234", []string{"user1@yandex.ru"})

	c.Assert(err, IsNil)
}

func (s *TelemostSuite) TestUpdateCohosts(c *C) {
	api, _ := NewClient("Test1234")
	err := api.UpdateCohosts("12345678901234", []string{"user1@yandex.ru"})

	c.Assert(err, IsNil)
}

func (s *TelemostSuite) TestDeleteCohosts(c *C) {
	api, _ := NewClient("Test1234")
	err := api.DeleteCohosts("12345678901234", []string{"user1@yandex.ru"})

	c.Assert(err, IsNil)
}

func (s *TelemostSuite) TestErrors(c *C) {
	var api *Client

	api.SetUserAgent("", "")

	_, err := api.Create(&Conference{})
	c.Assert(err, Equals, ErrNilClient)

	_, err = api.Get("1234")
	c.Assert(err, Equals, ErrNilClient)

	_, err = api.Update("1234", &Conference{})
	c.Assert(err, Equals, ErrNilClient)

	err = api.Delete("1234")
	c.Assert(err, Equals, ErrNilClient)

	_, err = api.GetCohosts("1234")
	c.Assert(err, Equals, ErrNilClient)

	err = api.AddCohosts("1234", []string{"user1@yandex.ru"})
	c.Assert(err, Equals, ErrNilClient)

	err = api.UpdateCohosts("1234", []string{"user1@yandex.ru"})
	c.Assert(err, Equals, ErrNilClient)

	err = api.DeleteCohosts("1234", []string{"user1@yandex.ru"})
	c.Assert(err, Equals, ErrNilClient)

	api, err = NewClient("")
	c.Assert(err, Equals, ErrEmptyToken)

	api, _ = NewClient("Test1234")

	_, err = api.Create(nil)
	c.Assert(err, Equals, ErrNilConference)

	_, err = api.Get("")
	c.Assert(err, Equals, ErrEmptyID)

	_, err = api.Update("", &Conference{})
	c.Assert(err, Equals, ErrEmptyID)
	_, err = api.Update("1234", nil)
	c.Assert(err, Equals, ErrNilConference)

	err = api.Delete("")
	c.Assert(err, Equals, ErrEmptyID)

	_, err = api.GetCohosts("")
	c.Assert(err, Equals, ErrEmptyID)

	err = api.AddCohosts("", []string{"A"})
	c.Assert(err, Equals, ErrEmptyID)
	err = api.AddCohosts("1234", nil)
	c.Assert(err, Equals, ErrEmptyCohosts)

	err = api.UpdateCohosts("", []string{"A"})
	c.Assert(err, Equals, ErrEmptyID)
	err = api.UpdateCohosts("1234", nil)
	c.Assert(err, Equals, ErrEmptyCohosts)

	err = api.DeleteCohosts("", []string{"A"})
	c.Assert(err, Equals, ErrEmptyID)
	err = api.DeleteCohosts("1234", nil)
	c.Assert(err, Equals, ErrEmptyCohosts)

	api, _ = NewClient("http-error")

	_, err = api.Get("12345678901234")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "API returned error: Conference not found. (ConferenceNotFound)")

	api, _ = NewClient("msg-error")

	_, err = api.Get("12345678901234")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "API returned non-ok status code 404")

	api, _ = NewClient("data-error")

	_, err = api.Get("12345678901234")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Can't decode API response: invalid character 'X' looking for beginning of value")

	API = "http://127.0.0.1:9999"

	api, _ = NewClient("Test1234")
	_, err = api.Get("12345678901234")
	c.Assert(err, NotNil)

	API = "http://127.0.0.1:" + TEST_PORT

	var conf *Conference
	c.Assert(conf.WithCohosts("user@domain.com"), IsNil)
}

// ////////////////////////////////////////////////////////////////////////////////// //

func handlerCreateConference(rw http.ResponseWriter, r *http.Request) {
	if writeErrorResponse(rw, r) {
		return
	}

	rw.WriteHeader(200)
	rw.Write([]byte(`{
  "id": "12345678901234",
  "join_url": "https://telemost.yandex.ru/j/12345678901234",
  "live_stream": {
    "watch_url": "https://telemost.yandex.ru/live/123456789abcdef0123456789abcdef0"
  }
}`))
}

func handlerGetConference(rw http.ResponseWriter, r *http.Request) {
	if writeErrorResponse(rw, r) {
		return
	}

	rw.WriteHeader(200)
	rw.Write([]byte(`{
  "id": "12345678901234",
  "join_url": "https://telemost.yandex.ru/j/12345678901234",
  "access_level": "ORGANIZATION",
  "waiting_room_level": "ORGANIZATION",
  "live_stream": {
    "watch_url": "https://telemost.yandex.ru/live/123456789abcdef0123456789abcdef0",
    "access_level": "PUBLIC",
    "title": "Example conference created via API",
    "description": "Some description of example conference created via API"
  },
  "sip_uri_meeting": "12345678901234567890@sip.t.ya.ru",
  "sip_uri_telemost": "j@sip.t.ya.ru",
  "sip_id": "12345678901234567890"
}`))
}

func handlerUpdateConference(rw http.ResponseWriter, r *http.Request) {
	if writeErrorResponse(rw, r) {
		return
	}

	rw.WriteHeader(200)
	rw.Write([]byte(`{
  "live_stream": {
    "access_level": "PUBLIC",
    "title": "Example conference created via API",
    "description": "Some description of example conference created via API"
  }
}`))
}

func handlerDeleteConference(rw http.ResponseWriter, r *http.Request) {
	if writeErrorResponse(rw, r) {
		return
	}

	rw.WriteHeader(204)
}

func handlerGetCohosts(rw http.ResponseWriter, r *http.Request) {
	if writeErrorResponse(rw, r) {
		return
	}

	rw.WriteHeader(200)
	rw.Write([]byte(`{
  "cohosts": [
    {
      "email": "user1@yandex.ru"
    },
    {
      "email": "user2@org-domain.ru"
    }
  ]
}`))
}

func handlerAddCohosts(rw http.ResponseWriter, r *http.Request) {
	if writeErrorResponse(rw, r) {
		return
	}

	rw.WriteHeader(204)
}

func handlerUpdateCohosts(rw http.ResponseWriter, r *http.Request) {
	if writeErrorResponse(rw, r) {
		return
	}

	rw.WriteHeader(204)
}

func handlerDeleteCohosts(rw http.ResponseWriter, r *http.Request) {
	if writeErrorResponse(rw, r) {
		return
	}

	rw.WriteHeader(204)
}

func writeErrorResponse(rw http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Authorization") == "OAuth http-error" {
		rw.WriteHeader(404)
		rw.Write([]byte(`{
  "message": "Видео встреча не найдена.",
  "description": "Conference not found.",
  "error": "ConferenceNotFound"
}`))
		return true
	}

	if r.Header.Get("Authorization") == "OAuth msg-error" {
		rw.WriteHeader(404)
		rw.Write([]byte(`XXX`))
		return true
	}

	if r.Header.Get("Authorization") == "OAuth data-error" {
		rw.WriteHeader(200)
		rw.Write([]byte(`XXX`))
		return true
	}

	return false
}
