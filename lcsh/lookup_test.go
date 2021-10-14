package lcsh

import (
	"context"
	"fmt"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress"
	"testing"
	"time"
)

func TestLCSHLookup(t *testing.T) {

	lcsh_tests := map[string]string{
		"Airplanes":                         "sh85002782",
		"Boeing airplanes":                  "sh85015277",
		"Aerial photography in archaeology": "sh85001256",
	}

	ctx := context.Background()

	schemes := []string{
		"lcsh://",
	}

	for _, s := range schemes {

		t1 := time.Now()

		lu, err := libraryofcongress.NewLookup(ctx, s)

		fmt.Printf("Time to load lookup %v\n", time.Since(t1))

		if err != nil {
			t.Fatalf("Failed to create lookup using scheme '%s', %v", s, err)
		}

		for label, lcsh_id := range lcsh_tests {

			results, err := lu.Find(ctx, label)

			if err != nil {
				t.Fatalf("Unable to find '%s' using scheme '%s', %v", label, s, err)
			}

			if len(results) != 1 {
				t.Fatalf("Invalid results for '%s' using scheme '%s'", s, label)
			}

			a := results[0].(*SubjectHeading)

			if a.Id != lcsh_id {
				t.Fatalf("Invalid match for '%s' using scheme '%s', expected '%s' but got '%s'", label, s, lcsh_id, a.Id)
			}
		}
	}
}
