package sqlite

import (
	"compress/bzip2"
	"context"
	_ "database/sql"
	"fmt"
	"github.com/aaronland/go-sqlite"
	"github.com/aaronland/go-sqlite/database"
	loc_database "github.com/sfomuseum/go-libraryofcongress-database"
	loc_sqlite "github.com/sfomuseum/go-libraryofcongress-database/sqlite"
	loc_tables "github.com/sfomuseum/go-libraryofcongress-database/sqlite/tables"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/lcnaf"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/lcsh"
	"github.com/sfomuseum/go-timings"
	"io"
	"os"
)

// NewIdentifiersDatabase() returns a `aaronland/go-sqlite/database.SQLiteDatabase` instance that has a 'identifers'
// table (sfomuseum/go-libraryofcongress-database/sqlite/tables) which has been indexed using the LCSH and LCNAF
// data bundled with `sfomuseum/go-sfomuseum-libraryofcongress`. This is primarily a helper method used by the
// `flysfo:go-sfomuseum-data-filemaker/cmd/merge-filemaker-objects-export` tool.
func NewIndentifiersDatabase(ctx context.Context, dsn string, data_uris map[string]string) (*database.SQLiteDatabase, error) {

	// Some or all of this code should be reconciled with cmd/to-sqlite but not today...
	// (20220708/thisisaaronland)

	sqlite_db, err := database.NewDB(ctx, dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new database, %v", err)
	}

	err = sqlite_db.LiveHardDieFast()

	if err != nil {
		return nil, fmt.Errorf("Failed to enable live hard, die fast settings, %v", err)
	}

	tables := make([]sqlite.Table, 0)

	identifiers_table, err := loc_tables.NewIdentifiersTableWithDatabase(ctx, sqlite_db)

	if err != nil {
		return nil, fmt.Errorf("Failed to create identifiers table, %v", err)
	}

	tables = append(tables, identifiers_table)

	data_sources := make([]*loc_database.Source, 0)

	for source, uri := range data_uris {

		var r io.ReadCloser

		switch source {
		case "lcsh":

			fh, err := lcsh.OpenData(ctx, uri)

			if err != nil {
				return nil, fmt.Errorf("Failed to open %s, %v", uri, err)
			}

			r = fh

		case "lcnaf":

			fh, err := lcnaf.OpenData(ctx, uri)

			if err != nil {
				return nil, fmt.Errorf("Failed to open %s, %v", uri, err)
			}

			r = fh

		default:
			return nil, fmt.Errorf("Unsupported source: %s", source)
		}

		src := &loc_database.Source{
			Label:  source,
			Reader: bzip2.NewReader(r),
		}

		data_sources = append(data_sources, src)
	}

	monitor, err := timings.NewCounterMonitor(ctx, "counter://PT60S")

	if err != nil {
		return nil, fmt.Errorf("Failed to create timings monitor, %v", err)
	}

	monitor.Start(ctx, os.Stdout)
	defer monitor.Stop(ctx)

	err = loc_sqlite.Index(ctx, data_sources, sqlite_db, tables, monitor)

	if err != nil {
		return nil, fmt.Errorf("Failed to index sources, %v", err)
	}

	return sqlite_db, nil
}

// NewIdentifiersLookup() returns a new `libraryofcongress.Lookup` for a `aaronland/go-sqlite/database.SQLiteDatabase` instance
// (identified) by 'dsn' which is produced using the `NewIdentifiersDatabase()` method. This is primarily a helper method used by the
// `flysfo:go-sfomuseum-data-filemaker/cmd/merge-filemaker-objects-export` tool.
func NewIdentifiersLookup(ctx context.Context, dsn string, data_uris map[string]string) (libraryofcongress.Lookup, error) {

	sqlite_db, err := NewIndentifiersDatabase(ctx, dsn, data_uris)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new database, %v", err)
	}

	loc_lookup, err := NewSQLiteLookupWithDatabase(ctx, sqlite_db)

	if err != nil {
		return nil, fmt.Errorf("Failed to create SQLite lookup for LoC data, %v", err)
	}

	return loc_lookup, nil
}
