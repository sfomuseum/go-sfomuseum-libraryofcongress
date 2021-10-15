package tables

import (
	"context"
	"fmt"
	"github.com/aaronland/go-sqlite"
	_ "log"
)

type IdentifiersTable struct {
	sqlite.Table
	name string
}

func NewIdentifiersTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	t, err := NewIdentifiersTable(ctx)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewIdentifiersTable(ctx context.Context) (sqlite.Table, error) {

	t := IdentifiersTable{
		name: "identifiers",
	}

	return &t, nil
}

func (t *IdentifiersTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *IdentifiersTable) Name() string {
	return t.name
}

func (t *IdentifiersTable) Schema() string {

	schema := `CREATE TABLE %s(
		id TEXT PRIMARY KEY, source TEXT, label TEXT
	);`

	// so dumb...
	return fmt.Sprintf(schema, t.Name())
}

func (t *IdentifiersTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexRow(ctx, db, i.(map[string]string))
}

func (t *IdentifiersTable) IndexRow(ctx context.Context, db sqlite.Database, row map[string]string) error {

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, source, label
		) VALUES (
		?, ?, ?
		)`, t.Name()) // ON CONFLICT DO BLAH BLAH BLAH

	args := []interface{}{
		row["id"],
		row["source"],
		row["label"],
	}

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	tx, err := conn.Begin()

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(args...)

	if err != nil {
		return err
	}

	return tx.Commit()
}
