package enum

const (
	ConfigTableName = "configs"
)

var Configs = []int32{
	ConfigIDChatFee,
	ConfigIDWorldChatFee,
	ConfigIDBanChatFee,
	ConfigIDPrisonFee,

	ConfigIDChatCD,
	ConfigIDWorldChatCD,
	ConfigIDBanChatCD,
	ConfigIDPersonCD,

	ConfigIDMaxBanChat,
	ConfigIDMaxPrison,
}

const (
	ConfigIDChatFee = iota + 100000
	ConfigIDWorldChatFee
	ConfigIDBanChatFee
	ConfigIDPrisonFee
)

const (
	ConfigIDChatCD = iota + 200000
	ConfigIDWorldChatCD
	ConfigIDBanChatCD
	ConfigIDPersonCD
)

const (
	ConfigIDMaxBanChat = iota + 300000
	ConfigIDMaxPrison
)
