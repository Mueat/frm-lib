package errors

import (
	"encoding/json"
	"runtime"
)

type Err struct {
	File string `json:"file"`
	Func string `json:"func"`
	Line int    `json:"line"`
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
