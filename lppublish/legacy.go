// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package lppublish

import (
	"crypto/sha512"
	"fmt"
	"io"
	"os"

	"github.com/juju/errgo"
	"gopkg.in/juju/charm.v4"

	"github.com/juju/charmstore/params"
)

// publishLegacyCharm publishes all existing revisions of the given urls
// in the charm store. Existing revisions are fetched from the legacy charm
// store.
func (cl *charmLoader) publishLegacyCharm(urls []*charm.Reference) error {
	// An Info call on the first URL *should* tell us the latest revision of
	// all the URLs because in legacy charm store, they all match,
	// but we'll fetch info for all the URLs just to sanity check.
	locs := make([]charm.Location, len(urls))
	for i, url := range locs {
		locs[i] = url
	}
	infos, err := charm.Store.Info(locs...)
	if err != nil {
		return errgo.Notef(err, "cannot get information on %q", urls)
	}
	if len(infos) != len(urls) {
		return errgo.Newf("unexpected response count %d, expected %d", len(infos), len(urls))
	}
	rev, digest, hash := infos[0].Revision, infos[0].Digest, infos[0].Sha256
	for i, info := range infos {
		if len(info.Errors) != 0 {
			return errgo.Newf("cannot retrieve info on %s: %s", urls[i], info.Errors[0])
		}
		if info.Revision != rev ||
			info.Digest != digest ||
			info.Sha256 != hash {
			return errgo.Newf("mismatched information from promulgated urls %q", urls)
		}
	}
	for rev := 0; rev <= infos[0].Revision; rev++ {
		for _, url := range urls {
			url.Revision = rev
		}
		if err := cl.putLegacyCharm(urls); err != nil {
			if errgo.Cause(err) == params.ErrUnauthorized {
				return err
			}
			logger.Errorf("cannot put legacy charm: %v", urls, err)
		}
	}
	return nil
}

func (cl *charmLoader) putLegacyCharm(urls []*charm.Reference) error {
	ch, err := legacyCharmStoreGet(urls[0])
	if err != nil {
		return errgo.Notef(err, "cannot get %q", urls[0])
	}
	f, err := os.Open(ch.Path)
	if err != nil {
		return errgo.Mask(err)
	}
	defer f.Close()
	hasher := sha512.New384()
	size, err := io.Copy(hasher, f)
	if err != nil {
		return errgo.Notef(err, "cannot read charm archive: %v", err)
	}
	hash := fmt.Sprintf("%x", hasher.Sum(nil))
	for _, url := range urls {
		_, err := f.Seek(0, 0)
		if err != nil {
			return errgo.Mask(err)
		}
		_, err = cl.uploadArchive("PUT", f, url, size, hash)
		if err != nil {
			return errgo.NoteMask(err, fmt.Sprintf("cannot put %q", url), errgo.Is(params.ErrUnauthorized))
		}
	}
	return nil
}

func legacyCharmStoreGet(url *charm.Reference) (*charm.CharmArchive, error) {
	url1, err := url.URL("")
	if err != nil {
		// We added the series earlier.
		panic(fmt.Errorf("cannot happen: %v", err))
	}
	ch0, err := charm.Store.Get(url1)
	if err != nil {
		return nil, errgo.Mask(err)
	}
	return ch0.(*charm.CharmArchive), nil
}
