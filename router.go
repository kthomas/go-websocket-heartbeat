package main

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"os"
)

var jwtSharedSecret = []byte(os.Getenv("JWT_SHARED_SECRET"))

var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return jwtSharedSecret, nil
	},
	ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
		w.Header().Set("content-type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusUnauthorized)
	},
	// When set, the middleware verifies that tokens are signed with the specific signing algorithm
	// If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
	// Important to avoid security issues described here: https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
	SigningMethod: jwt.SigningMethodHS256,
})

func Router() *httprouter.Router {
	router := httprouter.New()
	for _, route := range routes {
		if route.HttpHandlerFunc != nil {
			handler := BaseHandler(route)
			router.Handle(route.Method, route.Pattern,
				func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
					context.Set(r, "params", params)
					handler.ServeHTTP(w, r)
				},
			)
		} else if route.RouterHandleFunc != nil {
			router.Handle(route.Method, route.Pattern, route.RouterHandleFunc)
		}

	}
	return router
}
