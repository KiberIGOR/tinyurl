package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrConflict = errors.New("original URL already exists")
var ErrNoContent = errors.New("no content")

type DB struct {
	db *pgxpool.Pool
}

func NewDB(db *pgxpool.Pool) *DB {
	return &DB{db: db}
}

func (d *DB) Ping(ctx context.Context) error {
	return d.db.Ping(ctx)
}

func (d *DB) Get(ctx context.Context, id string) (string, bool) {
	var originalURL string
	err := d.db.QueryRow(ctx, "SELECT original_url FROM urls WHERE short_url = $1", id).Scan(&originalURL)
	if err != nil {
		return "", false
	}
	return originalURL, true
}

func (d *DB) Save(ctx context.Context, id, originalURL string, userID int) (string, error) {
	_, err := d.db.Exec(ctx, "INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)", id, originalURL, userID)
	if err == nil {
		return "", nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != pgerrcode.UniqueViolation {
		return "", err
	}

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

func (d *DB) BatchSave(ctx context.Context, entries []URLEntry, userID int) (BatchSaveResult, error) {
	result := BatchSaveResult{
		Existing: make(map[string]string),
	}

	tx, err := d.db.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)

	const insertSQL = `
		INSERT INTO urls (short_url, original_url, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (short_url) DO NOTHING`

	for _, item := range entries {
		tag, err := tx.Exec(ctx, insertSQL, item.ShortURL, item.OriginalURL, userID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				var existingShort string
				if err2 := tx.QueryRow(ctx,
					"SELECT short_url FROM urls WHERE original_url = $1",
					item.OriginalURL,
				).Scan(&existingShort); err2 == nil {
					result.Existing[item.OriginalURL] = existingShort
					continue
				}
			}
			return result, err
		}
		if tag.RowsAffected() == 0 {
			result.Retries = append(result.Retries, item)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return result, err
	}
	return result, nil
}

func (d *DB) GetURLsByUserID(ctx context.Context, userID int) ([]URLEntry, error) {
	rows, err := d.db.Query(ctx, "SELECT short_url, original_url FROM urls WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []URLEntry
	for rows.Next() {
		var v URLEntry
		if err := rows.Scan(&v.ShortURL, &v.OriginalURL); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, ErrNoContent
	}
	return result, nil
}

func (d *DB) GetLastID(ctx context.Context) int {
	var id int
	err := d.db.QueryRow(ctx, "SELECT nextval('user_id_seq')").Scan(&id)
	if err != nil {
		return 1
	}
	return id
}
