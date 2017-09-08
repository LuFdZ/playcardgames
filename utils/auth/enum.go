package auth

const RightsNone = 0
const (
	RightsPlayer = 1 << iota
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
		RightsPlayer
)
