package db

import (
	mdpage "playcards/model/page"
	"playcards/model/room/enum"
	enumr "playcards/model/room/enum"
	errr "playcards/model/room/errors"
	mdr "playcards/model/room/mod"
	"playcards/utils/db"
	"playcards/utils/errors"

	"github.com/Masterminds/squirrel"
	"github.com/jinzhu/gorm"
	"fmt"
	"playcards/utils/log"
)

func CreateRoom(tx *gorm.DB, r *mdr.Room) error {
	now := gorm.NowFunc()
	r.CreatedAt = &now
	r.UpdatedAt = &now
	if err := tx.Create(r).Error; err != nil {
		return errors.Internal("create room failed", err)
	}
	return nil
}

func UpdateRoom(tx *gorm.DB, r *mdr.Room) (*mdr.Room, error) {
	now := gorm.NowFunc()
	room := &mdr.Room{
		Users:          r.Users,
		Status:         r.Status,
		Giveup:         r.Giveup,
		RoundNow:       r.RoundNow,
		StartMaxNumber: r.StartMaxNumber,
		UpdatedAt:      &now,
		GiveupAt:       r.GiveupAt,
		GameParam:      r.GameParam,
	}
	if err := tx.Model(r).Updates(room).Error; err != nil {
		return nil, errors.Internal("update room failed", err)
	}
	return r, nil
}

func CreateSpecialGame(tx *gorm.DB, gr *mdr.PlayerSpecialGameRecord) error {
	now := gorm.NowFunc()
	gr.CreatedAt = &now
	gr.UpdatedAt = &now
	if err := tx.Create(gr).Error; err != nil {
		return errors.Internal("create player special game record failed", err)
	}
	return nil
}

func PageSpecialGameList(tx *gorm.DB, gr *mdr.PlayerSpecialGameRecord, page *mdpage.PageOption,
) ([]*mdr.PlayerSpecialGameRecord, int64, error) {
	var out []*mdr.PlayerSpecialGameRecord
	rows, rtx := page.Find(tx.Where(gr).
		Order("created_at desc").Find(&out), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page room result failed", rtx.Error)
	}
	return out, rows, nil
}

func SetRoomFlage(tx *gorm.DB, clubid int32, rid int32) error {
	var mr mdr.Room
	if err := tx.Model(mr).Where("room_id = ? and club_id = ?", rid, clubid).UpdateColumn("flag", enumr.RoomFlag).
		Error; err != nil {
		return errors.Internal("set room flag failed", err)
	}
	return nil
}

func GetRoomsByStatus(tx *gorm.DB, status int32) ([]*mdr.Room, error) {
	var (
		out []*mdr.Room
	)
	if err := tx.Where("status = ?", status).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errr.ErrRoomNotExisted
	}
	return out, nil
}

func GetRoomsByStatusAndGameType(tx *gorm.DB, status int32,
	GameType int32) ([]*mdr.Room, error) {
	var (
		out []*mdr.Room
	)
	if err := tx.Where("status = ? and game_type = ?", status, GameType).
		Order("created_at").Find(&out).Error; err != nil {
		return nil, errr.ErrRoomNotExisted
	}
	return out, nil
}

func GetRoomByID(tx *gorm.DB, rid int32) (*mdr.Room, error) {
	var (
		out mdr.Room
	)
	out.RoomID = rid
	found, err := db.FoundRecord(tx.Find(&out).Error)
	if err != nil {
		log.Err("get room by room_id err:%v", err)
		return nil, errors.Internal("get room failed", err)
	}

	if !found {
		return nil, errr.ErrRoomNotExisted
	}
	return &out, nil
}

func BatchUpdate(tx *gorm.DB, status int32, ids []int32) error {
	sql, param, _ := squirrel.Update(enum.RoomTableName).
		Set("status", status).
		Where("room_id in (?)", ids).ToSql()
	err := tx.Exec(sql, param...).Error
	if err != nil {
		return errors.Internal("set room finish failed", err)
	}
	return nil
}

func GetRoomsByStatusArrayAndGameType(tx *gorm.DB, status []int32,
	GameType int32) ([]*mdr.Room, error) {
	var (
		out []*mdr.Room
	)
	if err := tx.Where("status in (?) and game_type = ?", status, GameType).
		Order("created_at").Find(&out).Error; err != nil {
		return nil, errr.ErrRoomNotExisted
	}
	return out, nil
}

func CreateFeedback(tx *gorm.DB, fb *mdr.Feedback) (*mdr.Feedback, error) {
	if err := tx.Create(fb).Error; err != nil {
		return nil, errors.Internal("create feed back failed", err)
	}
	return fb, nil
}

func PageFeedbackList(tx *gorm.DB, page *mdpage.PageOption,
	fb *mdr.Feedback) ([]*mdr.Feedback, int64, error) {
	var out []*mdr.Feedback
	rows, rtx := page.Find(tx.Model(fb).Order("created_at desc").
		Where(fb), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page feed back failed", rtx.Error)
	}
	//fmt.Printf("Page Feedback List:%v", out[0])
	return out, rows, nil
}

func CreatePlayerRoom(tx *gorm.DB, r *mdr.PlayerRoom) error {
	if err := tx.Create(r).Error; err != nil {
		return errors.Internal("create player room failed", err)
	}
	return nil
}

func DeleteAll(tx *gorm.DB) error {
	tx.Where(" game_type = 1001 ").Delete(mdr.Room{})
	return nil
}

