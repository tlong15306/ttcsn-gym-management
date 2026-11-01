package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
)

type RoomCoverRepository interface {
	Set(context.Context, int64, []byte, []byte, string, bool, int64, string) (facility.Room, error)
	Read(context.Context, int64, bool) ([]byte, int64, error)
}
type PostgresRoomCoverRepository struct{ pool *pgxpool.Pool }

func NewRoomCoverRepository(pool *pgxpool.Pool) *PostgresRoomCoverRepository {
	return &PostgresRoomCoverRepository{pool}
}
func (r *PostgresRoomCoverRepository) Set(ctx context.Context, id int64, image, thumb []byte, alt string, demo bool, actor int64, rid string) (facility.Room, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (facility.Room, error) {
		var locked int64
		if err := tx.QueryRow(ctx, `SELECT id FROM rooms WHERE id=$1 FOR UPDATE`, id).Scan(&locked); err != nil {
			return facility.Room{}, err
		}
		before, err := jsonOne[facility.Room](ctx, tx, roomSelect+` WHERE id=$1`, id)
		if err != nil {
			return before, err
		}
		_, err = tx.Exec(ctx, `UPDATE rooms SET cover_image=$2,cover_thumbnail=$3,cover_alt_text=$4,cover_is_demo=$5,cover_version=cover_version+1 WHERE id=$1`, id, image, thumb, alt, demo)
		if err != nil {
			return before, err
		}
		after, err := jsonOne[facility.Room](ctx, tx, roomSelect+` WHERE id=$1`, id)
		if err != nil {
			return after, err
		}
		if err = writeAudit(ctx, tx, actor, "room.cover_updated", "room", id, "", before, after, rid); err != nil {
			return after, err
		}
		return after, nil
	})
}
func (r *PostgresRoomCoverRepository) Read(ctx context.Context, id int64, card bool) ([]byte, int64, error) {
	var data []byte
	var version int64
	column := "cover_image"
	if card {
		column = "cover_thumbnail"
	}
	err := r.pool.QueryRow(ctx, `SELECT `+column+`,cover_version FROM rooms WHERE id=$1 AND cover_image IS NOT NULL`, id).Scan(&data, &version)
	return data, version, mapDBError(err)
}
