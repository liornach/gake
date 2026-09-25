package logic

import (
	"fmt"
	"runtime"
)

func callerFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

// wrapError adds context to err; a nil err stays nil.
func wrapError(msg string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s, err : %w", msg, err)
}
