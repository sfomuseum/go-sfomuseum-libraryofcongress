package lcnaf

import (
	_ "bufio"
	_ "bytes"
	"compress/bzip2"
	"context"
	"fmt"
	"github.com/sfomuseum/go-csvdict"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/data"
	"gocloud.dev/blob"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

var lookup_table *sync.Map
var lookup_idx int64

var lookup_init sync.Once
var lookup_init_err error

type NamedAuthorityLookupFunc func(context.Context)

type NamedAuthorityLookup struct {
	libraryofcongress.Lookup
}

func init() {
	ctx := context.Background()
	libraryofcongress.RegisterLookup(ctx, "lcnaf", NewNamedAuthorityLookup)

	lookup_idx = int64(0)
}

func NewNamedAuthorityLookup(ctx context.Context, uri string) (libraryofcongress.Lookup, error) {

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

		defer r.Close()

		lookup_func := NewNamedAuthorityLookupFuncWithReader(ctx, r)
		return NewNamedAuthorityLookupWithLookupFunc(ctx, lookup_func)

	case "github":

		rsp, err := http.Get(DATA_GITHUB)

		if err != nil {
			return nil, fmt.Errorf("Failed to load remote data from Github, %w", err)
		}

		lookup_func := NewNamedAuthorityLookupFuncWithReader(ctx, rsp.Body)
		return NewNamedAuthorityLookupWithLookupFunc(ctx, lookup_func)

	case "file":

		path := u.Path

		r, err := os.Open(path)

		if err != nil {
			return nil, fmt.Errorf("Failed to load data from path (%s), %w", path, err)
		}

		defer r.Close()

		lookup_func := NewNamedAuthorityLookupFuncWithReader(ctx, r)
		return NewNamedAuthorityLookupWithLookupFunc(ctx, lookup_func)

	/*
	case "sqlite":

		path := u.Path

		db, err := sql.Open("sqlite3", path)

		if err != nil {
			return nil, fmt.Errorf("Failed to open database %s, %w", path, err)
		}
		
		return NewNamedAuthorityLookupWithDB(ctx, db)
	*/
		
	default:

		fs := data.FS
		r, err := fs.Open(DATA_JSON)

		if err != nil {
			return nil, fmt.Errorf("Failed to load local precompiled data, %w", err)
		}

		defer r.Close()

		lookup_func := NewNamedAuthorityLookupFuncWithReader(ctx, r)
		return NewNamedAuthorityLookupWithLookupFunc(ctx, lookup_func)
	}
}

// NewNamedAuthorityLookup will return an `NamedAuthorityLookupFunc` function instance that, when invoked, will populate an `airports.NamedAuthoritysLookup` instance with data stored in `r`.
// `r` will be closed when the `NamedAuthorityLookupFunc` function instance is invoked.
// It is assumed that the data in `r` will be formatted in the same way as the procompiled (embedded) data stored in `data/sfomuseum.json`.
func NewNamedAuthorityLookupFuncWithReader(ctx context.Context, r io.ReadCloser) NamedAuthorityLookupFunc {

	fh := bzip2.NewReader(r)

	csv_r, err := csvdict.NewReader(fh)

	if err != nil {

		lookup_func := func(ctx context.Context) {
			lookup_init_err = err
		}

		return lookup_func
	}

	lookup_func := func(ctx context.Context) {

		table := new(sync.Map)

		for {

			select {
			case <-ctx.Done():
				return
			default:
				// pass
			}

			row, err := csv_r.Read()

			if err == io.EOF {
				break
			}

			if err != nil {
				lookup_init_err = err
				return
			}

			sh := &NamedAuthority{
				Id:    row["id"],
				Label: row["label"],
			}

			err = appendData(ctx, table, sh)

			if err != nil {
				lookup_init_err = err
				return
			}
		}

		lookup_table = table
	}

	return lookup_func
}

// NewNamedAuthorityLookupWithLookupFunc will return an `airports.NamedAuthoritysLookup` instance derived by data compiled using `lookup_func`.
func NewNamedAuthorityLookupWithLookupFunc(ctx context.Context, lookup_func NamedAuthorityLookupFunc) (libraryofcongress.Lookup, error) {

	fn := func() {
		lookup_func(ctx)
	}

	lookup_init.Do(fn)

	if lookup_init_err != nil {
		return nil, lookup_init_err
	}

	l := NamedAuthorityLookup{}
	return &l, nil
}

func (l *NamedAuthorityLookup) Find(ctx context.Context, code string) ([]interface{}, error) {

	pointers, ok := lookup_table.Load(code)

	if !ok {
		return nil, NotFound{code}
	}

	airport := make([]interface{}, 0)

	for _, p := range pointers.([]string) {

		if !strings.HasPrefix(p, "pointer:") {
			return nil, fmt.Errorf("Invalid pointer, '%s'", p)
		}

		row, ok := lookup_table.Load(p)

		if !ok {
			return nil, fmt.Errorf("Invalid pointer, '%s'", p)
		}

		airport = append(airport, row.(*NamedAuthority))
	}

	return airport, nil
}

func (l *NamedAuthorityLookup) Append(ctx context.Context, data interface{}) error {
	return appendData(ctx, lookup_table, data.(*NamedAuthority))
}

func appendData(ctx context.Context, table *sync.Map, data *NamedAuthority) error {

	idx := atomic.AddInt64(&lookup_idx, 1)

	pointer := fmt.Sprintf("pointer:%d", idx)
	table.Store(pointer, data)

	possible_codes := []string{
		// data.Id,
		data.Label,
	}

	for _, code := range possible_codes {

		if code == "" {
			continue
		}

		pointers := make([]string, 0)
		has_pointer := false

		others, ok := table.Load(code)

		if ok {

			pointers = others.([]string)
		}

		for _, dupe := range pointers {

			if dupe == pointer {
				has_pointer = true
				break
			}
		}

		if has_pointer {
			continue
		}

		pointers = append(pointers, pointer)
		table.Store(code, pointers)
	}

	return nil
}
