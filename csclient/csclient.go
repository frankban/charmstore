// The csclient package provides access to the charm store API.
package csclient

// Client represents the client side of a charm store.
type Client struct {
	params Params
}

// Params holds parameters for creating a new charm store client.
type Params struct {
	// URL holds the root endpoint URL of the charmstore,
	// with no trailing slash, not including the version.
	// For example http://charms.ubuntu.com
	// TODO default this to global charm store address.
	URL string

	// User and Password hold the authentication credentials
	// for the client. If User is empty, no credentials will be
	// sent.
	User string
	Password string
}

// New returns a new charm store client.
func New(p Params) *Client {
	return &Client{
		params: p,
	}
}

// Do makes an arbitrary request to the charm store.
// It adds appropriate headers to the given HTTP request,
// sends it to the charm store and parses the result
// as JSON into the given result value, which should be a pointer to the
// expected data, but may be nil if no result is expected.
//
// This is a low level method - more specific Client methods
// should be used when possible.
func (c *Client) Do(req *http.Request, result interface{}) error {
	if c.params.User != "" {
		userPass := c.params.User + ":" + c.params.Password
		authBasic := base64.StdEncoding.EncodeToString([]byte(userPass))
		req.Header.Set("Authorization", "Basic "+authBasic)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errgo.Mask(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errgo.Notef(err, "cannot read response body")
	}
	if resp.StatusCode != http.StatusOK {
		var perr params.Error
		if err := json.Unmarshal(data, &perr); err != nil {
			return errgo.Notef(err, "cannot unmarshal error response %q", sizeLimit(data))
		}
		if perr.Message == "" {
			return errgo.New("error response with empty message %q", sizeLimit(data))
		}
		return &perr
	}
	if result == nil {
		// The caller doesn't care about the response body.
		return nil
	}
	if err := json.Unmarshal(data, result); err != nil {
		return errgo.Notef(err, "cannot unmarshal response %q", sizeLimit(data))
	}
	return nil
}

func sizeLimit(data []byte) []byte {
	if len(data) < 1024 {
		return data
	}
	return append(data[0:1024], fmt.Sprintf(" ... [%d bytes omitted]", len(data)-1024)...)
}


//func (c *Client) GetEntity(id *charm.Reference) (r io.ReadCloser, hash string, size int64, err error)
//
//func (c *Client) Upload(id *charm.Reference, r io.Reader, hash string, size int64) (*charm.Reference, error)
//
//func (c *Client) VerifyBundle(bd charm.Bundle) error
//
//func (c *Client) Meta(id *charm.Reference, values ...interface{}) error {
//	var fields map[string] interface{}
//	for _, v := range values {
//		switch v := v.(type) {
//		case *params.ArchiveUploadTimeResponse:
//			fields["archive-upload-time" = v
//		case *params.ArchiveSizeResponse:
//			fields["archive-size"] = v
//		case *extraInfo:
//		...
//	}
//
//	map[string] json.RawMessage
//		
//}
//
//type Meta interface {
//	private()
//}
//
//func ArchiveSize(r *int64) Meta
//
//func UploadTime(r *time.Time) Meta
//
//	var archiveSize int64
//	var uploadTime time.Time
//	client.Meta(id, csclient.ArchiveSize(&archiveSize), csclient.UploadTime(&uploadTime))
//
//
//
//// Meta returns metadata on the charm or bundle with the
//// given id. The result value provides a value
//// to be filled in with the result, which must be
//// a pointer to a struct containing members corresponding
//// to possible metadata include parameters (see https://docs.google.com/a/canonical.com/document/d/1TgRA7jW_mmXoKH3JiwBbtPvQu7WiM6XMrz1wSrhTMXw/edit#bookmark=id.p22xdlv0861a).
////
//// The name of the struct member is translated to
//// a lower case hyphen-separated form; for example,
//// ArchiveSize becomes "archive-size", and BundleMachineCount
//// becomes "bundle-machine-count", but may also
//// be specified in the field's tag.
////
//// This example will fill in the result structure with information
//// about the given id, including information on its archive
//// size (include archive-size), upload time (include archive-upload-time)
//// and digest (include extra-info/digest).
////
////	var result struct {
////		ArchiveSize params.ArchiveSizeResponse
////		ArchiveUploadTime params.ArchiveUploadTimeResponse
////		Digest string `csclient:"extra-info/digest"`
////	}
////	err := client.Meta(id, &result)
//func (c *Client) Meta(id *charm.Reference, result interface{}) error {
//	
//}
//
//func hyphenate(s string) string {
//	var buf bytes.Buffer
//	var prevLower bool
//	for i, r := range s {
//		if !unicode.IsUpper(r) {
//			prevLower = true
//			buf.WriteRune(r)
//			continue
//		}
//		if prevLower {
//			buf.WriteRune('-')
//		}
//		buf.WriteRune(unicode.ToLower(r)
//		prevLower = false
//	}
//	return buf.String()
//}
//
//func (c *Client) Meta(id *charm.Reference, result interface{}) error
//
//
//	var result struct {
//		ArchiveSize csclient.ArchiveSize
//		UploadTime csclient.UploadTime
//		Digest string			`csclient:"extra-info/digest"`
//	}
//
//
//
//
//
//	var result struct {
//		ArchiveSize int64
//		Uploaded time.Time
//	}
//
//	var result struct {
//		ArchiveSize params.ArchiveSizeResponse
//		ArchiveUploadTime params.ArchiveUploadTimeResponse
//		BundleCount params.BundleCount
//	}
//	jsonRequest(id.String() + "/meta/any?include=", &result)
//
//	var result struct {
//		ArchiveSize params.ArchiveSizeResponse
//	}
//	client.Meta(id, &result)
//
//
//	var results []struct {
//		ArchiveSize params.ArchiveSizeResponse
//		ArchiveUploadTime params.ArchiveUploadTimeResponse
//	}
//	client.BulkMeta(ids, &results)
//
//		jsonRequest(id.String() + "/meta/any?include=", &result)
//
//
//func (c *Client) BulkMeta(ids []*charm.Reference, valueSlicePtr interface{})
//
//{
//	id
//	value...
//}, ...
//
//id...
//value..., value..., value...
//
//id...
//{value, ...}, ...
//
//
//
//
//	var size params.ArchiveSizeResponse
//	var uploaded params.ArchiveUploadTimeResponse
//	var bd *charm.BundleData
//	err := client.Meta(wordpress, &size, &uploaded, &bd)
//
//	var digest string
//
//	err := client.Meta(wordpress, csclient.ExtraInfoField("digest", &digest))
//
//
//var result params.MetaAnyResult
//Get("wordpress/meta/any?include=....", &{})
//
//type AllInfo struct {
//}
//
