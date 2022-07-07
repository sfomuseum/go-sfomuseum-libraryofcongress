package main

import (
	"compress/bzip2"
	"context"
	_ "database/sql"
	"flag"
	"github.com/aaronland/go-sqlite"
	"github.com/aaronland/go-sqlite/database"
	// START OF this is a bit of a hot mess right now
	loc_database "github.com/sfomuseum/go-libraryofcongress-database"
	loc_sqlite "github.com/sfomuseum/go-libraryofcongress-database/sqlite"
	loc_tables "github.com/sfomuseum/go-sfomuseum-libraryofcongress/sqlite/tables"
	// END OF this is a bit of a hot mess right now
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/data"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/lcnaf"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/lcsh"
	"github.com/sfomuseum/go-timings"
	"log"
	"os"
	"time"
)

func main() {

	index_identifiers := flag.Bool("identifiers", true, "Index the identifiers tables.")
	index_search := flag.Bool("search", false, "Index the search table.")
	index_all := flag.Bool("all", false, "Index all tables.")

	dsn := flag.String("dsn", "libraryofcongress.db", "The output path for the new SQLite database.")

	flag.Parse()

	if *index_all {
		*index_identifiers = true
		*index_search = true
	}

	ctx := context.Background()

	sqlite_db, err := database.NewDB(ctx, *dsn)

	if err != nil {
		log.Fatalf("Failed to create new database, %v", err)
	}

	err = sqlite_db.LiveHardDieFast()

	if err != nil {
		log.Fatalf("Failed to enable live hard, die fast settings, %v", err)
	}

	tables := make([]sqlite.Table, 0)

	if *index_identifiers {

		identifiers_table, err := loc_tables.NewIdentifiersTableWithDatabase(ctx, sqlite_db)

		if err != nil {
			log.Fatalf("Failed to create identifiers table, %v", err)
		}

		tables = append(tables, identifiers_table)
	}

	if *index_search {

		search_table, err := loc_tables.NewSearchTableWithDatabase(ctx, sqlite_db)

		if err != nil {
			log.Fatalf("Failed to create search table, %v", err)
		}

		tables = append(tables, search_table)
	}

	//

	data_sources := make([]*loc_database.Source, 0)

	data_paths := map[string]string{
		"lcsh":  lcsh.DATA_JSON,
		"lcnaf": lcnaf.DATA_JSON,
	}

	for source, path := range data_paths {

		r, err := data.FS.Open(path)

		if err != nil {
			log.Fatalf("Failed to open %s, %v", path, err)
		}

		src := &loc_database.Source{
			Label:  source,
			Reader: bzip2.NewReader(r),
		}

		data_sources = append(data_sources, src)
	}

	d := time.Second * 60
	monitor, err := timings.NewCounterMonitor(ctx, d)

	if err != nil {
		log.Fatalf("Failed to create timings monitor, %v", err)
	}

	monitor.Start(ctx, os.Stdout)
	defer monitor.Stop(ctx)

	err = loc_sqlite.Index(ctx, data_sources, sqlite_db, tables, monitor)

	if err != nil {
		log.Fatalf("Failed to index sources, %v", err)
	}

}
