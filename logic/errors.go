package logic

import (
	"fmt"
	"runtime"
)

func callerFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func wrapError(msg string, err error) error {
	return fmt.Errorf("%s, err : %w", msg, err)
}
