package logger

import (
	"io"

	"github.com/ninja0404/meme-signal/pkg/logger/rotate"
)

func newRotate(config *Config) io.Writer {
	rotateLog := rotate.NewLogger()
	rotateLog.Filename = config.Filename()
	rotateLog.MaxSize = config.MaxSize // MB
	rotateLog.MaxAge = config.MaxAge   // days
	rotateLog.MaxBackups = config.MaxBackup
	rotateLog.Interval = config.Interval
	rotateLog.LocalTime = true
	rotateLog.Compress = false
	return rotateLog
}
