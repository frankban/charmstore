// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package storetesting

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	jc "github.com/juju/testing/checkers"
	gc "launchpad.net/gocheck"
)

type JSONCallParams struct {
	Handler         http.Handler
	Method          string
	URL             string
	Body            string
	BodyContentType string
	ExpectCode      int
	ExpectBody      interface{}
}

// AssertJSONCall asserts that when the given handler is called with
// the given method, URL, and body, the result has the expected
// status code and body.
func AssertJSONCall(c *gc.C, p JSONCallParams) {
	if p.Method == "" {
		p.Method = "GET"
	}
	rec := DoRequest(c, p.Handler, p.Method, p.URL, p.Body, p.BodyContentType, nil)
	c.Assert(rec.Code, gc.Equals, p.ExpectCode, gc.Commentf("body: %s", rec.Body.Bytes()))
	if p.ExpectBody == nil {
		c.Assert(rec.Body.Bytes(), gc.HasLen, 0)
		return
	}
	// Rather than unmarshaling into something of the expected
	// body type, we reform the expected body in JSON and
	// back to interface{}, so we can check the whole content.
	// Otherwise we lose information when unmarshaling.
	expectBodyBytes, err := json.Marshal(p.ExpectBody)
	c.Assert(err, gc.IsNil)
	var expectBodyVal interface{}
	err = json.Unmarshal(expectBodyBytes, &expectBodyVal)
	c.Assert(err, gc.IsNil)

	var gotBodyVal interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &gotBodyVal)
	c.Assert(err, gc.IsNil, gc.Commentf("json body: %q", rec.Body.Bytes()))
	// TODO(rog) check that content type is application/json
	c.Assert(gotBodyVal, jc.DeepEquals, expectBodyVal)
}

// DoRequest invokes a request on the given handler with the given
// method, URL, body and headers.
func DoRequest(c *gc.C, handler http.Handler, method string, urlStr string, body, bodyContentType string, header map[string][]string) *httptest.ResponseRecorder {
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, urlStr, r)
	c.Assert(err, gc.IsNil)
	if header != nil {
		req.Header = header
	}
	if bodyContentType != "" {
		req.Header.Set("Content-Type", bodyContentType)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
