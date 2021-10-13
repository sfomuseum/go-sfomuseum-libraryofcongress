package data

import (
	"embed"
)

//go:embed *.csv.bz2
var FS embed.FS
