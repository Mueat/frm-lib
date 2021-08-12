package errors

import (
	"encoding/json"
	"runtime"
)

type Err struct {
	File string `json:"file"`
	Func string `json:"func"`
	Line int    `json:"line"`
	Code int    `json:"code"`
	Msg  string `json:"error"`
}

func (e Err) Error() string {
	eb, err := json.Marshal(e)
	if err != nil {
		return err.Error()
	}
	return string(eb)
}

func New(err error) Err {
	pc, file, line, ok := runtime.Caller(1)
	e := Err{}
	if ok {
		f := runtime.FuncForPC(pc)
		e.File = file
		e.Func = f.Name()
		e.Line = line
	}
	e.Msg = err.Error()
	return e
}

func Code(code int) Err {
	pc, file, line, ok := runtime.Caller(1)
	e := Err{}
	if ok {
		f := runtime.FuncForPC(pc)
		e.File = file
		e.Func = f.Name()
		e.Line = line
	}
	e.Code = code
	e.Msg = GetErrorMsg(code)
	return e
}

func Msg(msg string) Err {
	pc, file, line, ok := runtime.Caller(1)
	e := Err{}
	if ok {
		f := runtime.FuncForPC(pc)
		e.File = file
		e.Func = f.Name()
		e.Line = line
	}
	e.Msg = msg
	return e
}
