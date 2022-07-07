package lcnaf

import (
	"testing"
)

func TestNamedAuthority(t *testing.T) {

	na := NamedAuthority{
		Id:    "no2006016666",
		Label: "Paige, Colin",
	}

	if na.String() != "no2006016666 Paige, Colin" {
		t.Fatalf("Invalid stringification")
	}
}
