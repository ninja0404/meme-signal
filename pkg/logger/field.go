package logger

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

func FieldMod(value string) Field {
	value = strings.Replace(value, " ", ".", -1)
	return String("mod", value)
}

// FieldErr ...
func FieldErr(err error) Field {
	return zap.Error(err)
}

func FieldErrKind(value string) Field {
	return String("err_kind", value)
}

// FieldKey ...
func FieldKey(value string) Field {
	return String("key", value)
}

func FieldMethod(value string) Field {
	return String("method", value)
}

// FieldEvent ...
func FieldEvent(value string) Field {
	return String("event", value)
}

func FieldCode(value int32) Field {
	return Int32("code", value)
}

func FieldTraceId(tid string) Field {
	return String("trace_id", tid)
}

func FieldCost(value time.Duration) Field {
	return String("cost", fmt.Sprintf("%.3f", float64(value.Round(time.Microsecond))/float64(time.Millisecond)))
}

// FieldStack ...
func FieldStack(value []byte) Field {
	return ByteString("stack", value)
}
