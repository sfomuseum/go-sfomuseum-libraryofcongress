package lcsh

import (
	"fmt"
)

type SubjectHeading struct {
	Id    string `json:"id"`
	Label string `json:"label"`
}

func (sh *SubjectHeading) String() string {
	return fmt.Sprintf("%s %s", sh.Id, sh.Label)
}
