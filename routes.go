package main

import (
	"golang.org/x/net/websocket"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Route struct {
	Name			string
	Method			string
	Pattern			string
	HttpHandlerFunc		http.HandlerFunc
	RouterHandleFunc	httprouter.Handle
	Authorize		bool
}
type Routes []Route

var routes = Routes{

	Route{
		Name: "Websocket",
		Method: "GET",
		Pattern: "/",
		HttpHandlerFunc: websocket.Handler(WebsocketHandler).ServeHTTP,
		Authorize: true,
	},
}
