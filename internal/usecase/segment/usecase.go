package segment

import (
	"context"
	"errors"
	"fmt"

	repo "github.com/ninashvl/avito-backend-test/internal/repository/segment"
)

var (
	ErrSegmentIsFound  = errors.New("segment is found")
	ErrSegmentNotFound = errors.New("segment not found")
)

type segmentRepository interface {
	CreateSegment(ctx context.Context, segmentName string) error
	DeleteSegment(ctx context.Context, segmentName string) error
}

type UseCase struct {
	Repo segmentRepository
}

func NewUseCase(r segmentRepository) *UseCase {
	return &UseCase{r}
}

func (s *UseCase) CreateSegment(ctx context.Context, segmentName string) error {
	err := s.Repo.CreateSegment(ctx, segmentName)
	if err != nil {
		if errors.Is(err, repo.ErrSegmentIsFound) {
			return ErrSegmentIsFound
		}
		return fmt.Errorf("repo CreateSegment error: %w", err)
	}
	return nil
}

func (s *UseCase) DeleteSegment(ctx context.Context, segmentName string) error {
	err := s.Repo.DeleteSegment(ctx, segmentName)
	if err != nil {
		if errors.Is(err, repo.ErrSegmentNotFound) {
			return ErrSegmentNotFound
		}
		return err
	}
	return nil
}
