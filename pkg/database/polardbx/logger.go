package polardbx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	gormUtils "gorm.io/gorm/utils"

	"github.com/ninja0404/meme-signal/pkg/logger"
)

var _ gormLogger.Interface = new(MysqlLogger)

var (
	traceStr     = "%s\n[%.3fms] [rows:%v] %s"
	traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
	traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
)

var _ gormLogger.Interface = &MysqlLogger{}

func NewMysqlLogger(logger *logger.Logger, loggerLevel gormLogger.LogLevel) *MysqlLogger {
	l := MysqlLogger{
		logger:      logger,
		loggerLevel: loggerLevel,
	}
	if loggerLevel < gormLogger.Info {
		l.loggerConfig = gormLogger.Config{
			SlowThreshold:             1000 * time.Millisecond,
			LogLevel:                  gormLogger.Error,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  false,
		}
	} else {
		l.loggerConfig = gormLogger.Config{
			SlowThreshold:             1000 * time.Millisecond,
			LogLevel:                  gormLogger.Info,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  true,
		}
	}

	return &l
}

type MysqlLogger struct {
	logger       *logger.Logger
	loggerLevel  gormLogger.LogLevel
	loggerConfig gormLogger.Config
}

func (l *MysqlLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newLogger := *l
	newLogger.loggerLevel = level
	return &newLogger
}

func (l *MysqlLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.loggerLevel >= gormLogger.Info {
		formattedMsg := fmt.Sprintf(msg, data...)
		l.logger.Info(formattedMsg)
	}
}

func (l *MysqlLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.loggerLevel >= gormLogger.Warn {
		formattedMsg := fmt.Sprintf(msg, data...)
		l.logger.Warn(formattedMsg)
	}
}

func (l *MysqlLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.loggerLevel >= gormLogger.Error {
		formattedMsg := fmt.Sprintf(msg, data...)
		l.logger.Error(formattedMsg)
	}
}

func (l *MysqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.loggerLevel <= gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.loggerLevel >= gormLogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.loggerConfig.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.logger.Error(fmt.Sprintf(traceErrStr, gormUtils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql))
		} else {
			l.logger.Error(fmt.Sprintf(traceErrStr, gormUtils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql))
		}
	case elapsed > l.loggerConfig.SlowThreshold && l.loggerConfig.SlowThreshold != 0 && l.loggerLevel >= gormLogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.loggerConfig.SlowThreshold)
		if rows == -1 {
			l.logger.Warn(fmt.Sprintf(traceWarnStr, gormUtils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql))
		} else {
			l.logger.Warn(fmt.Sprintf(traceWarnStr, gormUtils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql))
		}
	case l.loggerLevel == gormLogger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.logger.Info(fmt.Sprintf(traceStr, gormUtils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql))
		} else {
			l.logger.Info(fmt.Sprintf(traceStr, gormUtils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql))
		}
	}
}

func mappingLoggerLevel(level string, openDebug bool) gormLogger.LogLevel {
	if openDebug {
		return gormLogger.Info
	} else {
		switch string(level) {
		case "debug", "DEBUG", "info", "INFO", "": // make the zero value useful
			return gormLogger.Warn
		case "warn", "WARN":
			return gormLogger.Warn
		case "error", "ERROR":
			return gormLogger.Error
		case "dpanic", "DPANIC":
			return gormLogger.Error
		case "panic", "PANIC":
			return gormLogger.Error
		case "fatal", "FATAL":
			return gormLogger.Error
		default:
			return gormLogger.Silent
		}
	}
}
