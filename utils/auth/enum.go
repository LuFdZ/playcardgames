package auth

import "time"

const RightsNone = 0
const SessionExpireTime = 1800 * time.Second

const (
	RightsPlayer       = 1 << iota
	RightsUserView
	RightsUserEdit
	RightsBillView
	RightsBillEdit
	RightsChatView
	RightsChatEdit
	RightsRoomView
	RightsRoomEdit
	RightsThirteenView
	RightsThirteenEdit
	RightsConfigView
	RightsConfigEdit
	RightsNoticeView
	RightsNoticeEdit
	RightsActivityView
	RightsActivityEdit
	RightsClubView
	RightsClubEdit
	RightsCommonView
	RightsCommonEdit
)

const (
	RightsUserAdmin     = RightsUserView | RightsUserEdit
	RightsBillAdmin     = RightsBillView | RightsBillEdit
	RightsGameAdmin     = RightsChatView | RightsChatEdit
	RightsRoomAdmin     = RightsRoomView | RightsRoomEdit
	RightsThirteenAdmin = RightsThirteenView | RightsThirteenEdit
	RightsConfigAdmin   = RightsConfigView | RightsConfigEdit
	RightsNoticeAdmin   = RightsNoticeView | RightsNoticeEdit
	RightsActivityAdmin = RightsActivityView | RightsActivityEdit
	RightsClubAdmin     = RightsClubView | RightsClubEdit
	RightsCommonAdmin   = RightsCommonView | RightsCommonEdit
	RightsViewer        = RightsUserView | RightsChatView | RightsBillView |
		RightsActivityView | RightsRoomView | RightsThirteenView |
		RightsNoticeView | RightsConfigView

	RightsAdmin = RightsUserAdmin |
		RightsGameAdmin |
		RightsBillAdmin |
		RightsRoomAdmin |
		RightsThirteenAdmin |
		RightsConfigAdmin |
		RightsNoticeAdmin |
		RightsActivityAdmin |
		RightsClubAdmin|
		RightsCommonAdmin|
		RightsPlayer
)
