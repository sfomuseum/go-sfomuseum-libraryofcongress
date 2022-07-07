// Package lcnaf provides methods for working with Library of Congress Named Authority File (LCNAF) data.
package lcnaf

import (
	"fmt"
)

// NamedAuthority is a struct containing a subset of data for a LCNAF record.
type NamedAuthority struct {
	// Id is the unique identifier for this LCNAF record.
	Id string `json:"id"`
	// Label is the name (or title) for this LCNAF record.
	Label string `json:"label"`
}

// String() returns the a string-ified representation of the record's Id and Label properties.
func (na *NamedAuthority) String() string {
	return fmt.Sprintf("%s %s", na.Id, na.Label)
}
