package utils

import (
	"reflect"
	"runtime"
)

func FunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func GetStack() []byte {
	buf := make([]byte, 10240)
	stackSize := runtime.Stack(buf, false)
	return buf[:stackSize]
}
