// Package data provides embedded Library of Congress (LoC) data files.
package data

import (
	"embed"
)

//go:embed *.csv.bz2
var FS embed.FS
