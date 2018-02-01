package enum

import "time"

const (
	LoopTime             = 1
	MaxMailRecordCount   = 10              //最大页数
	PlayerMailOverTime   = 1800* time.Second //1800//用户邮件过期时间(小时)
	MailEndLogOverTime   = 5* time.Second //7200//邮件有效期 过期时间(小时)
	PlayerMailMaxNumber  = 3               //30 用户邮件最大保存天数
	MailSendLogMaxNumber = 3               //60 邮件发送记录最大保存天数
)

const (
	MailTypeSysMode    = 1 //邮件类型 系统模板邮件
	MailTypeSysCustom  = 2 //邮件类型 系统自定义邮件
	MailTypeUserCustom = 3 //邮件类型 用户邮件
)

// mail info status
const (
	MailInfoStatusNom = 110 //邮件使用状态 正常
	MailInfoStatusBan = 120 //邮件使用状态 禁用
	MailInfoStatusDel = 130 //邮件使用状态 删除
)

// mail send log status
const (
	MailSendNom     = 110 //邮件使用状态 正常
	MailSendOverdue = 120 //邮件使用状态 过期
	MailSendDel     = 130 //邮件使用状态 取消
)

// player mail status
const (
	PlayermailNon          = 110 //邮件使用状态 正常
	PlayermailRead         = 120 //邮件使用状态 已读
	PlayermailReadClose    = 130 //邮件使用状态 已读关闭
	PlayermailReceiveClose = 140 //邮件使用状态 已领取关闭
	PlayermailOverdue      = 150 //邮件使用状态 过期
)

const (
	MailRecharge   = 1001
	MailClubJoin   = 1101
	MailClubUnJoin = 1102
	MailGameGiveUp = 1201
	MailGameOver   = 1202
	MailInvite     = 1301
	MailShare      = 1302
)

const (
	Success = 1
	Fail    = 2
)

//邮件附件物品分类
const (
	//一级
	CurrencyType = "100"
	//二级
	CurrencySubTypeGold    = "1" //金币
	CurrencySubTypeDiamond = "2" //钻石

	//一级
	ItemType = "200"
)
