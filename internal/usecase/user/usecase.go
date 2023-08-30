package user

import (
	"context"
	"database/sql"
	repo "github.com/ninashvl/avito-backend-test/internal/repository/user"
	"github.com/ninashvl/avito-backend-test/internal/store"
)

type userRepository interface {
	GetSegmentsByUserID(ctx context.Context, userID int) ([]string, error)
	DeleteUserSegments(ctx context.Context, userID int64, segments []string) error
	AssignUserSegments(ctx context.Context, userID int64, segments []*repo.AssignedSegment) error
}

type UseCase struct {
	Repo userRepository
	txtr store.Transactor
}

func NewUseCase(r userRepository, txtr store.Transactor) *UseCase {
	return &UseCase{r, txtr}
}

func (u *UseCase) GetSegmentsByUserID(ctx context.Context, userID int) ([]string, error) {
	segments, err := u.Repo.GetSegmentsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return segments, nil
}

func (u *UseCase) ChangeSegmentsByUserID(ctx context.Context, changes *ChangeUserSegmentDTO) error {
	err := u.txtr.RunInTx(ctx, func(ctx context.Context) error {
		if len(changes.SegmentToDelete) != 0 {
			err := u.Repo.DeleteUserSegments(ctx, changes.UserID, changes.SegmentToDelete)
			if err != nil {
				return err
			}
		}
		if len(changes.SegmentsToAssign) != 0 {
			segments := make([]*repo.AssignedSegment, 0, len(changes.SegmentsToAssign))
			for _, v := range changes.SegmentsToAssign {
				segments = append(segments, &repo.AssignedSegment{
					SegmentName: v.SegmentName,
					TTL:         sql.NullInt64{},
				})
			}
			err := u.Repo.AssignUserSegments(ctx, changes.UserID, segments)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
