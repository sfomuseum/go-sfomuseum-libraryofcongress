package lcnaf

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

func TestLCNAFLookup(t *testing.T) {

	lcnaf_tests := map[string]string{
		"Lindbergh, Charles A. (Charles Augustus), 1902-1974": "n79100565",
	}

	ctx := context.Background()

	schemes := []string{
		"lcnaf://",
		"lcnaf://github",
	}

	// START OF build local file URI

	rel_path := "../data/lcnaf.csv.bz2"
	abs_path, err := filepath.Abs(rel_path)

	if err != nil {
		t.Fatalf("Failed to derive absolute path for %s, %v", rel_path, err)
	}

	file_uri := fmt.Sprintf("lcnaf://%s", abs_path)
	schemes = append(schemes, file_uri)

	// END OF build local file URI

	// START OF build blob URI

	root := filepath.Dir(abs_path)
	fname := filepath.Base(abs_path)

	fileblob_uri := fmt.Sprintf("file://%s", root)

	v := &url.Values{}
	v.Set("uri", fileblob_uri)

	blob_uri := fmt.Sprintf("lcnaf://blob/%s?%s", fname, v.Encode())

	schemes = append(schemes, blob_uri)

	// END OF build blob URI

	for _, s := range schemes {

		t1 := time.Now()

		lu, err := libraryofcongress.NewLookup(ctx, s)

		fmt.Printf("Time to load lookup '%s' %v\n", s, time.Since(t1))

		if err != nil {
			t.Fatalf("Failed to create lookup using scheme '%s', %v", s, err)
		}

		for label, lcnaf_id := range lcnaf_tests {

			results, err := lu.Find(ctx, label)

			if err != nil {
				t.Fatalf("Unable to find '%s' using scheme '%s', %v", label, s, err)
			}

			if len(results) != 1 {
				t.Fatalf("Invalid results for '%s' using scheme '%s'", s, label)
			}

			a := results[0].(*NamedAuthority)

			if a.Id != lcnaf_id {
				t.Fatalf("Invalid match for '%s' using scheme '%s', expected '%s' but got '%s'", label, s, lcnaf_id, a.Id)
			}
		}
	}
}
