package time

import (
	errtime "playcards/model/time/errors"
	pbtime "playcards/proto/time"
	"time"
)

type TimeRange struct {
	Start time.Time
	End   time.Time
}

func (tr *TimeRange) ToProto() *pbtime.TimeRange {
	return &pbtime.TimeRange{
		Start: tr.Start.Unix(),
		End:   tr.End.Unix(),
	}
}

func TimeRangeFromProto(tr *pbtime.TimeRange) *TimeRange {
	return &TimeRange{
		Start: time.Unix(tr.Start, 0),
		End:   time.Unix(tr.End, 0),
	}
}

func TimeToProto(t *time.Time) int64 {
	if t == nil || t.IsZero() {
		return 0
	}
	return t.Unix()
}

func TimeFromProto(t int64) *time.Time {
	if t == 0 {
		return nil
	}
	tm := time.Unix(t, 0)
	return &tm
}

func (tr *TimeRange) Valid() error {
	if tr.Start.IsZero() || tr.End.IsZero() {
		return errtime.ErrInvalidTimeRange
	}
	if tr.End.Before(tr.Start) {
		return errtime.ErrInvalidTimeRange
	}
	return nil
}
