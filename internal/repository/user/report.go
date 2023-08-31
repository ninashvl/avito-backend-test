package user

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) GetReportDataByUserIDs(ctx context.Context, userIDs []int, from, to *time.Time) ([]*SegmentActivity, error) {
	if len(userIDs) > 0 {
		query, args, err := squirrel.Select("count(DISTINCT user_id)").From("user_segment").
			Where(squirrel.Eq{"user_id": userIDs}).PlaceholderFormat(squirrel.Dollar).ToSql()
		if err != nil {
			return nil, fmt.Errorf("create query check users in GetReportDataByUserIDs: %w", err)
		}
		count := 0
		err = r.executor(ctx).QueryRowContext(ctx, query, args...).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("exec query check users in GetReportDataByUserIDs: %w", err)
		}
		if count != len(userIDs) {
			return nil, ErrUserNotFound
		}
	}

	q := squirrel.Select("user_id", "segment_name", "created_at", "deleted_at").
		From("user_segment").OrderBy("created_at")

	if from != nil || to != nil {
		cond1 := squirrel.And{}
		cond2 := squirrel.And{}
		if from != nil {
			cond1 = append(cond1, squirrel.GtOrEq{"created_at": from})
			cond2 = append(cond2, squirrel.GtOrEq{"deleted_at": from})
		}
		if to != nil {
			cond1 = append(cond1, squirrel.LtOrEq{"created_at": to})
			cond2 = append(cond2, squirrel.LtOrEq{"deleted_at": to})
		}
		q = q.Where(squirrel.Or{cond1, cond2})
	}

	if len(userIDs) > 0 {
		q = q.Where(squirrel.Eq{"user_id": userIDs})
	}
	query, args, err := q.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("create query GetReportDataByUserIDs: %w", err)
	}
	segments := make([]*SegmentActivity, 0)
	err = r.executor(ctx).SelectContext(ctx, &segments, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing query GetReportDataByUserIDs: %w", err)
	}

	return segments, nil
}
