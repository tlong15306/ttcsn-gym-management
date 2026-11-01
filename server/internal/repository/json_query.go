package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
)

// SQL projections below contain only public fields, never password hashes or file paths.
func jsonOne[T any](ctx context.Context, db queryRower, query string, args ...any) (T, error) {
	var result T
	var raw []byte
	if err := db.QueryRow(ctx, `SELECT to_jsonb(record) FROM (`+query+`) record`, args...).Scan(&raw); err != nil {
		return result, mapDBError(err)
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, err
	}
	return result, nil
}
func jsonList[T any](ctx context.Context, pool *pgxpool.Pool, query string, args []any, f catalog.Filter) ([]T, int, error) {
	var total int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM (`+query+`) record`, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	bound := append(append([]any{}, args...), f.PageSize, (f.Page-1)*f.PageSize)
	rows, err := pool.Query(ctx, `SELECT to_jsonb(record) FROM (`+query+fmt.Sprintf(` ORDER BY 1 LIMIT $%d OFFSET $%d) record`, len(bound)-1, len(bound)), bound...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	result := make([]T, 0)
	for rows.Next() {
		var raw []byte
		var value T
		if err := rows.Scan(&raw); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, 0, err
		}
		result = append(result, value)
	}
	return result, total, rows.Err()
}
func inTransaction[T any](ctx context.Context, pool *pgxpool.Pool, work func(pgx.Tx) (T, error)) (T, error) {
	var zero T
	tx, err := pool.Begin(ctx)
	if err != nil {
		return zero, err
	}
	defer tx.Rollback(ctx)
	result, err := work(tx)
	if err != nil {
		return zero, mapDBError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return zero, mapDBError(err)
	}
	return result, nil
}
