package lcnaf

import (
	"context"
	"fmt"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress"
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
	}

	for _, s := range schemes {

		t1 := time.Now()

		lu, err := libraryofcongress.NewLookup(ctx, s)

		fmt.Printf("Time to load lookup %v\n", time.Since(t1))

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
