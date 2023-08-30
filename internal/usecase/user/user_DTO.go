package user

type ChangeUserSegmentDTO struct {
	UserID           int64
	SegmentsToAssign []*AssignedSegment
	SegmentToDelete  []string
}

type AssignedSegment struct {
	SegmentName string
	TTL         int
}
