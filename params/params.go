// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// The params package holds types that are a part of the charm store's external
// contract - they will be marshalled (or unmarshalled) as JSON
// and delivered through the HTTP API.
package params

import (
	"time"

	"gopkg.in/juju/charm.v3"
)

// MetaAnyResponse holds the result of a meta/any
// request. See http://tinyurl.com/q5vcjpk
type MetaAnyResponse struct {
	Id   *charm.Reference
	Meta map[string]interface{} `json:",omitempty"`
}

// Statistic holds one element of a stats/counter
// response. See http://tinyurl.com/nkdovcf
type Statistic struct {
	Key   string `json:",omitempty"`
	Date  string `json:",omitempty"`
	Count int64
}

// ArchivePostResponse holds the result of
// a post to /$id/archive. See http://tinyurl.com/lzrzrgb
type ArchivePostResponse struct {
	Id *charm.Reference
}

// ArchiveSizeResponse holds the result of an
// id/meta/archive-size GET request. See http://tinyurl.com/m8b9geq
type ArchiveSizeResponse struct {
	Size int64
}

// ManifestFile holds information about a charm or bundle file.
// A slice of ManifestFile is used as response for
// id/meta/manifest GET requests. See http://tinyurl.com/p3xdcto
type ManifestFile struct {
	Name string
	Size int64
}

// ArchiveUploadTimeResponse holds the result of an
// id/meta/archive-upload-time GET request. See http://tinyurl.com/nmujuqk
type ArchiveUploadTimeResponse struct {
	UploadTime time.Time
}

// ExtraInfo holds additional metadata recorded for a charm or bundle.
type ExtraInfo struct {
	Values map[string]interface{}
}
