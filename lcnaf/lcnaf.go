package lcnaf

import (
	"fmt"
)

type NamedAuthority struct {
	Id    string `json:"id"`
	Label string `json:"label"`
}

func (na *NamedAuthority) String() string {
	return fmt.Sprintf("%s %s", na.Id, na.Label)
}
