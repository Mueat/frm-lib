package http

type RouterFun func(app *App)

type Router struct {
	Method  string
	URL     string
	Handler RouterFun
}

func MergeRouters(routers ...[]Router) []Router {
	rts := make([]Router, 0)
	for _, rs := range routers {
		rts = append(rts, rs...)
	}
	return rts
}
