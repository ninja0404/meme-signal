package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	Field  = zap.Field
	Logger = zap.Logger
	Option = zap.Option
)

var (
	// String ...
	String = zap.String
	// Any ...
	Any = zap.Any
	// Int64 ...
	Int64 = zap.Int64
	// Int ...
	Int = zap.Int
	// Int32 ...
	Int32 = zap.Int32
	//Uint64
	Uint64 = zap.Uint64
	Uint8  = zap.Uint8
	// Uint ...
	Uint = zap.Uint
	//Uint32
	Uint32 = zap.Uint32
	//Bool
	Bool = zap.Bool
	//Time
	Time = zap.Time
	// Duration ...
	Duration = zap.Duration
	// Durationp ...
	Durationp = zap.Durationp
	// Object ...
	Object = zap.Object
	// Namespace ...
	Namespace = zap.Namespace
	// Reflect ...
	Reflect = zap.Reflect
	// Skip ...
	Skip = zap.Skip()
	// ByteString ...
	ByteString = zap.ByteString

	Float64 = zap.Float64
)

func newLogger(c *Config) *zap.Logger {
	// 强制启用颜色输出（用于美观的控制台日志）
	//if c.Debug {
	color.NoColor = false
	//}

	zapOptions := make([]zap.Option, 0)
	zapOptions = append(zapOptions, zap.AddStacktrace(zap.DPanicLevel))
	if c.AddCaller {
		zapOptions = append(zapOptions, zap.AddCaller(), zap.AddCallerSkip(c.CallerSkip))
	}

	var ws zapcore.WriteSyncer = os.Stdout
	if c.Discard {
		ws, _ = os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	} else if c.OUTPUT == "file" {
		ws = zapcore.AddSync(newRotate(c))
	}

	if c.Async {
		ws = &zapcore.BufferedWriteSyncer{
			WS:            zapcore.AddSync(ws),
			FlushInterval: c.FlushInterval,
			Size:          c.FlushBufferSize,
		}
	}
	// if config.Debug {
	// 	ws = os.Stdout
	// } else {
	// 	ws = zapcore.AddSync(newRotate(config))
	// }

	// if c.Async {
	// 	ws = &zapcore.BufferedWriteSyncer{
	// 		WS: zapcore.AddSync(ws), FlushInterval: defaultFlushInterval, Size: defaultBufferSize}
	// }

	lv := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if err := lv.UnmarshalText([]byte(c.Level)); err != nil {
		panic(err)
	}

	if !c.DisableSentry {
		var sentryLevel zapcore.Level = zapcore.ErrorLevel
		err := sentryLevel.UnmarshalText([]byte(c.SentryLevel))
		if err != nil {
			panic("sentry level not valid")
		}
		sentryCore := NewSentryCore(sentryLevel)
		zapOptions = append(zapOptions, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, sentryCore)
		}))
	}

	encoderConfig := defaultZapConfig()
	if c.Debug {
		encoderConfig.EncodeLevel = debugEncodeLevel
	}
	core := zapcore.NewCore(
		func() zapcore.Encoder {
			if c.Debug {
				return zapcore.NewConsoleEncoder(*encoderConfig)
			}
			return zapcore.NewJSONEncoder(*encoderConfig)
		}(),
		ws,
		lv,
	)

	zapLogger := zap.New(
		core,
		zapOptions...,
	)

	return zapLogger.Named(c.Name)
}

func defaultZapConfig() *zapcore.EncoderConfig {
	return &zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func debugEncodeLevel(lv zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var colorize = color.RedString
	switch lv {
	case zapcore.DebugLevel:
		colorize = color.BlueString
	case zapcore.InfoLevel:
		colorize = color.GreenString
	case zapcore.WarnLevel:
		colorize = color.YellowString
	case zapcore.ErrorLevel, zapcore.PanicLevel, zapcore.DPanicLevel, zapcore.FatalLevel:
		colorize = color.RedString
	default:
	}
	enc.AppendString(colorize(fmt.Sprintf("[%s]", lv.CapitalString())))
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Local().Format("2006-01-02 15:04:05.000"))
}
