//+build ignore

package unused

// stuff that's unused for now.

// EntityDoc holds the in-database representation of charm or bundle's
// document in the charm store.
type EntityDoc struct {
	// URL holds the fully specified URL of the charm or bundle.
	// e.g. cs:precise/wordpress-34, cs:~user/quantal/foo-2
	URL *charm.URL `bson:"_id"`

	// BaseURL holds the URL of the charm or bundle with the
	// series and revision omitted.
	// e.g. cs:wordpress, cs:~user/foo
	BaseURL *charm.URL

	Revision int
	Sha256   string // This is also used as a blob reference.
	Size     int64

	UploadTime time.Time

	CharmMeta    *charm.Meta
	CharmConfig  *charm.Config
	CharmActions *charm.Actions

	// CharmProvidedInterfaces holds all the relation
	// interfaces provided by the charm
	CharmProvidedInterfaces []string

	// CharmRequiredInterfaces is similar to CharmProvidedInterfaces
	// for required interfaces.
	CharmRequiredInterfaces []string

	BundleData   *charm.BundleData
	BundleReadMe string

	// BundleCharms includes all the charm URLs referenced
	// by the bundle, including base URLs where they are
	// not already included.
	BundleCharms []*charm.URL

	// TODO Add fields denormalized for search purposes
	// and search ranking field(s).
}

// wordpress/meta/any?include=charm-metadata&include=color&include=charm-actions

func (api *apiV4) metaCharmMetadata(getter Getter, id, path string, flags url.Values) (interface{}, error) {
	var meta *charm.Meta
	err := getter.GetInfo("charms", "wordpress", "charmmeta", &meta)

}

type MetaHandler func(getter ItemGetter, ids []string, path string, flags url.Values) (map[string]interface{}, error)

func SingleId(f func(getter ItemGetter, id string, path string, flags url.Values) (map[string]interface{}, error)) MetaHandler {
	return func(getter ItemGetter, ids []string, path string, flags url.Values) (map[string]interface{}, error) {
		m := make(map[string]interface{})
		for _, id := range ids {
			item, err := f(getter, id, path, flag)
			if err != nil {
				return nil, err
			}
			m[id] = item
		}
		return m, nil
	}
}

func (h *APIHander) getMetadataSingleFlight(db *mgo.Database, id string, includes []string) (interface{}, error) {
	var group singleflight.Group
	var getter getterFunc = func(collection string, id string, field string, val interface{}) error {
		result, err := group.Do(collection+":"+id, func() (interface{}, error) {
			var result bson.Raw
			if err := db.C(collection).Find(bson.D{{"_id", id}}).One(&result); err != nil {
				return nil, err
			}
			return &result, nil
		})
		if err != nil {
			return err
		}
		return bson.Unmarshal(result.(*bson.Raw).Data, val)
	}

	type result struct {
		include string
		val     interface{}
		err     error
	}
	resultCh := make(chan result)
	for _, include := range includes {
		include := include
		handler := h.Meta[include]
		if handler == nil {
			TODOgofmt
			continue
		}
		go func() {
			val, err := handler(getter, id, include, nil)
			resultCh <- result{
				include: include,
				val:     val,
				err:     err,
			}
		}()
	}
	results := make(map[string]interface{})
	for _ = range includes {
		r := <-resultCh
		if r.err != nil {
			TODO
		}
		results[r.include] = r.val
	}
	return results, nil
}

func (h *APIHander) getMetadataOptimized(db *mgo.Database, id string, includes []string) (interface{}, error) {
	type request struct {
		collection string
		id         string
		field      string
		val        interface{}
		reply      chan error
	}
	reqChan := make(chan request)
	var getter getterFunc = func(collection string, id string, field string, val interface{}) error {
		reply := make(chan error)
		reqChan <- request{
			collection: collection,
			id:         id,
			field:      field,
			val:        val,
			reply:      reply,
		}
		return <-reply
	}
	type handlerResult struct {
		include string
		val     interface{}
		err     error
	}
	done := make(chan struct{})
	for _, include := range includes {
		include := include
		handler := h.Meta[include]
		if handler == nil {
			TODO
			continue
		}

		go func() {
			val, err := handler(getter, id, include)
			done <- handlerResult{
				include: include,
				val:     val,
				err:     err,
			}
		}()
	}
	type item struct {
		collection string
		id         string
	}

	results := make(map[string]interface{})
	remaining := len(includes)
	blocked := make(map[item][]request)
	numBlocked := 0
	for len(remaining) > 0 {
		if numBlocked == len(remaining) {
			for item, reqs := range blocked {
				vals := make(map[string]interface{})
				for _, req := range reqs {
					vals[req.field] = req.val
				}
				err := db.C(item.collection).Find(bson.D{{"_id", item.id}}).One(vals)
				if err != nil {
					TODO
				}
				for _, req := range reqs {
					req.reply <- vals[req.field]
				}
			}
		}
		select {
		case result := <-done:
			if result.err != nil {
				return error
			}
			results[result.include] = result.val
		case request := <-reqChan:
			item := item{
				collection: request.collection,
				id:         request.id,
			}
			blocked[item] = append(blocked[item], request)
			numBlocked++
		}
	}
}
