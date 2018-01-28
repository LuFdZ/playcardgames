package page

import (
	mdtime "playcards/model/time"
	pbpage "playcards/proto/page"
	"playcards/utils/db"
	"time"

	"github.com/jinzhu/gorm"
)

const (
	MinPageSize = 1
	MaxPageSize = 20

	MinTimeDuration = 0
	MaxTimeDuration = time.Hour * 24
)

type QueryFunc func(*gorm.DB) *gorm.DB

type PageOption struct {
	Page     int32
	PageSize int32
	Alias    string
	Sum      bool
	Count    bool
	StartAt  *time.Time
	EndAt    *time.Time
}

type PageReply struct {
	PageNow   int32
	PageTotal int32
}

func (pr *PageReply) ToProto() *pbpage.PageReply {
	return &pbpage.PageReply{
		PageNow:   pr.PageNow,
		PageTotal: pr.PageTotal,
	}
}

func (o *PageOption) RowOffset() int32 {
	return o.Page * o.PageSize
}

func (o *PageOption) BindTimeBetween(tx *gorm.DB) *gorm.DB {
	if o.StartAt != nil {
		tx = tx.Where(o.Alias+"created_at >= ?", o.StartAt)
	}

	if o.EndAt != nil {
		tx = tx.Where(o.Alias+"created_at <= ?", o.EndAt)
	}
	return tx
}

func (o *PageOption) BindPage(tx *gorm.DB) *gorm.DB {
	return tx.Offset(o.RowOffset()).Limit(int(o.PageSize))
}

func (o *PageOption) Find(tx *gorm.DB, out interface{}) (int64, *gorm.DB) {
	tx = o.BindTimeBetween(tx)

	rows := int64(0)
	if o.Count {
		var rstx *gorm.DB
		rows, rstx = db.RecordCount(tx)
		_, err := db.FoundRecord(rstx.Error)
		if err != nil {
			return rows, rstx
		}
	}

	tx = o.BindPage(tx)
	return rows, tx.Find(out)
}

func (o *PageOption) Scan(tx *gorm.DB, out interface{}) (int64, *gorm.DB) {
	tx = o.BindTimeBetween(tx)

	rows := int64(0)
	if o.Count {
		var rstx *gorm.DB
		rows, rstx = db.RecordCount(tx)
		_, err := db.FoundRecord(rstx.Error)
		if err != nil {
			return rows, rstx
		}
	}

	tx = o.BindPage(tx)
	return rows, tx.Scan(out)
}

func (o *PageOption) SumScan(tx *gorm.DB, sumq string,
	sum, out interface{}) *gorm.DB {

	tx = o.BindTimeBetween(tx)

	s := ""
	if o.Sum {
		s = "," + sumq
	}

	if o.Count {
		s = "count(0) as `count`" + s
	}

	ctx := tx.Select(s).Scan(sum)
	_, err := db.FoundRecord(ctx.Error)
	if err != nil {
		return ctx
	}

	tx = o.BindPage(tx)
	return tx.Scan(out)
}

func (p *PageOption) CheckPageSize() {
	size := p.PageSize
	if size > MaxPageSize {
		size = MaxPageSize
	}

	if size < MinPageSize {
		size = MinPageSize
	}

	p.PageSize = size
}

func NewPageOption() *PageOption {
	return &PageOption{
		Page:     0,
		Count:    true,
		PageSize: MaxPageSize,
	}
}

func PageOptionFromProto(p *pbpage.PageOption) *PageOption {
	r := NewPageOption()
	if p == nil {
		r.CheckPageSize()
		return r
	}

	r.Sum = p.Sum
	r.Page = p.Page
	r.PageSize = p.PageSize
	r.CheckPageSize()

	ct := p.Time
	if ct == nil {
		return r
	}

	if ct.Start > 0 {
		r.StartAt = mdtime.TimeFromProto(ct.Start)
	}

	if ct.End > 0 {
		r.EndAt = mdtime.TimeFromProto(ct.End)
	}
	return r
}
