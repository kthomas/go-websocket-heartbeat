package main

import (
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"time"
)

func BaseHandler(route Route) http.Handler {
	var handler http.Handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		route.HttpHandlerFunc.ServeHTTP(w, r)

		log.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			route.Name,
			time.Since(start),
		)
	})

	if route.Authorize {
		handler = jwtMiddleware.Handler(handler)
	}

	return handler
}

func WebsocketHandler(ws *websocket.Conn) {
	go ping(ws)

	for {
		err := readFrame(ws)
		if err != nil {
			log.Printf("Stopping listener for websocket: %s", ws.Request().RemoteAddr)
			break
		}
	}
}

func readFrame(ws *websocket.Conn) (error) {
	frameReader, err := ws.NewFrameReader()
	if err != nil {
		log.Printf("Failed to receive on websocket: %s", ws.Request().RemoteAddr)
		return err
	}

	var payload []byte
	len, err := frameReader.Read(payload)
	if err != nil {
		log.Printf("Failed to receive on websocket: %s", ws.Request().RemoteAddr)
		return err
	}
	payloadType := frameReader.PayloadType()
	//log.Printf("Received %v-byte payload (type %v) on websocket: %s", len, payloadType, ws.Request().RemoteAddr)

	if payloadType == websocket.PingFrame {
		log.Printf("Received %v-byte ping frame on websocket: %s", len, ws.Request().RemoteAddr)
		pong(ws)
	} else if payloadType == websocket.PongFrame {
		log.Printf("Received %v-byte pong frame on websocket: %s", len, ws.Request().RemoteAddr)
	} else if payloadType == websocket.TextFrame {
		log.Printf("Received %v-byte text frame on websocket %s; payload: %s", len, ws.Request().RemoteAddr, string(payload))
	}

	return nil
}

func ping(ws *websocket.Conn) {
	for {
		log.Printf("Attempting to send ping frame on websocket: %s", ws.Request().RemoteAddr)
		frameWriter, err := ws.NewFrameWriter(websocket.PingFrame)
		if err != nil {
			log.Printf("Failed to send ping frame on websocket %s; error: %s", ws.Request().RemoteAddr, err.Error())
			break
		}

		len, err := frameWriter.Write([]byte("ping"))
		if err != nil {
			log.Printf("Failed to send ping frame on websocket %s; error: %s", ws.Request().RemoteAddr, err.Error())
			break
		}
		log.Printf("Sent %v-byte ping frame on websocket: %s", len, ws.Request().RemoteAddr)
		time.Sleep(time.Second * 30)
	}
}

func pong(ws *websocket.Conn) {
	log.Printf("Attempting to send pong frame on websocket: %s", ws.Request().RemoteAddr)
	frameWriter, err := ws.NewFrameWriter(websocket.PongFrame)
	if err != nil {
		log.Printf("Failed to send pong frame on websocket %s; error: %s", ws.Request().RemoteAddr, err.Error())
	} else {
		len, err := frameWriter.Write([]byte("pong"))
		if err != nil {
			log.Printf("Failed to send pong frame on websocket %s; error: %s", ws.Request().RemoteAddr, err.Error())
		}
		log.Printf("Sent %v-byte pong frame on websocket: %s", len, ws.Request().RemoteAddr)
	}
}
