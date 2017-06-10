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
	RightsRegionEdit
	RightsStoreEdit
	RightsConfigEdit
	RightsZodiacEdit
	RightsActivityView
	RightsActivityEdit
)

const (
	RightsUserAdmin     = RightsUserView | RightsUserEdit
	RightsBillAdmin     = RightsBillView | RightsBillEdit
	RightsGameAdmin     = RightsChatView | RightsChatEdit
	RightsRegionAdmin   = RightsRegionEdit
	RightsStoreAdmin    = RightsStoreEdit
	RightsConfigAdmin   = RightsConfigEdit
	RightsZodiacAdmin   = RightsZodiacEdit
	RightsActivityAdmin = RightsActivityView | RightsActivityEdit
	RightsViewer        = RightsUserView | RightsChatView | RightsBillView | RightsActivityView

	RightsAdmin = RightsUserAdmin |
		RightsGameAdmin |
		RightsBillAdmin |
		RightsRegionAdmin |
		RightsStoreAdmin |
		RightsConfigAdmin |
		RightsZodiacAdmin |
		RightsActivityAdmin |
		RightsPlayer
)
