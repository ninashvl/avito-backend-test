package user

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"strconv"
	"time"
)

const (
	AssignedSegmentActivityType = "assigned"
	DeletedSegmentActivityType  = "deleted"
)

type ReportRaw struct {
	UserID       int64
	Segment      string
	ActivityType string
	Timestamp    time.Time
}

func (u *UseCase) GetReportLink(ctx context.Context, userIDs []int, from, to *time.Time) (string, error) {
	data, err := u.userRepo.GetReportDataByUserIDs(ctx, userIDs, from, to)
	if err != nil {
		return "", err
	}

	report := make([]*ReportRaw, 0)
	for _, v := range data {
		if from == nil || (from != nil && v.CreatedAt.Unix() > from.Unix()) {
			report = append(report, &ReportRaw{
				UserID:       v.UserID,
				Segment:      v.SegmentName,
				ActivityType: AssignedSegmentActivityType,
				Timestamp:    v.CreatedAt,
			})
		}
		if v.DeletedAt.Valid {
			if to == nil || (to != nil && v.DeletedAt.Valid && v.DeletedAt.Time.Unix() < to.Unix()) {
				report = append(report, &ReportRaw{
					UserID:       v.UserID,
					Segment:      v.SegmentName,
					ActivityType: DeletedSegmentActivityType,
					Timestamp:    v.CreatedAt,
				})
			}
		}

	}

	sort.Slice(report, func(i, j int) bool {
		return report[i].Timestamp.Unix() < report[j].Timestamp.Unix()
	})

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	_ = writer.Write([]string{"userIDs", "segments", "activityTypes", "Datetime"})
	for _, r := range report {
		row := []string{strconv.FormatInt(r.UserID, 10), r.Segment, r.ActivityType, r.Timestamp.Format(time.DateTime)}
		_ = writer.Write(row)
	}
	writer.Flush()

	link, err := u.reportRepo.SaveReportFile(ctx, buffer)
	if err != nil {
		return "", fmt.Errorf("saving report error: %w", err)
	}

	return link, nil
}
