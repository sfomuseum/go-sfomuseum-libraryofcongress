package lcnaf

import (
	"fmt"
)

// type NotFound is a struct for representing missing LCNAF records.
type NotFound struct{ Code string }

// Error() returns a stringified representation of 'e'.
func (e NotFound) Error() string {
	return fmt.Sprintf("named authority '%s' not found", e.Code)
}

// String() returns a stringified representation of 'e'.
func (e NotFound) String() string {
	return e.Error()
}

// type NotFound is a struct for representing LCNAF identifiers that return multiple records.
type MultipleCandidates struct{ Code string }

// Error() returns a stringified representation of 'e'.
func (e MultipleCandidates) Error() string {
	return fmt.Sprintf("Multiple candidates for named authority '%s'", e.Code)
}

// String() returns a stringified representation of 'e'.
func (e MultipleCandidates) String() string {
	return e.Error()
}

// IsNotFound returns a boolean value indicating whether 'e' is of type `NotFound`.
func IsNotFound(e error) bool {

	switch e.(type) {
	case NotFound, *NotFound:
		return true
	default:
		return false
	}
}

// IsMultipleCandidates returns a boolean value indicating whether 'e' is of type `MultipleCandidates`.
func IsMultipleCandidates(e error) bool {

	switch e.(type) {
	case MultipleCandidates, *MultipleCandidates:
		return true
	default:
		return false
	}
}