func PageRoomResultList(tx *gorm.DB, uid int32, gtype int32, page *mdpage.PageOption,
) ([]*mdr.Room, int64, error) {
	var out []*mdr.Room
	//rows, rtx := page.Find(tx.Model(r).Order("created_at desc").
	//	Where(r), &out)
	//if rtx.Error != nil {
	//	return nil, 0, errors.Internal("page notice failed", rtx.Error)
	//}
	sqlstr := " game_type =? and round_now >1 and room_id in (select room_id from player_rooms where user_id = ?)"
	rows, rtx := page.Find(tx.Where(sqlstr, gtype, uid).
		Order("created_at desc").Find(&out), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page room result failed", rtx.Error)
	}
	return out, rows, nil
}

func GetRoomResultByUserIdAndGameType(tx *gorm.DB, uid int32, gtype int32) ([]*mdr.Room, error) {
	var out []*mdr.Room
	// sql, param, err := squirrel.
	// 	Select(" r.room_id,status,game_user_result "). //
	// 	From(enum.RoomTableName+" r ").
	// 	LeftJoin(enum.PlayerRoomTableName+" pr on r.room_id = pr.room_id").
	// 	Where("pr.user_id = ? ", uid).
	// 	Limit(20).ToSql()

	// if err != nil {
	// 	return nil, errors.Internal("get player room list failed", err)
	// }

	// err = tx.Raw(sql, param...).Scan(&out).Error
	// if err != nil {
	// 	return nil, errors.Internal("get player list failed", err)
	// }

	sqlstr := " game_type =? and room_id in (select room_id from player_rooms where user_id = ?)"
	if err := tx.Where(sqlstr, gtype, uid).
		Order("created_at desc").Find(&out).Limit(20).Error; err != nil {
		return nil, errr.ErrRoomNotExisted
	}
	return out, nil
}

func UpdateRoomPlayTimes(tx *gorm.DB, rid int32, gtype int32) error {
	if err := tx.Model(&mdr.PlayerRoom{}).Where("room_id = ? and game_type = ? ",
		rid, gtype).UpdateColumn("play_times", gorm.Expr("play_times + 1")).
		Error; err != nil {
		return errors.Internal("update player room play times failed", err)
	}
	return nil
}

func GetGiveUpRoomIDByGameType(tx *gorm.DB,
	gtype int32) ([]int32, error) {
	var (
		out []int32
	)
	sql, param, err := squirrel.
	Select(" room_id "). //
		From(enum.RoomTableName + " r ").
		Where(" status = ? and game_type = ? and created_at >curdate() ",
		enumr.RoomStatusGiveUp, gtype).ToSql()

	if err != nil {
		return nil, errors.Internal("get room list failed", err)
	}

	err = tx.Raw(sql, param...).Scan(&out).Error
	if err != nil {
		return nil, errors.Internal("get list failed", err)
	}
	return out, nil
}

func GetDeadRoomPassword(tx *gorm.DB) ([]string, error) {
	var (
		out []string
	)
	sql, param, err := squirrel.
	Select(" password "). //
		From(enum.RoomTableName + " r ").
		Where("status < ? and updated_at <  date_sub(curdate(),interval 1 day)",
		enumr.RoomStatusDone).ToSql()

	if err != nil {
		return nil, errors.Internal("get dead room list failed", err)
	}

	err = tx.Raw(sql, param...).Scan(&out).Error
	if err != nil {
		return nil, errors.Internal("get list failed", err)
	}
	return out, nil
}

func CleanDeadRoomByUpdateAt(tx *gorm.DB) error {
	if err := tx.Model(&mdr.Room{}).
		Where("status < ? and updated_at <  date_sub(curdate(),interval 1 day)",
		enumr.RoomStatusDone).
		UpdateColumn("status", gorm.Expr("status*? + ?", 10, enumr.RoomStatusOverTimeClean)).
		Error; err != nil {
		return errors.Internal("update player room play times failed", err)
	}
	return nil
}

func GetRoomsGiveup(tx *gorm.DB) ([]*mdr.Room, error) {
	var (
		out []*mdr.Room
	)
	if err := tx.Where(" giveup = ? and status < ?", enumr.WaitGiveUp,
		enumr.RoomStatusDone).Order("created_at").Find(&out).Error; err != nil {
		return nil, errr.ErrRoomNotExisted
	}
	return out, nil
}

func PageRoom(tx *gorm.DB, page *mdpage.PageOption,
	mdroom *mdr.Room) ([]*mdr.Room, int64, error) {
	var out []*mdr.Room
	rows, rtx := page.Find(tx.Model(mdroom).Order("created_at desc").
		Where(mdroom), &out)
	if rtx.Error != nil {
		return nil, 0, errors.Internal("page room failed", rtx.Error)
	}
	return out, rows, nil
}

func PageRoomList(tx *gorm.DB, clubid int32, page int32, pagesize int32, flag int32) ([]*mdr.Room, error) {
	var out []*mdr.Room
	//if err := tx.Select(" room_id,status,game_type,max_number,round_now,round_number,game_param,game_user_result").
	//
	//	Where(" club_id = ? and flag = 2 and status >3 ", clubid).
	//	Offset((page - 1) * pagesize).Limit(pagesize).Order("created_at").
	//	Find(&test).Error; err != nil {
	//	return nil, errr.ErrRoomNotExisted
	//}

	strWhere := fmt.Sprintf("club_id = ? and status > %d ", enum.RoomStatusStarted)
	if flag > 0 {
		strWhere += fmt.Sprintf(" and flag = %d ", flag)
	}
	sql, param, err := squirrel.
	Select(" room_id,password,status,game_type,max_number,round_now,round_number,game_param,game_user_result,created_at").
		From(enum.RoomTableName + " r ").
		Where(strWhere, clubid).ToSql()

	if err != nil {
		return nil, errors.Internal("get room list failed", err)
	}

	err = tx.Raw(sql, param...).Scan(&out).Error
	if err != nil {
		return nil, errors.Internal("get list failed", err)
	}
	return out, nil
}
