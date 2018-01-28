package enum

const (
	MaxMailRecordCount = 10 //最大页数
)

const (
	MailTypeSysMode = 1 //邮件类型 系统模板邮件
	MailTypeSysCustom = 2 //邮件类型 系统自定义邮件
	MailTypeUserCustom = 3 //邮件类型 用户邮件
)

// mail info status
const (
	MailInfoStatusNom = 1 //邮件使用状态 正常
	MailInfoStatusBan = 2 //邮件使用状态 禁用
	MailInfoStatusDel = 3 //邮件使用状态 删除
)

// mail send log status
const (
	MailSendNom = 1 //邮件使用状态 正常
	MailSendOverdue = 2 //邮件使用状态 过期
	MailSendDel = 3 //邮件使用状态 取消
)

// player mail status
const (
	PlayermailNon = 1 //邮件使用状态 正常
	PlayermailRead = 2 //邮件使用状态 已读
	PlayermailReceive = 3 //邮件使用状态 已领取
	PlayermailOverdue = 4 //邮件使用状态 过期
)

const (
	Success = 1
	Fail    = 2
)

//邮件附件物品一级分类
const (
	CurrencyType   = 100
	ItemType = 200
)

