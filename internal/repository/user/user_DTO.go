package user

import "database/sql"

type AssignedSegment struct {
	SegmentName string
	TTL         sql.NullInt64
}
