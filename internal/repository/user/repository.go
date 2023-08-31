package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/ninashvl/avito-backend-test/internal/store"
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

var (
	ErrSegmentNotFound         = errors.New("segment not found error")
	ErrUserNotFound            = errors.New("user not found error")
	ErrSegmentIsAssignedToUser = errors.New("segment is already assigned to user")
)

func (r *Repo) GetSegmentsByUserID(ctx context.Context, userID int) ([]string, error) {
	query, args, err := squirrel.Select("segment_name").
		From("user_segment").
		Where(squirrel.Eq{"user_id": userID, "deleted_at": nil}).PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("create query GetSegmentsByUserID: %w", err)
	}
	segments := make([]string, 0)
	err = r.executor(ctx).SelectContext(ctx, &segments, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing query GetSegmentsByUserID: %w", err)
	}
	if len(segments) == 0 {
		return nil, ErrUserNotFound
	}
	return segments, nil
}

func (r *Repo) AssignUserSegments(ctx context.Context, userID int64, segments []*AssignedSegment) error {
	q := squirrel.Insert("user_segment").Columns("user_id", "segment_name")
	for _, segment := range segments {
		q = q.Values(userID, segment.SegmentName)
	}
	query, args, err := q.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("create query AssignUserSegments: %w", err)
	}
	_, err = r.executor(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == "23505" {
			return ErrSegmentIsAssignedToUser
		}
		return fmt.Errorf("executing query AssignUserSegments: %w", err)
	}
	return nil
}

func (r *Repo) DeleteUserSegments(ctx context.Context, userID int64, segments []string) error {
	query, args, err := squirrel.Select("count(*)").From("user_segment").
		Where(squirrel.Eq{"user_id": userID, "segment_name": segments, "deleted_at": nil}).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("create query check user segments in DeleteUserSegments: %w", err)
	}
	count := 0
	err = r.executor(ctx).QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return fmt.Errorf("exec query check user segments in DeleteUserSegments: %w", err)
	}
	if count != len(segments) {
		return ErrSegmentNotFound
	}
	query, args, err = squirrel.Update("user_segment").Set("deleted_at", time.Now()).
		Where(squirrel.Eq{"user_id": userID, "segment_name": segments, "deleted_at": nil}).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("create query update user segments in DeleteUserSegments: %w", err)
	}
	_, err = r.executor(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing query update user segments in DeleteUserSegments: %w", err)
	}
	return nil
}
