package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/KiberIGOR/tinyurl/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrConflict = errors.New("original URL already exists")

type DB struct {
    mu sync.RWMutex
    db *pgxpool.Pool
}
func NewDB(db *pgxpool.Pool) *DB {
    return &DB{db: db}
}
func (d *DB) Ping(ctx context.Context) error {
    return d.db.Ping(ctx)
}
func (d *DB) Get(ctx context.Context, id string) (string, bool) {
    d.mu.Lock()
    defer d.mu.Unlock()
    originalURL, err := d.get(ctx,id)
    if err != nil {
        return "", false
    }
    return originalURL, true
}

func (d *DB) Save(ctx context.Context, id,originalURL string) (string, error) {
    d.mu.Lock()
    defer d.mu.Unlock()
    _, err := d.get(ctx,id)
    if err == nil {
        // строка нашлась
        return "", ErrAlreadyExist
    }
    if !errors.Is(err, pgx.ErrNoRows) {
        // реальная ошибка БД
        return "", err
    }
    _, err = d.db.Exec(ctx, "INSERT INTO urls (short_url, original_url) VALUES ($1, $2)", id, originalURL)
    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
            var shortURL string
            err2 := d.db.QueryRow(ctx, "SELECT short_url FROM urls WHERE original_url = $1", originalURL).Scan(&shortURL)
            if err2 == nil {
                return shortURL, ErrConflict
            }
            if errors.Is(err2, pgx.ErrNoRows) {
                return "", ErrAlreadyExist
            }
            return "", err2
        }
        return "", err
    }
    return "", nil
}

func (d *DB) MassiveSave(ctx context.Context, MassiveURLs []model.MassiveRequest) error {
	d.mu.Lock()
	defer d.mu.Unlock()
    tx,err := d.db.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)
    
    stmtName := "insert_url"
    _, err = tx.Prepare(ctx, stmtName,"INSERT INTO urls (short_url,original_url) VALUES ($1,$2)")
    if err != nil {
        return err
    }
    for _,item :=range MassiveURLs {
		_, err := tx.Exec(ctx,stmtName, item.ShortURL, item.OriginalURL)
        if err != nil {
            return err
        }
	}
	return tx.Commit(ctx)
}

func (d *DB) get(ctx context.Context, id string) (string, error) {
    row := d.db.QueryRow(ctx,"SELECT original_url FROM urls WHERE short_url = $1", id)
    var originalURL string
    err := row.Scan(&originalURL)
    return originalURL, err
}
 