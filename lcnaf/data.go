package lcnaf

import (
	"context"
	"fmt"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/data"
	"gocloud.dev/blob"
	"io"
	"net/http"
	"net/url"
	"os"
)

// DATA_JSON is the name of the embedded LCNAF data included with this package.
const DATA_JSON string = "lcnaf.csv.bz2"

// DATA_GITHUB is the URL for the embedded LCNAF data included with this package on GitHub.
const DATA_GITHUB string = "https://github.com/sfomuseum/go-sfomuseum-libraryofcongress/raw/main/data/lcnaf.csv.bz2"

// OpenData() returns an `io.ReadCloser` instance containing LCNAF data derived from 'uri' which is expected to
// take the form of:
func OpenData(ctx context.Context, uri string) (io.ReadCloser, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	var source string

	switch u.Host {
	case "sfomuseum":
		source = u.Path
	default:
		source = u.Host
	}

	switch source {

	case "blob":

		path := u.Path
		q := u.Query()

		bucket_uri := q.Get("uri")

		bucket, err := blob.OpenBucket(ctx, bucket_uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to open bucket (%s), %w", bucket_uri, err)
		}

		defer bucket.Close()

		r, err := bucket.NewReader(ctx, path, nil)

		if err != nil {
			return nil, fmt.Errorf("Failed to create new reader for %s, %w", path, err)
		}

		return r, nil

	case "github":

		rsp, err := http.Get(DATA_GITHUB)

		if err != nil {
			return nil, fmt.Errorf("Failed to load remote data from Github, %w", err)
		}

		return rsp.Body, nil

	case "file":

		path := u.Path

		r, err := os.Open(path)

		if err != nil {
			return nil, fmt.Errorf("Failed to load data from path (%s), %w", path, err)
		}

		return r, nil

	default:

		fs := data.FS
		r, err := fs.Open(DATA_JSON)

		if err != nil {
			return nil, fmt.Errorf("Failed to load local precompiled data, %w", err)
		}

		return r, nil
	}

}
