package log

import (
	"log"
	"os"
	"playcards/utils/env"
)

const (
	EMERG = iota
	ALERT
	CRIT
	ERR
	WARN
	NOTICE
	INFO
	DEBUG
)

var LEVELS = map[string]uint{
	"emerg":  EMERG,
	"alert":  ALERT,
	"crit":   CRIT,
	"err":    ERR,
	"warn":   WARN,
	"notice": NOTICE,
	"info":   INFO,
	"debug":  DEBUG,
}

type Logger struct {
	file  *os.File
	log   *log.Logger
	level uint
	pid   []interface{}
}

func NewLogger(path string, level string) (*Logger, error) {

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, err
	}

	tlog := new(Logger)

	l := log.New(f, "", log.LstdFlags|log.LstdFlags)

	tlog.log = l
	tlog.level = LEVELS[level]
	tlog.file = f
	tlog.pid = []interface{}{env.Pid}

	return tlog, nil
}

func (tlog *Logger) Close() {
	tlog.file.Close()
}

func (tlog *Logger) Emerg(format string, v ...interface{}) {

	if tlog.level < EMERG {
		return
	}

	tlog.log.Printf("[EMERG] #%d "+format, append(tlog.pid, v...)...)
}

func (tlog *Logger) Alert(format string, v ...interface{}) {

	if tlog.level < ALERT {
		return
	}

	tlog.log.Printf("[ALERT] #%d "+format, append(tlog.pid, v...)...)
}

func (tlog *Logger) Crit(format string, v ...interface{}) {

	if tlog.level < CRIT {
		return
	}

	tlog.log.Printf("[CRIT] #%d "+format, append(tlog.pid, v...)...)
}

func (tlog *Logger) Err(format string, v ...interface{}) {

	if tlog.level < ERR {
		return
	}

	tlog.log.Printf("[ERROR] #%d "+format, append(tlog.pid, v...)...)
}

func (tlog *Logger) Warn(format string, v ...interface{}) {

	if tlog.level < WARN {
		return
	}

	tlog.log.Printf("[WARN] #%d "+format, append(tlog.pid, v...)...)
}

func (tlog *Logger) Notice(format string, v ...interface{}) {

	if tlog.level < NOTICE {
		return
	}

	tlog.log.Printf("[NOTICE] #%d "+format, append(tlog.pid, v...)...)
}

func (tlog *Logger) Info(format string, v ...interface{}) {

	if tlog.level < INFO {
		return
	}

	tlog.log.Printf("[INFO] #%d "+format, append(tlog.pid, v...)...)
}

func (tlog *Logger) Debug(format string, v ...interface{}) {

	if tlog.level < DEBUG {
		return
	}

	tlog.log.Printf("[DEBUG] #%d "+format, append(tlog.pid, v...)...)
}

var _logger *Logger

func SetDefault(l *Logger) {
	_logger = l
}

func Stdout() {
	l := log.New(os.Stdout, "", log.LstdFlags)
	tlog := new(Logger)
	tlog.log = l
	tlog.level = DEBUG
	tlog.file = os.Stdout
	tlog.pid = []interface{}{env.Pid}
	SetDefault(tlog)
}

func Emerg(format string, v ...interface{}) {
	_logger.Emerg(format, v...)
}

func Alert(format string, v ...interface{}) {
	_logger.Alert(format, v...)
}

func Crit(format string, v ...interface{}) {
	_logger.Crit(format, v...)
}

func Err(format string, v ...interface{}) {
	_logger.Err(format, v...)
}

func Warn(format string, v ...interface{}) {
	_logger.Warn(format, v...)
}

func Notice(format string, v ...interface{}) {
	_logger.Notice(format, v...)
}

func Info(format string, v ...interface{}) {
	_logger.Info(format, v...)
}

func Debug(format string, v ...interface{}) {
	_logger.Debug(format, v...)
}
