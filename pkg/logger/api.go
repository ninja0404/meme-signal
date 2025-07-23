package logger

import "go.uber.org/zap"

var defaultLogger *Logger
var defaultLoggerL1 *Logger

func Default() *Logger {
	return defaultLogger
}

func DefaultL1() *Logger {
	return defaultLoggerL1
}

func SetDefault(logger *Logger) {
	defaultLogger = logger
}

func SetDefaultL1(logger *Logger) {
	defaultLoggerL1 = logger
}

func Debug(msg string, fields ...Field) {
	defaultLogger.Debug(msg, fields...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Info(msg string, fields ...Field) {
	defaultLogger.Info(msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Warn(msg string, fields ...Field) {
	defaultLogger.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Error(msg string, fields ...Field) {
	defaultLogger.Error(msg, fields...)
}

// DPanic logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func DPanic(msg string, fields ...Field) {
	defaultLogger.DPanic(msg, fields...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func Panic(msg string, fields ...Field) {
	defaultLogger.Panic(msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func Fatal(msg string, fields ...Field) {
	defaultLogger.Fatal(msg, fields...)
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func With(fields ...Field) *Logger {
	return defaultLogger.With(fields...)
}

// WithOptions clones the current Logger, applies the supplied Options, and
// returns the resulting Logger. It's safe to use concurrently.
func WithOptions(opts ...Option) *Logger {
	return defaultLogger.WithOptions(opts...)
}

// Named adds a new path segment to the logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func Named(s string) *Logger {
	return defaultLogger.Named(s).WithOptions(zap.AddCallerSkip(-1))
}

func Level() string {
	return defaultLogger.Level().String()
}

func Close() {
	defaultLogger.Sync()
}
