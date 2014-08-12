// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package v4

import (
	"net/url"

	"gopkg.in/juju/charm.v3"

	"github.com/juju/charmstore/internal/mongodoc"
    "github.com/juju/charmstore/params"
)
)

// GET id/meta/extra-info
// http://tinyurl.com/keos7wd
func (h *handler) metaExtraInfo(entity *mongodoc.Entity, id *charm.Reference, path string, method string, flags url.Values) (interface{}, error) {
	switch method {
	case "PUT":
		return h.servePutExtraInfo(entity)
	case "GET":
        return entity.ExtraInfo, nil
	}
    // TODO(rog) params.ErrMethodNotAllowed
    return errgo.Newf("method not allowed")
}

// GET id/meta/extra-info/key
// http://tinyurl.com/polrbn7
func (h *handler) metaExtraInfoWithKey(entity *mongodoc.Entity, id *charm.Reference, path string, method string, flags url.Values) (interface{}, error) {
	return nil, errNotImplemented
}

func (h *handler) servePutExtraInfo(entity *mongodoc.Entity) (interface{}, error) {
    switch method {
    case "PUT":
        return h.servePutExtraInfo(entity)
    case "GET":
        return entity.ExtraInfo, nil
    }
    // TODO(rog) params.ErrMethodNotAllowed
    return errgo.Newf("method not allowed")
}
