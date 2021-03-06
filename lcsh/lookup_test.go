package lcsh

import (
	"context"
	"fmt"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress"
	_ "gocloud.dev/blob/fileblob"
	"net/url"
	"path/filepath"
	"testing"
	"time"
)

func TestLCSHLookup(t *testing.T) {

	lcsh_tests := map[string][]string{
		"Airplanes":                         []string{"sh85002782"},
		"Boeing airplanes":                  []string{"sh85015277"},
		"Aerial photography in archaeology": []string{"sh85001256"},
		"Cooking":                           []string{"sh2010007517", "sh2010008400"},
		"Aeronautics -- Popular works":      []string{"sh2007100714"}, // SFOM syntax
		"Aeronautics--Popular works":        []string{"sh2007100714"}, // LoC syntax
	}

	ctx := context.Background()

	schemes := []string{
		"lcsh://",
		"lcsh://github",
	}

	rel_path := "../data/lcsh.csv.bz2"
	abs_path, err := filepath.Abs(rel_path)

	if err != nil {
		t.Fatalf("Failed to derive absolute path for %s, %v", rel_path, err)
	}

	file_uri := fmt.Sprintf("lcsh://file%s", abs_path)
	schemes = append(schemes, file_uri)

	// START OF build blob URI

	root := filepath.Dir(abs_path)
	fname := filepath.Base(abs_path)

	fileblob_uri := fmt.Sprintf("file://%s", root)

	v := &url.Values{}
	v.Set("uri", fileblob_uri)

	blob_uri := fmt.Sprintf("lcsh://blob/%s?%s", fname, v.Encode())

	schemes = append(schemes, blob_uri)

	// END OF build blob URI

	for _, s := range schemes {

		t1 := time.Now()

		lu, err := libraryofcongress.NewLookup(ctx, s)

		fmt.Printf("Time to load lookup '%s' %v\n", s, time.Since(t1))

		if err != nil {
			t.Fatalf("Failed to create lookup using scheme '%s', %v", s, err)
		}

		for label, lcsh_ids := range lcsh_tests {

			results, err := lu.Find(ctx, label)

			if err != nil {
				t.Fatalf("Unable to find '%s' using scheme '%s', %v", label, s, err)
			}

			if len(results) != len(lcsh_ids) {
				t.Fatalf("Invalid result count for '%s' using scheme '%s'", s, label)
			}

			for idx, r := range results {

				a := r.(*SubjectHeading)

				if a.Id != lcsh_ids[idx] {
					t.Fatalf("Invalid match for '%s' using scheme '%s', expected '%s' but got '%s'", label, s, lcsh_ids[idx], a.Id)
				}
			}
		}
	}
}
