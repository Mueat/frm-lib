package errors

const (
	// success
	OK = 0
	// 系统错误
	System = 1
	// 参数错误
	Params = 2
)

var Errors = map[int]string{
	OK:     "success",
	System: "系统错误",
	Params: "数据库错误",
}

func AddErrors(errMap map[int]string) {
	for k, v := range errMap {
		Errors[k] = v
	}
}
