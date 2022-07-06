// index-es is a command-tool to index the `lcsh` and `lcnaf` data embedded in the `go-sfomuseum-libraryofcongress` package
// in an Elasticsearch index.
package main

import (
	"compress/bzip2"
	"context"
	"flag"
	"github.com/cenkalti/backoff/v4"
	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	loc_database "github.com/sfomuseum/go-libraryofcongress-database"
	loc_elasticsearch "github.com/sfomuseum/go-libraryofcongress-database/elasticsearch"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/data"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/lcnaf"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/lcsh"
	"github.com/sfomuseum/go-timings"
	"log"
	"os"
	"time"
)

func main() {

	es_endpoint := flag.String("elasticsearch-endpoint", "http://localhost:9200", "The Elasticsearch endpoint where data should be indexed.")
	es_index := flag.String("elasticsearch-index", "libraryofcongress", "The Elasticsearch index where data should be stored.")

	workers := flag.Int("workers", 10, "...")

	flag.Parse()

	ctx := context.Background()

	retry := backoff.NewExponentialBackOff()

	es_cfg := es.Config{
		Addresses: []string{*es_endpoint},

		RetryOnStatus: []int{502, 503, 504, 429},
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retry.Reset()
			}
			return retry.NextBackOff()
		},
		MaxRetries: 5,
	}

	es_client, err := es.NewClient(es_cfg)

	if err != nil {
		log.Fatalf("Failed to create ES client, %v", err)
	}

	_, err = es_client.Indices.Create(*es_index)

	if err != nil {
		log.Fatalf("Failed to create index, %v", err)
	}

	// https://github.com/elastic/go-elasticsearch/blob/master/_examples/bulk/indexer.go

	bi_cfg := esutil.BulkIndexerConfig{
		Index:         *es_index,
		Client:        es_client,
		NumWorkers:    *workers,
		FlushInterval: 30 * time.Second,
	}

	bi, err := esutil.NewBulkIndexer(bi_cfg)

	if err != nil {
		log.Fatalf("Failed to create bulk indexer, %v", err)
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

	err = loc_elasticsearch.Index(ctx, data_sources, bi, monitor)

	if err != nil {
		log.Fatalf("Failed to index sources, %v", err)
	}

	err = bi.Close(ctx)

	if err != nil {
		log.Fatalf("Failed to close indexer, %v", err)
	}

}
