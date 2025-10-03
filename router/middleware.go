package router

type Middleware func(Context, HandlerFunc)

func Chain(handler HandlerFunc, middlewares ...Middleware) HandlerFunc {
	if len(middlewares) == 0 {
		return handler
	}

	wrapped := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		mw := middlewares[i]
		next := wrapped
		wrapped = func(ctx Context) {
			mw(ctx, next)
		}
	}
	return wrapped
}
