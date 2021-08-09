package http

type RouterFun func(app *App)

type Router struct {
	Method  string
	URL     string
	Handler RouterFun
}
