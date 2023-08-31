package user

import (
	"database/sql"
	"time"
)

type AssignedSegment struct {
	SegmentName string
	TTL         sql.NullInt64
}

type SegmentActivity struct {
	UserID      int64        `db:"user_id"`
	SegmentName string       `db:"segment_name"`
	CreatedAt   time.Time    `db:"created_at"`
	DeletedAt   sql.NullTime `db:"deleted_at"`
}
