package kafka

import (
	"fmt"
	"strings"

	"github.com/IBM/sarama"

	"github.com/ninja0404/meme-signal/pkg/logger"
)

var _ sarama.StdLogger = (*LoggerKafka)(nil)

const (
	LOGGER_DEBUG = iota + 1
	LOGGER_INFO
)

type LoggerKafka struct {
	l     *logger.Logger
	level int
}

func NewLoggerKafka(l *logger.Logger, level int) *LoggerKafka {
	return &LoggerKafka{l: l, level: level}
}

func (l *LoggerKafka) Print(v ...interface{}) {
	format := make([]string, 0, len(v))
	for i := 0; i < len(v); i++ {
		format = append(format, "%+v")
	}
	lmsg := fmt.Sprintf(strings.Join(format, " "), v...)
	if l.level == LOGGER_DEBUG {
		l.l.Debug(lmsg)
	} else {
		l.l.Info(lmsg)
	}
}

func (l *LoggerKafka) Printf(format string, v ...interface{}) {
	lmsg := fmt.Sprintf(format, v...)
	if l.level == LOGGER_DEBUG {
		l.l.Debug(lmsg)
	} else {
		l.l.Info(lmsg)
	}
}

func (l *LoggerKafka) Println(v ...interface{}) {
	format := make([]string, 0, len(v))
	for i := 0; i < len(v); i++ {
		format = append(format, "%+v")
	}
	lmsg := fmt.Sprintf(strings.Join(format, " ")+"\n", v...)
	if l.level == LOGGER_DEBUG {
		l.l.Debug(lmsg)
	} else {
		l.l.Info(lmsg)
	}
}
