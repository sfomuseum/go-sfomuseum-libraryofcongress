module github.com/sfomuseum/go-sfomuseum-libraryofcongress

go 1.16

// Pin to elastic/go-elasticsearch/v7 v7.13.0 because later versions
// don't work with AWS Elasticsearch anymore. Sigh...

require (
	github.com/aaronland/go-roster v1.0.0
	github.com/aaronland/go-sqlite v0.2.0
	github.com/sfomuseum/go-csvdict v1.0.0
	github.com/sfomuseum/go-libraryofcongress-database v0.0.3
	github.com/sfomuseum/go-timings v1.0.0
	gocloud.dev v0.25.0

)
