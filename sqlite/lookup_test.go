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

	lcsh_tests := map[string][]string{
		"Grave Creek Watershed (Josephine County, Or.)": []string{"sh00000019"},
		"Grave Creek (Josephine County, Or.)":           []string{"sh00000021"},
		"Cooking":                                       []string{"sh2010007517", "sh2010008400"},
		"Aeronautics -- Popular works":                  []string{"sh2007100714"}, // SFOM syntax
		"Aeronautics--Popular works":                    []string{"sh2007100714"}, // LoC syntax
	}

	lcnaf_tests := map[string][]string{
		"Neefe, Christian Gottlob, 1748-1798. Veränderungen über den Priestermarsch aus Mozarts Zauberflöte": []string{"no2018099999"},
		"Halstenberg, Friedrich": []string{"n2003099999"},
	}

	for label, expected_ids := range lcsh_tests {

		rsp, err := l.Find(ctx, label)

		if err != nil {
			t.Fatalf("Failed to find '%s', %v", label, err)
		}

		if len(rsp) != len(expected_ids) {
			t.Fatalf("No results for '%s'", label)
		}

		for idx, r := range rsp {
			sh := r.(*lcsh.SubjectHeading)

			if sh.Id != expected_ids[idx] {
				t.Fatalf("Unexpected ID for '%s': %s", label, sh.Id)
			}
		}
	}

	for label, expected_ids := range lcnaf_tests {

		rsp, err := l.Find(ctx, label)

		if err != nil {
			t.Fatalf("Failed to find '%s', %v", label, err)
		}

		if len(rsp) != len(expected_ids) {
			t.Fatalf("No results for '%s'", label)
		}

		for idx, r := range rsp {

			na := r.(*lcnaf.NamedAuthority)

			if na.Id != expected_ids[idx] {
				t.Fatalf("Unexpected ID for '%s': %s", label, na.Id)
			}
		}
	}

}
