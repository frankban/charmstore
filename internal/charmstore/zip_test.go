// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package charmstore_test

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"

	jujutesting "github.com/juju/testing"
	gc "gopkg.in/check.v1"

	"github.com/juju/charmstore/internal/charmstore"
)

type zipSuite struct {
	jujutesting.IsolationSuite
	contents map[string]string
}

var _ = gc.Suite(&zipSuite{})

func (s *zipSuite) SetUpSuite(c *gc.C) {
	s.IsolationSuite.SetUpSuite(c)
	s.contents = map[string]string{
		"readme.md":     "readme contents",
		"icon.svg":      "icon contents",
		"metadata.yaml": "metadata contents",
	}
}

func (s *zipSuite) makeZipReader(c *gc.C, contents map[string]string) (io.ReadSeeker, []*zip.File) {
	// Create a customized zip archive in memory.
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for name, content := range contents {
		f, err := w.Create(name)
		c.Assert(err, gc.IsNil)
		_, err = f.Write([]byte(content))
		c.Assert(err, gc.IsNil)
	}
	c.Assert(w.Close(), gc.IsNil)

	// Retrieve the zip files in the archive.
	zipReader := bytes.NewReader(buf.Bytes())
	r, err := zip.NewReader(zipReader, int64(buf.Len()))
	c.Assert(err, gc.IsNil)
	c.Assert(r.File, gc.HasLen, len(contents))
	return zipReader, r.File
}

func (s *zipSuite) TestZipFileReader(c *gc.C) {
	zipReader, files := s.makeZipReader(c, s.contents)

	// A new ZipFile is correctly created from each zip file in the archive.
	for i, f := range files {
		c.Logf("test %d: %s", i, f.Name)
		zf, err := charmstore.NewZipFile(f)
		c.Assert(err, gc.IsNil)
		zfr, err := charmstore.ZipFileReader(zipReader, zf)
		c.Assert(err, gc.IsNil)
		content, err := ioutil.ReadAll(zfr)
		c.Assert(err, gc.IsNil)
		c.Assert(string(content), gc.Equals, s.contents[f.Name])
	}
}

func (s *zipSuite) TestNewZipFile(c *gc.C) {
	_, files := s.makeZipReader(c, s.contents)

	// A new ZipFile is correctly created from each zip file in the archive.
	for i, f := range files {
		c.Logf("test %d: %s", i, f.Name)
		zf, err := charmstore.NewZipFile(f)
		c.Assert(err, gc.IsNil)
		offset, err := f.DataOffset()
		c.Assert(err, gc.IsNil)

		c.Assert(zf.Offset, gc.Equals, offset)
		c.Assert(zf.Size, gc.Equals, int64(f.CompressedSize64))
		c.Assert(zf.Compressed, gc.Equals, true)
	}
}

func (s *zipSuite) TestNewZipFileWithCompressionMethodError(c *gc.C) {
	_, files := s.makeZipReader(c, map[string]string{"foo": "contents"})
	f := files[0]
	f.Method = 99
	_, err := charmstore.NewZipFile(f)
	c.Assert(err, gc.ErrorMatches, `unknown zip compression method for "foo"`)
}
