package middleware

import "net/http"

func Chain(h http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) http.Handler {
	result := http.Handler(h)
	for i := len(middlewares) - 1; i >= 0; i-- {
		result = middlewares[i](result)
	}
	return result
}
