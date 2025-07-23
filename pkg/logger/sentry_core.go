package logger

import (
	"fmt"
	"math"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type SentryCore struct {
	level        zapcore.Level
	fields       []zapcore.Field
	flushTimeout time.Duration
}

func NewSentryCore(level zapcore.Level) zapcore.Core {
	core := &SentryCore{
		level:        level,
		flushTimeout: 5 * time.Second,
		fields:       make([]zapcore.Field, 0),
	}

	return core
}

func (c *SentryCore) Enabled(l zapcore.Level) bool {
	return c.level.Enabled(l)
}

func (c *SentryCore) With(f []zapcore.Field) zapcore.Core {
	clone := c.clone()
	clone.fields = append(clone.fields, f...)
	return clone
}

func (c *SentryCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(e.Level) {
		return ce.AddCore(e, c)
	}
	return ce
}

func (c *SentryCore) Write(ent zapcore.Entry, fields []zap.Field) error {
	newFields := make([]zapcore.Field, 0, len(c.fields)+len(fields))
	newFields = append(newFields, c.fields...)
	newFields = append(newFields, fields...)

	meta := fieldToSentryMeta(newFields)

	sentry.WithScope(func(scope *sentry.Scope) {
		if len(newFields) > 0 {
			scope.SetExtras(meta)
		}
		scope.SetLevel(sentryLevel(ent.Level))
		//scope.SetTag(ALERT_TAG_KEY, defalutAlert)
		sentry.CaptureMessage(ent.Message)

		if ent.Level > zapcore.ErrorLevel {
			c.Sync()
		}
	})
	return nil
}

func (c *SentryCore) Sync() error {
	sentry.Flush(c.flushTimeout)
	return nil
}

func (c *SentryCore) clone() *SentryCore {
	newFields := make([]zapcore.Field, 0, len(c.fields))
	newFields = append(newFields, c.fields...)
	return &SentryCore{
		level:        c.level,
		fields:       newFields,
		flushTimeout: c.flushTimeout,
	}
}

func fieldToSentryMeta(fields []zapcore.Field) map[string]interface{} {
	retMap := make(map[string]interface{})

	for _, f := range fields {
		switch f.Type {
		case zapcore.ArrayMarshalerType:
			// 支持
			retMap[f.Key] = f.Interface.(zapcore.ArrayMarshaler)
		case zapcore.ObjectMarshalerType:
			// 支持
			retMap[f.Key] = f.Interface.(zapcore.ObjectMarshaler)
		case zapcore.BinaryType:
			retMap[f.Key] = f.Interface.([]byte)
		case zapcore.BoolType:
			retMap[f.Key] = f.Integer == 1
		case zapcore.ByteStringType:
			retMap[f.Key] = string(f.Interface.([]byte))
		case zapcore.Complex128Type:
			retMap[f.Key] = f.Interface.(complex128)
		case zapcore.Complex64Type:
			retMap[f.Key] = f.Interface.(complex64)
		case zapcore.DurationType:
			retMap[f.Key] = time.Duration(f.Integer)
		case zapcore.Float64Type:
			retMap[f.Key] = math.Float64frombits(uint64(f.Integer))
		case zapcore.Float32Type:
			retMap[f.Key] = math.Float32frombits(uint32(f.Integer))
		case zapcore.Int64Type:
			retMap[f.Key] = f.Integer
		case zapcore.Int32Type:
			retMap[f.Key] = int32(f.Integer)
		case zapcore.Int16Type:
			retMap[f.Key] = int16(f.Integer)
		case zapcore.Int8Type:
			retMap[f.Key] = int8(f.Integer)
		case zapcore.StringType:
			retMap[f.Key] = f.String
		case zapcore.TimeType:
			if f.Interface != nil {
				retMap[f.Key] = time.Unix(0, f.Integer).In(f.Interface.(*time.Location))
			} else {
				// Fall back to UTC if location is nil.
				retMap[f.Key] = time.Unix(0, f.Integer)
			}
		case zapcore.TimeFullType:
			retMap[f.Key] = f.Interface.(time.Time)
		case zapcore.Uint64Type:
			retMap[f.Key] = uint64(f.Integer)
		case zapcore.Uint32Type:
			retMap[f.Key] = uint32(f.Integer)
		case zapcore.Uint16Type:
			retMap[f.Key] = uint16(f.Integer)
		case zapcore.Uint8Type:
			retMap[f.Key] = uint8(f.Integer)
		case zapcore.UintptrType:
			retMap[f.Key] = uintptr(f.Integer)
		case zapcore.ReflectType:
			retMap[f.Key] = f.Interface
		case zapcore.NamespaceType:
			continue
		case zapcore.StringerType:
			//retMap[f.Key] = stringer.(fmt.Stringer).String() 有可能panic 暂不支持
			continue
		case zapcore.ErrorType:
			internalErr := f.Interface.(error)
			retMap[f.Key] = internalErr.Error()
		case zapcore.SkipType:
			continue
		default:
			retMap[f.Key] = fmt.Sprintf("unknown field type: %v", f)
		}
	}

	return retMap
}

func sentryLevel(lvl zapcore.Level) sentry.Level {
	switch lvl {
	case zapcore.DebugLevel:
		return sentry.LevelDebug
	case zapcore.InfoLevel:
		return sentry.LevelInfo
	case zapcore.WarnLevel:
		return sentry.LevelWarning
	case zapcore.ErrorLevel:
		return sentry.LevelError
	case zapcore.DPanicLevel:
		return sentry.LevelFatal
	case zapcore.PanicLevel:
		return sentry.LevelFatal
	case zapcore.FatalLevel:
		return sentry.LevelFatal
	default:
		// Unrecognized levels are fatal.
		return sentry.LevelFatal
	}
}
