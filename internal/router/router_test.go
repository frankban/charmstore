package router

import (
	"encoding/json"
	"fmt"
	"github.com/juju/charmstore/params"
	jujutesting "github.com/juju/testing"
	"gopkg.in/juju/charm.v2"
	"io"
	gc "launchpad.net/gocheck"
	"strings"

	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestPackage(t *testing.T) {
	jujutesting.MgoTestPackage(t, nil)
}

type RouterSuite struct {
	jujutesting.MgoSuite
}

var _ = gc.Suite(&RouterSuite{})

var routerTests = []struct {
	about      string
	handlers   Handlers
	urlStr     string
	expectCode int
	expectBody interface{}
}{{
	about: "global handler",
	handlers: Handlers{
		Global: map[string]http.Handler{
			"foo": HandleJSON(func(w http.ResponseWriter, req *http.Request) (interface{}, error) {
				return &Foo{"hello"}, nil
			}),
		},
	},
	urlStr:     "http://example.com/foo",
	expectCode: http.StatusOK,
	expectBody: Foo{"hello"},
}, {
	about: "id handler",
	handlers: Handlers{
		Id: map[string]IdHandler{
			"foo": testIdHandler,
		},
	},
	urlStr:     "http://example.com/precise/wordpress-34/foo",
	expectCode: http.StatusOK,
	expectBody: idHandlerTestResp{
		CharmURL: "cs:precise/wordpress-34",
	},
}, {
	about: "id handler with extra path",
	handlers: Handlers{
		Id: map[string]IdHandler{
			"foo/": testIdHandler,
		},
	},
	urlStr:     "http://example.com/precise/wordpress-34/foo/blah/arble",
	expectCode: http.StatusOK,
	expectBody: idHandlerTestResp{
		CharmURL: "cs:precise/wordpress-34",
		Path:     "blah/arble",
	},
}, {
	about: "id handler with allowed extra path but none given",
	handlers: Handlers{
		Id: map[string]IdHandler{
			"foo/": testIdHandler,
		},
	},
	urlStr:     "http://example.com/precise/wordpress-34/foo",
	expectCode: http.StatusInternalServerError,
	expectBody: params.Error{
		Message: "not found",
	},
}, {
	about: "id handler with unwanted extra path",
	handlers: Handlers{
		Id: map[string]IdHandler{
			"foo": testIdHandler,
		},
	},
	urlStr:     "http://example.com/precise/wordpress-34/foo/blah",
	expectCode: http.StatusInternalServerError,
	expectBody: params.Error{
		Message: "not found",
	},
}, {
	about: "id handler with user",
	handlers: Handlers{
		Id: map[string]IdHandler{
			"foo": testIdHandler,
		},
	},
	urlStr:     "http://example.com/~joe/precise/wordpress-34/foo",
	expectCode: http.StatusOK,
	expectBody: idHandlerTestResp{
		CharmURL: "cs:~joe/precise/wordpress-34",
	},
}, {
	about: "id handler with user and extra path",
	handlers: Handlers{
		Id: map[string]IdHandler{
			"foo/": testIdHandler,
		},
	},
	urlStr:     "http://example.com/~joe/precise/wordpress-34/foo/blah/arble",
	expectCode: http.StatusOK,
	expectBody: idHandlerTestResp{
		CharmURL: "cs:~joe/precise/wordpress-34",
		Path:     "blah/arble",
	},
}}

type idHandlerTestResp struct {
	CharmURL string
	Path     string
}

func testIdHandler(charmId *charm.URL, w http.ResponseWriter, req *http.Request) error {
	WriteJSON(w, http.StatusOK, idHandlerTestResp{
		CharmURL: charmId.String(),
		Path:     req.URL.Path,
	})
	return nil
}

func (s *RouterSuite) TestRouter(c *gc.C) {
	for i, test := range routerTests {
		c.Logf("test %d: %s", i, test.about)
		router := New(s.Session.DB("database"), &test.handlers)
		assertJSONCall(c, router, "GET", test.urlStr, "", test.expectCode, test.expectBody)
	}
}

var splitIdTests = []struct {
	path        string
	expectURL   string
	expectError string
}{{
	path:      "precise/wordpress-23",
	expectURL: "cs:precise/wordpress-23",
}, {
	path:      "~user/precise/wordpress-23",
	expectURL: "cs:~user/precise/wordpress-23",
}, {
	path:      "wordpress",
	expectURL: "cs:wordpress",
}, {
	path:      "~user/wordpress",
	expectURL: "cs:~user/wordpress",
}, {
	path:        "",
	expectError: `charm URL has invalid charm name: ""`,
}, {
	path:        "~foo-bar-/wordpress",
	expectError: `charm URL has invalid user name: "~foo-bar-/wordpress"`,
}}

func (s *RouterSuite) TestSplitId(c *gc.C) {
	for i, test := range splitIdTests {
		c.Logf("test %d: %s", i, test.path)
		url, rest, err := splitId(test.path)
		if test.expectError != "" {
			c.Assert(err, gc.ErrorMatches, test.expectError)
			c.Assert(url, gc.IsNil)
			c.Assert(rest, gc.Equals, "")
			continue
		}
		c.Assert(url.String(), gc.Equals, test.expectURL)
		c.Assert(rest, gc.Equals, "")

		url, rest, err = splitId(test.path + "/some/more")
		c.Assert(err, gc.Equals, nil)
		c.Assert(url.String(), gc.Equals, test.expectURL)
		c.Assert(rest, gc.Equals, "some/more")
	}
}

func (s *RouterSuite) TestWriteJSON(c *gc.C) {
	rec := httptest.NewRecorder()
	type Foo struct {
		N int
	}
	err := WriteJSON(rec, http.StatusTeapot, Foo{1234})
	c.Assert(err, gc.IsNil)
	c.Assert(rec.Code, gc.Equals, http.StatusTeapot)
	c.Assert(rec.Body.String(), gc.Equals, `{"N":1234}`)
	c.Assert(rec.Header().Get("content-type"), gc.Equals, "application/json")
}

func (s *RouterSuite) TestWriteError(c *gc.C) {
	rec := httptest.NewRecorder()
	WriteError(rec, fmt.Errorf("an error"))
	var errResp params.Error
	err := json.Unmarshal(rec.Body.Bytes(), &errResp)
	c.Assert(err, gc.IsNil)
	c.Assert(errResp, gc.Equals, params.Error{Message: "an error"})

	rec = httptest.NewRecorder()
	errResp0 := params.Error{
		Message: "a message",
		Code:    "some code",
	}
	WriteError(rec, &errResp0)
	var errResp1 params.Error
	err = json.Unmarshal(rec.Body.Bytes(), &errResp1)
	c.Assert(err, gc.IsNil)
	c.Assert(errResp1, gc.Equals, errResp0)
}

var handlerTests = []struct {
	about      string
	handler    http.Handler
	urlStr     string
	expectCode int
	expectBody interface{}
}{{
	about: "handleErrors, normal error",
	handler: HandleErrors(func(http.ResponseWriter, *http.Request) error {
		return fmt.Errorf("an error")
	}),
	urlStr:     "http://example.com",
	expectCode: http.StatusInternalServerError,
	expectBody: params.Error{
		Message: "an error",
	},
}, {
	about: "handleErrors, error with code",
	handler: HandleErrors(func(http.ResponseWriter, *http.Request) error {
		return &params.Error{
			Message: "something went wrong",
			Code:    "snafu",
		}
	}),
	urlStr:     "http://example.com",
	expectCode: http.StatusInternalServerError,
	expectBody: params.Error{
		Message: "something went wrong",
		Code:    "snafu",
	},
}, {
	about: "handleErrors, no error",
	handler: HandleErrors(func(w http.ResponseWriter, req *http.Request) error {
		w.WriteHeader(http.StatusTeapot)
		return nil
	}),
	expectCode: http.StatusTeapot,
}, {
	about: "handleJSON, normal case",
	handler: HandleJSON(func(w http.ResponseWriter, req *http.Request) (interface{}, error) {
		return &Foo{"hello"}, nil
	}),
	expectCode: http.StatusOK,
	expectBody: Foo{"hello"},
}, {
	about: "handleJSON, error case",
	handler: HandleJSON(func(w http.ResponseWriter, req *http.Request) (interface{}, error) {
		return nil, fmt.Errorf("an error")
	}),
	expectCode: http.StatusInternalServerError,
	expectBody: params.Error{
		Message: "an error",
	},
}}

type Foo struct {
	S string
}

func (s *RouterSuite) TestHandlers(c *gc.C) {
	for i, test := range handlerTests {
		c.Logf("test %d: %s", i, test.about)
		assertJSONCall(c, test.handler, "GET", "http://example.com", "", test.expectCode, test.expectBody)
	}
}

func assertJSONCall(
	c *gc.C,
	handler http.Handler,
	method string,
	urlStr string,
	body string,
	expectCode int,
	expectBody interface{},
) {
	rec := callHandler(c, handler, method, urlStr, body)
	c.Assert(rec.Code, gc.Equals, expectCode, gc.Commentf("body: %s", rec.Body.Bytes()))
	if expectBody == nil {
		c.Assert(rec.Body.Bytes(), gc.HasLen, 0)
		return
	}
	resp := reflect.New(reflect.TypeOf(expectBody))
	err := json.Unmarshal(rec.Body.Bytes(), resp.Interface())
	c.Assert(err, gc.IsNil)
	c.Assert(resp.Elem().Interface(), gc.DeepEquals, expectBody)
}

func callHandler(c *gc.C, handler http.Handler, method string, urlStr string, body string) *httptest.ResponseRecorder {
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, urlStr, r)
	c.Assert(err, gc.IsNil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
