package segment

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/ninashvl/avito-backend-test/internal/store"
	"time"
)

var (
	ErrSegmentNotFound = errors.New("segment not found")
	ErrSegmentIsFound  = errors.New("segment is found")
	segmentIsDeleted   = errors.New("segment is deleted")
)

type Repo struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) executor(ctx context.Context) store.Executor {
	tx := store.TxFromContext(ctx)
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *Repo) CreateSegment(ctx context.Context, segmentName string) error {
	query, args, err := squirrel.Insert("segments").
		Columns("name").
		Values(segmentName).
		Suffix("ON CONFLICT (name) DO UPDATE SET created_at = NOW(), deleted_at = NULL WHERE segments.deleted_at IS NOT NULL").
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("create query CreateSegment: %w", err)
	}
	res, err := r.executor(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing query CreateSegment: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("executing query CreateSegment rows: %w", err)
	}
	if rowsAffected == 0 {
		return ErrSegmentIsFound
	}
	return nil
}

func (r *Repo) DeleteSegment(ctx context.Context, segmentName string) error {
	query, args, err := squirrel.Update("segments").
		Set("deleted_at", time.Now()).
		Where(squirrel.Eq{"name": segmentName, "deleted_at": nil}).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("create query DeleteSegment: %w", err)
	}
	rows, err := r.executor(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing query DeleteSegment: %w", err)
	}
	rowsAffected, err := rows.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected DeleteSegment error: %w", err)
	}
	if rowsAffected == 0 {
		return ErrSegmentNotFound
	}
	return nil
}
