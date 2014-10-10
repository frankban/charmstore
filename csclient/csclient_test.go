// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package csclient_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	gc "gopkg.in/check.v1"
	"gopkg.in/juju/charm.v4"
	charmtesting "gopkg.in/juju/charm.v4/testing"
	"gopkg.in/mgo.v2"

	"github.com/juju/charmstore/csclient"
	"github.com/juju/charmstore/internal/charmstore"
	"github.com/juju/charmstore/internal/storetesting"
	"github.com/juju/charmstore/internal/v4"
	"github.com/juju/charmstore/params"
)

type suite struct {
	storetesting.IsolatedMgoSuite
	client *csclient.Client
	srv    *httptest.Server
	store  *charmstore.Store
}

var _ = gc.Suite(&suite{})

var serverParams = charmstore.ServerParams{
	AuthUsername: "test-user",
	AuthPassword: "test-password",
}

func newServer(c *gc.C, session *mgo.Session, config charmstore.ServerParams) (*httptest.Server, *charmstore.Store) {
	db := session.DB("charmstore")
	store, err := charmstore.NewStore(db, nil)
	c.Assert(err, gc.IsNil)
	handler, err := charmstore.NewServer(db, nil, config, map[string]charmstore.NewAPIHandlerFunc{"v4": v4.NewAPIHandler})
	c.Assert(err, gc.IsNil)
	return httptest.NewServer(handler), store
}

func (s *suite) SetUpTest(c *gc.C) {
	s.IsolatedMgoSuite.SetUpTest(c)
	s.srv, s.store = newServer(c, s.Session, serverParams)
	s.client = csclient.New(csclient.Params{
		URL:      s.srv.URL,
		User:     serverParams.AuthUsername,
		Password: serverParams.AuthPassword,
	})
}

func (s *suite) TearDownTest(c *gc.C) {
	s.srv.Close()
	s.IsolatedMgoSuite.TearDownTest(c)
}

var doTests = []struct {
	about        string
	method       string
	path         string
	expectResult interface{}
	expectError  string
}{{
	about: "example 1",
	path:  "/wordpress/expand-id",
	expectResult: []params.ExpandedId{{
		Id: "cs:utopic/wordpress-42",
	}},
}}

func (s *suite) TestDo(c *gc.C) {
	ch := charmtesting.Charms.CharmDir("wordpress")
	url := mustParseReference("utopic/wordpress-42")
	err := s.store.AddCharmWithArchive(url, ch)
	c.Assert(err, gc.IsNil)

	for i, test := range doTests {
		c.Logf("test %d: %s", i, test.about)

		if test.method == "" {
			test.method = "GET"
		}

		// Set up the request.
		req, err := http.NewRequest(test.method, "", nil)
		c.Assert(err, gc.IsNil)

		// Send the request.
		var result json.RawMessage
		err = s.client.Do(req, test.path, &result)

		// Check the response.
		if test.expectError != "" {
			c.Assert(err, gc.ErrorMatches, test.expectError)
			continue
		}
		c.Assert(err, gc.IsNil)
		if test.expectResult != nil {
			c.Assert([]byte(result), storetesting.JSONEquals, test.expectResult)
		}
	}
}

func mustParseReference(url string) *charm.Reference {
	// TODO implement MustParseReference in charm.
	ref, err := charm.ParseReference(url)
	if err != nil {
		panic(err)
	}
	return ref
}
