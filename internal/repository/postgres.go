package repository

import (
	"context"
	"database/sql"
)

type DB struct {
    db *sql.DB
}
func NewDB(db *sql.DB) *DB {
    return &DB{db: db}
}
func (d *DB) Ping(ctx context.Context) error {
    return d.db.PingContext(ctx)
}