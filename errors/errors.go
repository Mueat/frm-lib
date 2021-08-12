package errors

const (
	// success
	OK = 0
	// SystemError
	System = 1
	// ParamsError
	Params = 2
	// ModelNotFound
	ModelNotFound = 3
)

var Errors = map[int]string{
	OK:            "success",
	System:        "SystemError",
	Params:        "ParamsError",
	ModelNotFound: "ModelNotFound",
}

func GetErrorMsg(code int) string {
	if msg, ok := Errors[code]; ok {
		return msg
	}
	return "UnknowError"
}

func AddErrors(errMap map[int]string) {
	for k, v := range errMap {
		Errors[k] = v
	}
}
