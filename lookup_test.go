package libraryofcongress

import (
	"context"
	"testing"
)

type TestLookup struct {
	Lookup
}

func (t *TestLookup) Find(ctx context.Context, id string) ([]interface{}, error) {
	return []interface{}{"found"}, nil
}

func (t *TestLookup) Append(ctx context.Context, rec interface{}) error {
	return nil
}

func NewTestLookup(ctx context.Context, uri string) (Lookup, error) {

	t := &TestLookup{}
	return t, nil
}

func TestTestLookup(t *testing.T) {

	ctx := context.Background()

	err := RegisterLookup(ctx, "test", NewTestLookup)

	if err != nil {
		t.Fatalf("Failed to register TestLookup")
	}

	tl, err := NewLookup(ctx, "test://")

	if err != nil {
		t.Fatalf("Failed to create test lookup, %v", err)
	}

	err = tl.Append(ctx, "testing")

	if err != nil {
		t.Fatalf("Failed to append record")
	}

	v, err := tl.Find(ctx, "testing")

	if err != nil {
		t.Fatalf("Failed to perform lookup")
	}

	if v[0].(string) != "found" {
		t.Fatalf("Unexpected value: %v", v)
	}
}
