// Package lcsh provides methods for working with Library of Congress Subject Heading (LCSH) data.
package lcsh

import (
	"fmt"
)

// SubjectHeading is a struct containing a subset of data for a LCSH record.
type SubjectHeading struct {
	// Id is the unique identifier for this LCSH record.
	Id string `json:"id"`
	// Label is the name (or title) for this LCSH record.
	Label string `json:"label"`
}

// String() returns the a string-ified representation of the record's Id and Label properties.
func (sh *SubjectHeading) String() string {
	return fmt.Sprintf("%s %s", sh.Id, sh.Label)
}
