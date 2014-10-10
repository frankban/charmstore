// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package storetesting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

// JSONCallParams holds parameters for AssertJSONCall.
// If left empty, some fields will automatically be filled with defaults.
type JSONCallParams struct {
	// Handler holds the handler to use to make the request.
	Handler http.Handler

	// Method holds the HTTP method to use for the call.
	// GET is assumed if this is empty.
	Method string

	// URL holds the URL to pass when making the request.
	URL string

	// Body holds the body to send in the request.
	Body io.Reader

	// Header specifies the HTTP headers to use when making
	// the request.
	Header http.Header

	// ContentLength specifies the length of the body.
	// It may be zero, in which case the default net/http
	// content-length behaviour will be used.
	ContentLength int64

	// Username, if specified, is used for HTTP basic authentication.
	Username string

	// Password, if specified, is used for HTTP basic authentication.
	Password string

	// ExpectStatus holds the expected HTTP status code.
	// http.StatusOK is assumed if this is zero.
	ExpectStatus int

	// ExpectBody holds the expected JSON body.
	ExpectBody interface{}
}

// AssertJSONCall asserts that when the given handler is called with
// the given parameters, the result is as specified.
func AssertJSONCall(c *gc.C, p JSONCallParams) {
	c.Logf("JSON call, url %q", p.URL)
	if p.ExpectStatus == 0 {
		p.ExpectStatus = http.StatusOK
	}
	rec := DoRequest(c, DoRequestParams{
		Handler:       p.Handler,
		Method:        p.Method,
		URL:           p.URL,
		Body:          p.Body,
		Header:        p.Header,
		ContentLength: p.ContentLength,
		Username:      p.Username,
		Password:      p.Password,
	})
	AssertResponse(c, rec, p.ExpectStatus, p.ExpectBody)
}

// AssertResponse asserts that the given response recorder has recorded the
// given HTTP status and response body.
func AssertResponse(c *gc.C, rec *httptest.ResponseRecorder, expectStatus int, expectBody interface{}) {
	c.Assert(rec.Code, gc.Equals, expectStatus, gc.Commentf("body: %s", rec.Body.Bytes()))

	// Ensure the response includes the expected body.
	if expectBody == nil {
		c.Assert(rec.Body.Bytes(), gc.HasLen, 0)
		return
	}
	c.Assert(rec.Header().Get("Content-Type"), gc.Equals, "application/json")
	c.Assert(rec.Body.Bytes(), JSONEquals, expectBody)
}

type jsonEqualChecker struct {
	*gc.CheckerInfo
}

func (checker *jsonEqualChecker) Check(params []interface{}, names []string) (result bool, error string) {
	gotContent, ok := params[0].([]byte)
	if !ok {
		return false, fmt.Sprintf("expected []byte, got %T", params[0])
	}
	expectContent := params[1]
	expectContentBytes, err := json.Marshal(expectContent)
	if err != nil {
		return false, fmt.Sprintf("cannot marshal expected contents: %v", err)
	}
	var expectContentVal interface{}
	if err := json.Unmarshal(expectContentBytes, &expectContentVal); err != nil {
		return false, fmt.Sprintf("cannot unmarshal expected contents: %v", err)
	}

	var gotContentVal interface{}
	if err := json.Unmarshal(gotContent, &gotContentVal); err != nil {
		return false, fmt.Sprintf("cannot unmarshal obtained contents: %v; %q", err, gotContent)
	}

	if ok, err := jc.DeepEqual(gotContentVal, expectContentVal); !ok {
		return false, err.Error()
	}
	return true, ""
}

// JSONEquals defines a checker that checks whether a byte slice, when
// unmarshaled as JSON, is equal to the given value.
// Rather than unmarshaling into something of the expected
// body type, we reform the expected body in JSON and
// back to interface{}, so we can check the whole content.
// Otherwise we lose information when unmarshaling.
var JSONEquals = &jsonEqualChecker{
	&gc.CheckerInfo{Name: "JSONEquals", Params: []string{"obtained", "expected"}},
}

// DoRequestParams holds parameters for DoRequest.
// If left empty, some fields will automatically be filled with defaults.
type DoRequestParams struct {
	// Handler holds the handler to use to make the request.
	Handler http.Handler

	// Method holds the HTTP method to use for the call.
	// GET is assumed if this is empty.
	Method string

	// URL holds the URL to pass when making the request.
	URL string

	// Body holds the body to send in the request.
	Body io.Reader

	// Header specifies the HTTP headers to use when making
	// the request.
	Header http.Header

	// ContentLength specifies the length of the body.
	// It may be zero, in which case the default net/http
	// content-length behaviour will be used.
	ContentLength int64

	// Username, if specified, is used for HTTP basic authentication.
	Username string

	// Password, if specified, is used for HTTP basic authentication.
	Password string
}

// DoRequest invokes a request on the given handler with the given
// parameters.
func DoRequest(c *gc.C, p DoRequestParams) *httptest.ResponseRecorder {
	if p.Method == "" {
		p.Method = "GET"
	}
	srv := httptest.NewServer(p.Handler)
	defer srv.Close()

	req, err := http.NewRequest(p.Method, srv.URL+p.URL, p.Body)
	c.Assert(err, gc.IsNil)
	if p.Header != nil {
		req.Header = p.Header
	}
	if p.ContentLength != 0 {
		req.ContentLength = p.ContentLength
	}
	if p.Username != "" || p.Password != "" {
		req.SetBasicAuth(p.Username, p.Password)
	}
	resp, err := http.DefaultClient.Do(req)
	c.Assert(err, gc.IsNil)
	defer resp.Body.Close()

	// TODO(rog) don't return a ResponseRecorder because we're not actually
	// using httptest.NewRecorder ?
	var rec httptest.ResponseRecorder
	rec.HeaderMap = resp.Header
	rec.Code = resp.StatusCode
	rec.Body = new(bytes.Buffer)
	_, err = io.Copy(rec.Body, resp.Body)
	c.Assert(err, gc.IsNil)
	return &rec
}
