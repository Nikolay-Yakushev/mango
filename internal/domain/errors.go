package models

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

type StackTrace struct{
	trace     []byte
	cause     error
}


type MangoErrors struct {
	code   uint32
	msg    []string 
	strace *StackTrace
}
 
func (me *MangoErrors) ErrorExtra() string {
	message := make([]string, len(me.msg))
	copy(me.msg, message)
	message = append(message, fmt.Sprintf("stack-trace: %s", me.strace.trace))

	return fmt.Sprintf(strings.Join(message, ", "))
}

func (me *MangoErrors) Error() string {
	return me.msg[0]
}

func (me *MangoErrors) Unwrap() error {
	return me.strace.cause
}


func NewMangoError(msg string, code uint32, cause error) *MangoErrors {
	var buf []byte
	_, file , line, _ := runtime.Caller(1)
	buf = debug.Stack()

	fmtMsg := fmt.Sprintf("message: %s", msg)
	fmtFile := fmt.Sprintf("file: %s", file)
	fmtLine := fmt.Sprintf("line: %d", line)

	message := []string{fmtMsg, fmtFile, fmtLine}
	return &MangoErrors{
		code: code,
		msg : message,
		strace: &StackTrace{
			trace: buf,
			cause: cause,
		},
	} 
}

// TODO when should i use iota?
const (
	ForbiddenErrCode        uint32 = 1 << iota 
	TokenExpiredErrCode                        
	TokenInvalidErrCode
	NotFoundErrCode
	ConflictErrCode
)

var (
	ForbiddenErr =  NewMangoError("forbiden", ForbiddenErrCode, nil)
	TokenExpiredErr = NewMangoError("token expired", TokenExpiredErrCode, nil)
	TokenInvalidErr = NewMangoError("token invalid", TokenInvalidErrCode, nil)
	NotFoundErr = NewMangoError("not found", NotFoundErrCode, nil)
	ConflictErr =  NewMangoError("already exist", ConflictErrCode, nil)	
)