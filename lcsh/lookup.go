package lcsh

import (
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

type SubjectHeadingLookupFunc func(context.Context)

type SubjectHeadingLookup struct {
	libraryofcongress.Lookup
}

func init() {
	ctx := context.Background()
	libraryofcongress.RegisterLookup(ctx, "lcsh", NewSubjectHeadingLookup)

	lookup_idx = int64(0)
}

func NewSubjectHeadingLookup(ctx context.Context, uri string) (libraryofcongress.Lookup, error) {

	r, err := OpenData(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to open data for '%s', %w", uri, err)
	}

	defer r.Close()

	lookup_func := NewSubjectHeadingLookupFuncWithReader(ctx, r)
	return NewSubjectHeadingLookupWithLookupFunc(ctx, lookup_func)
}

// NewSubjectHeadingLookup will return an `SubjectHeadingLookupFunc` function instance that, when invoked, will populate an `lcsh.SubjectHeadingsLookup` instance with data stored in `r`.
// `r` will be closed when the `SubjectHeadingLookupFunc` function instance is invoked.
// It is assumed that the data in `r` will be formatted in the same way as the procompiled (embedded) data stored in `data/sfomuseum.json`.
func NewSubjectHeadingLookupFuncWithReader(ctx context.Context, r io.ReadCloser) SubjectHeadingLookupFunc {

	fh := bzip2.NewReader(r)

	csv_r, err := csvdict.NewReader(fh)

	if err != nil {

		lookup_func := func(ctx context.Context) {
			lookup_init_err = fmt.Errorf("Failed to create CSV reader, %w", err)
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
				lookup_init_err = fmt.Errorf("Failed to read row, %w", err)
				return
			}

			sh := &SubjectHeading{
				Id:    row["id"],
				Label: row["label"],
			}

			err = appendData(ctx, table, sh)

			if err != nil {
				lookup_init_err = fmt.Errorf("Failed to append row (%s), %w", sh, err)
				return
			}
		}

		lookup_table = table
	}

	return lookup_func
}

// NewSubjectHeadingLookupWithLookupFunc will return an `lcsh.SubjectHeadingsLookup` instance derived by data compiled using `lookup_func`.
func NewSubjectHeadingLookupWithLookupFunc(ctx context.Context, lookup_func SubjectHeadingLookupFunc) (libraryofcongress.Lookup, error) {

	fn := func() {
		lookup_func(ctx)
	}

	lookup_init.Do(fn)

	if lookup_init_err != nil {
		return nil, lookup_init_err
	}

	l := SubjectHeadingLookup{}
	return &l, nil
}

func (l *SubjectHeadingLookup) Find(ctx context.Context, code string) ([]interface{}, error) {

	pointers, ok := lookup_table.Load(code)

	if !ok {
		return nil, NotFound{code}
	}

	subject_headers := make([]interface{}, 0)

	for _, p := range pointers.([]string) {

		if !strings.HasPrefix(p, "pointer:") {
			return nil, fmt.Errorf("Invalid pointer, '%s'", p)
		}

		row, ok := lookup_table.Load(p)

		if !ok {
			return nil, fmt.Errorf("Invalid pointer, '%s'", p)
		}

		subject_headers = append(subject_headers, row.(*SubjectHeading))
	}

	return subject_headers, nil
}

func (l *SubjectHeadingLookup) Append(ctx context.Context, data interface{}) error {
	return appendData(ctx, lookup_table, data.(*SubjectHeading))
}

func appendData(ctx context.Context, table *sync.Map, data *SubjectHeading) error {

	idx := atomic.AddInt64(&lookup_idx, 1)

	pointer := fmt.Sprintf("pointer:%d", idx)
	table.Store(pointer, data)

	possible_codes := []string{
		data.Id,
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
