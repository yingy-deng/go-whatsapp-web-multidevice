package chatstorage

import (
	"database/sql"
	"strconv"
	"strings"
)

// dialectDB wraps *sql.DB and transparently rebinds SQL placeholders for the
// target driver. For PostgreSQL, '?' is replaced with '$N' (1-indexed).
// For SQLite (and any other driver), queries are passed through unchanged.
type dialectDB struct {
	db     *sql.DB
	driver string
}

func newDialectDB(db *sql.DB, driver string) *dialectDB {
	return &dialectDB{db: db, driver: driver}
}

// rebind converts ? placeholders to $N for PostgreSQL. Other drivers are no-ops.
func (d *dialectDB) rebind(query string) string {
	if d.driver != "postgres" {
		return query
	}
	var b strings.Builder
	n := 1
	for _, ch := range query {
		if ch == '?' {
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(n))
			n++
		} else {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

// blobType returns the SQL type for binary data: BYTEA for PostgreSQL, BLOB otherwise.
func (d *dialectDB) blobType() string {
	if d.driver == "postgres" {
		return "BYTEA"
	}
	return "BLOB"
}

func (d *dialectDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(d.rebind(query), args...)
}

func (d *dialectDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(d.rebind(query), args...)
}

func (d *dialectDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(d.rebind(query), args...)
}

func (d *dialectDB) Begin() (*dialectTx, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	return &dialectTx{tx: tx, driver: d.driver}, nil
}

// dialectTx wraps *sql.Tx and rebinds SQL placeholders like dialectDB.
type dialectTx struct {
	tx     *sql.Tx
	driver string
}

func (t *dialectTx) rebind(query string) string {
	if t.driver != "postgres" {
		return query
	}
	var b strings.Builder
	n := 1
	for _, ch := range query {
		if ch == '?' {
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(n))
			n++
		} else {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func (t *dialectTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(t.rebind(query), args...)
}

func (t *dialectTx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.Query(t.rebind(query), args...)
}

func (t *dialectTx) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRow(t.rebind(query), args...)
}

func (t *dialectTx) Prepare(query string) (*sql.Stmt, error) {
	return t.tx.Prepare(t.rebind(query))
}

func (t *dialectTx) Rollback() error {
	return t.tx.Rollback()
}

func (t *dialectTx) Commit() error {
	return t.tx.Commit()
}
