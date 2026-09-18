package repository

import (
	"context"
	"database/sql"
	"errors"
	"sync"
)

type DB struct {
    mu sync.RWMutex
    db *sql.DB
}
func NewDB(db *sql.DB) *DB {
    return &DB{db: db}
}
func (d *DB) Ping(ctx context.Context) error {
    return d.db.PingContext(ctx)
}
func (d *DB) Get(id string) (string, bool) {
    d.mu.Lock()
    defer d.mu.Unlock()
    originalURL, err := d.get(id)
    if err != nil {
        return "", false
    }
    return originalURL, true
}

func (d *DB) Save(id,originalURL string) (error) {
    d.mu.Lock()
    defer d.mu.Unlock()
    _, err := d.get(id)
    if err == nil {
        // строка нашлась
        return ErrAlreadyExist
    }
    if !errors.Is(err, sql.ErrNoRows) {
        // реальная ошибка БД
        return err
    }
    _, err = d.db.ExecContext(context.Background(), "INSERT INTO urls (short_url,original_url) VALUES ($1,$2)",id,originalURL)
    if err != nil {
        return err
    }
    return nil
}

func (d *DB) get(id string) (string, error) {
    row := d.db.QueryRowContext(context.Background(),"SELECT original_url FROM urls WHERE short_url = $1", id)
    var originalURL string
    err := row.Scan(&originalURL)
    return originalURL, err
}
 