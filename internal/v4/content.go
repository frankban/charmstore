// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package v4

import (
	"archive/zip"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/juju/jujusvg"
	"github.com/juju/utils/jsonhttp"
	"gopkg.in/errgo.v1"
	"gopkg.in/juju/charm.v4"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/juju/charmstore/internal/charmstore"
	"github.com/juju/charmstore/internal/mongodoc"
	"github.com/juju/charmstore/internal/router"
	"github.com/juju/charmstore/params"
)

// GET id/diagram.svg
// http://tinyurl.com/nqjvxov
func (h *Handler) serveDiagram(id *charm.Reference, w http.ResponseWriter, req *http.Request) error {
	if id.Series != "bundle" {
		return errgo.WithCausef(nil, params.ErrNotFound, "diagrams not supported for charms")
	}
	entity, err := h.store.FindEntity(id, "bundledata")
	if err != nil {
		return errgo.Mask(err)
	}

	var urlErr error
	// TODO consider what happens when a charm's SVG does not exist.
	canvas, err := jujusvg.NewFromBundle(entity.BundleData, func(id *charm.Reference) string {
		// TODO change jujusvg so that the iconURL function can
		// return an error.
		absPath := "/" + id.Path() + "/archive/icon.svg"
		p, err := router.RelativeURLPath(req.RequestURI, absPath)
		if err != nil {
			urlErr = errgo.Notef(err, "cannot make relative URL from %q and %q", req.RequestURI, absPath)
		}
		return p
	})
	if err != nil {
		return errgo.Notef(err, "cannot create canvas")
	}
	if urlErr != nil {
		return urlErr
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	canvas.Marshal(w)
	return nil
}

func (h *Handler) serveContent(
	id *charm.Reference,
	w http.ResponseWriter,
	req *http.Request,
	fileId mongodoc.FileId,
	isContent func(*zip.File) bool,
) error {

func (h *Handler) serveIcon(id *charm.Reference, w http.ResponseWriter, req *http.Request) error {
	if id.Series == "bundle" {
		return errgo.WithCausef(nil, params.ErrNotFound, "icons not supported for bundles")
	}
	
	entity, err := h.store.FindEntity(id, "_id", "contents", "blobname")
	if err != nil {
		return errgo.Mask(err, errgo.Is(params.ErrNotFound))
	}
	isIconFile := func(f *zip.File) bool {
		return path.Clean(f.Name) == "icon.svg"
	}

	r, err := h.store.OpenCachedFile(entity, mongodoc.FileIcon, isIconFile)
	if err != nil {
		if errgo.Cause(err) != params.ErrNotFound {
			return errgo.Mask(err)
		}
		return h.serveDefaultIcon(w, req)
	}
	w.Header().Set("ContentType", "image/svg+xml")
	io.Copy(w, r)
	return nil
}
