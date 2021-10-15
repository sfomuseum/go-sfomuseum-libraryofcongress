cli:
	go build -mod vendor -o bin/lookup cmd/lookup/main.go
	go build -mod vendor -o bin/to-elasticsearch cmd/to-elasticsearch/main.go
	go build -mod vendor --tags fts5 -o bin/to-sqlite cmd/to-sqlite/main.go
