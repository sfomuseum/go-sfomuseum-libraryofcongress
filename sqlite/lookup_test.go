package sqlite

import (
	"context"
	"fmt"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/lcnaf"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/lcsh"
	"path/filepath"
	"testing"
)

func TestSQLiteLookup(t *testing.T) {

	ctx := context.Background()

	rel_path := "../fixtures/libraryofcongress.db"
	abs_path, err := filepath.Abs(rel_path)

	if err != nil {
		t.Fatalf("Failed to derive absolute path for %s, %v", rel_path, err)
	}

	lookup_uri := fmt.Sprintf("sqlite://%s", abs_path)

	l, err := libraryofcongress.NewLookup(ctx, lookup_uri)

	if err != nil {
		t.Fatalf("Failed to create new lookup, %v", err)
	}

	lcsh_tests := map[string]string{
		"Grave Creek Watershed (Josephine County, Or.)": "sh00000019",
		"Grave Creek (Josephine County, Or.)":           "sh00000021",
	}

	lcnaf_tests := map[string]string{
		"Neefe, Christian Gottlob, 1748-1798. Veränderungen über den Priestermarsch aus Mozarts Zauberflöte": "no2018099999",
		"Halstenberg, Friedrich": "n2003099999",
	}

	for label, expected_id := range lcsh_tests {

		rsp, err := l.Find(ctx, label)

		if err != nil {
			t.Fatalf("Failed to find '%s', %v", label, err)
		}

		if len(rsp) == 0 {
			t.Fatalf("No results for '%s'", label)
		}

		first := rsp[0].(*lcsh.SubjectHeading)

		if first.Id != expected_id {
			t.Fatalf("Unexpected ID for '%s': %s", label, first.Id)
		}
	}

	for label, expected_id := range lcnaf_tests {

		rsp, err := l.Find(ctx, label)

		if err != nil {
			t.Fatalf("Failed to find '%s', %v", label, err)
		}

		if len(rsp) == 0 {
			t.Fatalf("No results for '%s'", label)
		}

		first := rsp[0].(*lcnaf.NamedAuthority)

		if first.Id != expected_id {
			t.Fatalf("Unexpected ID for '%s': %s", label, first.Id)
		}
	}

}
