package lcsh

import (
	"fmt"
	"testing"
)

func TestNotFound(t *testing.T) {

	e := NotFound{Code: "1234"}

	if !IsNotFound(e) {
		t.Fatalf("Expected error to be NotFound")
	}

	e2 := fmt.Errorf("Testing")

	if IsNotFound(e2) {
		t.Fatalf("Expected error to not be NotFound")
	}
}

func TestMultipleCandidates(t *testing.T) {

	e := MultipleCandidates{Code: "1234"}

	if !IsMultipleCandidates(e) {
		t.Fatalf("Expected error to be MultipleCandidates")
	}

	e2 := fmt.Errorf("Testing")

	if IsMultipleCandidates(e2) {
		t.Fatalf("Expected error to not be MultipleCandidates")
	}
}
