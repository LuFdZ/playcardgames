package errors

import "playcards/utils/errors"

const (
	ErrPlayerRemoved      = 22101
)

var (
	ErrRobotNotFind = errors.NotFound(22001, "没有可用的机器人了！")
)
