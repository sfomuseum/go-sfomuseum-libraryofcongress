package sqlite

import (
	"context"
	"fmt"
	"github.com/aaronland/go-sqlite"
	"github.com/aaronland/go-sqlite/database"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/lcnaf"
	"github.com/sfomuseum/go-sfomuseum-libraryofcongress/lcsh"
	"net/url"
)

type SQLiteLookup struct {
	libraryofcongress.Lookup
	db *database.SQLiteDatabase
}

func init() {
	ctx := context.Background()
	libraryofcongress.RegisterLookup(ctx, "sqlite", NewSQLiteLookup)
}

func NewSQLiteLookup(ctx context.Context, uri string) (libraryofcongress.Lookup, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	dsn := u.Path

	db, err := database.NewDB(ctx, dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to create database, %w", err)
	}

	exists, err := sqlite.HasTable(ctx, db, "identifiers")

	if err != nil {
		return nil, fmt.Errorf("Failed to determine whether identifiers table exists, %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("Database is missing identifiers table")
	}

	l := &SQLiteLookup{
		db: db,
	}

	return l, nil
}

func (l *SQLiteLookup) Find(ctx context.Context, code string) ([]interface{}, error) {

	conn, err := l.db.Conn()

	if err != nil {
		return nil, fmt.Errorf("Failed to establish database connection, %w", err)
	}

	q := "SELECT id, label, source FROM identifiers WHERE label = ?"

	rows, err := conn.QueryContext(ctx, q, code)

	if err != nil {
		return nil, fmt.Errorf("Failed to query database, %w", err)
	}

	defer rows.Close()

	rsp := make([]interface{}, 0)

	for rows.Next() {

		var id string
		var label string
		var source string

		err := rows.Scan(&id, &label, &source)

		if err != nil {
			return nil, fmt.Errorf("Failed to scan database row, %w", err)
		}

		switch source {
		case "lcnaf":

			r := &lcnaf.NamedAuthority{
				Id:    id,
				Label: label,
			}

			rsp = append(rsp, r)

		case "lcsh":

			r := &lcsh.SubjectHeading{
				Id:    id,
				Label: label,
			}

			rsp = append(rsp, r)

		default:
			return nil, fmt.Errorf("Unsupported source, %s", source)
		}
	}

	err = rows.Close()

	if err != nil {
		return nil, fmt.Errorf("Failed to close database connection, %w", err)
	}

	err = rows.Err()

	if err != nil {
		return nil, fmt.Errorf("Database reported an error, %w", err)
	}

	return rsp, nil
}

func (l *SQLiteLookup) Append(ctx context.Context, data interface{}) error {
	return fmt.Errorf("Not implemented.")
}
