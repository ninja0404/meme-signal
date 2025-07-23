package logger

import "context"

var activeLogKey struct{}

func ContextWithLog(ctx context.Context, l *Logger) context.Context {
	if l == nil {
		return ctx
	}
	return context.WithValue(ctx, activeLogKey, l)
}

func LogFromContext(ctx context.Context) *Logger {
	val := ctx.Value(activeLogKey)
	if l, ok := val.(*Logger); ok {
		return l
	}
	return defaultLogger
}
