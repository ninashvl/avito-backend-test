package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	segment "github.com/ninashvl/avito-backend-test/internal/usecase/segment"
	"github.com/ninashvl/avito-backend-test/internal/usecase/user"
	pointer "github.com/ninashvl/avito-backend-test/pkg"
)

type segmentUseCase interface {
	CreateSegment(ctx context.Context, segmentName string) error
	DeleteSegment(ctx context.Context, segmentName string) error
}

type userUseCase interface {
	GetSegmentsByUserID(ctx context.Context, userID int) ([]string, error)
	ChangeSegmentsByUserID(ctx context.Context, changes *user.ChangeUserSegmentDTO) error
	GetReportLink(ctx context.Context, userIDs []int, from, to *time.Time) (string, error)
}

type Handler struct {
	segmentUseCase segmentUseCase
	userUseCase    userUseCase
}

func NewHandler(s segmentUseCase, u userUseCase) *Handler {
	return &Handler{
		s,
		u,
	}
}

func (h Handler) PostSegment(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	req := &CreateSegmentReq{}
	err := eCtx.Bind(req)
	if err != nil {
		return eCtx.NoContent(http.StatusBadRequest)
	}
	err = h.segmentUseCase.CreateSegment(ctx, req.SegmentName)
	if err != nil {
		if errors.Is(err, segment.ErrSegmentIsFound) {
			return eCtx.NoContent(http.StatusConflict)
		}
		return eCtx.NoContent(http.StatusNotFound)
	}
	return eCtx.NoContent(http.StatusCreated)
}

func (h Handler) DeleteSegment(eCtx echo.Context, params DeleteSegmentParams) error {
	ctx := eCtx.Request().Context()
	err := h.segmentUseCase.DeleteSegment(ctx, params.SegmentName)
	if err != nil {
		if errors.Is(err, segment.ErrSegmentNotFound) {
			return eCtx.NoContent(http.StatusNotFound)
		}
		return eCtx.NoContent(http.StatusBadRequest)
	}
	return eCtx.NoContent(http.StatusOK)
}

func (h Handler) GetUserSegments(eCtx echo.Context, params GetUserSegmentsParams) error {
	ctx := eCtx.Request().Context()
	segments, err := h.userUseCase.GetSegmentsByUserID(ctx, params.UserId)
	if err != nil {
		return eCtx.NoContent(http.StatusBadRequest)
	}
	return eCtx.JSON(http.StatusOK, segments)
}

func (h Handler) PostUserSegments(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	req := &ChangeUserSegmentsReq{}
	err := eCtx.Bind(req)
	if err != nil {
		return eCtx.NoContent(http.StatusBadRequest)
	}
	addSegments := make([]*user.AssignedSegment, 0, len(req.AddSegments))
	for _, seg := range req.AddSegments {
		addSegments = append(addSegments, &user.AssignedSegment{
			SegmentName: seg.SegmentName,
			TTL:         int(pointer.Value(seg.Ttl)),
		})
	}
	err = h.userUseCase.ChangeSegmentsByUserID(ctx, &user.ChangeUserSegmentDTO{
		UserID:           int64(req.UserId),
		SegmentsToAssign: addSegments,
		SegmentToDelete:  req.DeleteSegments,
	})
	if err != nil {
		return eCtx.NoContent(http.StatusBadRequest)
	}
	return eCtx.NoContent(http.StatusOK)
}

func (h Handler) PostReport(eCtx echo.Context) error {
	ctx := eCtx.Request().Context()
	req := &CreateReportReq{}
	err := eCtx.Bind(req)
	if err != nil {
		fmt.Println(err)
		return eCtx.NoContent(http.StatusBadRequest)
	}
	var (
		from, to *time.Time
	)
	if req.From != nil && req.To != nil {
		if req.From.Unix() <= req.To.Unix() {
			return eCtx.NoContent(http.StatusBadRequest)
		}
	}
	if req.From != nil {
		from = &req.From.Time
	}
	if req.To != nil {
		from = &req.To.Time
	}
	link, err := h.userUseCase.GetReportLink(ctx, pointer.Value(req.UserIds), from, to)
	if err != nil {
		fmt.Println(err)
		return eCtx.NoContent(http.StatusConflict)
	}
	return eCtx.JSON(http.StatusCreated, &CreateReportResp{Link: link})
}
