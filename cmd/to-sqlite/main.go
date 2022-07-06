package main

import (
	"compress/bzip2"
	"context"
	_ "database/sql"
	"flag"
	"github.com/aaronland/go-sqlite"
	"github.com/aaronland/go-sqlite/database"
	loc_database "github.com/sfomuseum/go-libraryofcongress-database"
	loc_sqlite "github.com/sfomuseum/go-libraryofcongress-database/sqlite"
	loc_tables "github.com/sfomuseum/go-libraryofcongress-database/sqlite/tables"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/data"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/lcnaf"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/lcsh"
	"github.com/sfomuseum/go-timings"
	"log"
	"os"
	"time"
)

func main() {

	dsn := flag.String("dsn", "libraryofcongress.db", "")

	flag.Parse()

	ctx := context.Background()

	sqlite_db, err := database.NewDB(ctx, *dsn)

	if err != nil {
		log.Fatalf("Failed to create new database, %v", err)
	}

	err = sqlite_db.LiveHardDieFast()

	if err != nil {
		log.Fatalf("Failed to enable live hard, die fast settings, %v", err)
	}

	search_table, err := loc_tables.NewSearchTableWithDatabase(ctx, sqlite_db)

	if err != nil {
		log.Fatalf("Failed to create search table, %v", err)
	}

	tables := []sqlite.Table{
		search_table,
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
