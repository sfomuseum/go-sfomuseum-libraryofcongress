package lcnaf

import (
	_ "bufio"
	_ "bytes"
	"compress/bzip2"
	"context"
	"fmt"
	"github.com/sfomuseum/go-csvdict"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress"
	"io"
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

	r, err := OpenData(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to open data for '%s', %w", uri, err)
	}

	defer r.Close()

	lookup_func := NewNamedAuthorityLookupFuncWithReader(ctx, r)
	return NewNamedAuthorityLookupWithLookupFunc(ctx, lookup_func)
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

		// START OF hack to account for the difference in syntax between SFOM and LoC

		if strings.Contains(code, " -- ") {
			code = strings.Replace(code, " -- ", "--", -1)
			return l.Find(ctx, code)
		}

		// END OF hack to account for the difference in syntax between SFOM and LoC

		return nil, NotFound{code}
	}

	name_authorities := make([]interface{}, 0)

	for _, p := range pointers.([]string) {

		if !strings.HasPrefix(p, "pointer:") {
			return nil, fmt.Errorf("Invalid pointer, '%s'", p)
		}

		row, ok := lookup_table.Load(p)

		if !ok {
			return nil, fmt.Errorf("Invalid pointer, '%s'", p)
		}

		name_authorities = append(name_authorities, row.(*NamedAuthority))
	}

	return name_authorities, nil
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
