package main

import (
	"bytes"
	"encoding/json"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/aws"

	ktsqs "github.com/kthomas/go-sqs-consumer"
)

var sockets []*websocket.Conn

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
	link(ws)

	go ping(ws)

	for {
		err := readFrame(ws)
		if err != nil {
			log.Printf("Stopping listener for websocket: %s", ws.Request().RemoteAddr)
			unlink(ws)
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

	var payload = make([]byte, frameReader.Len(), frameReader.Len())
	length, err := frameReader.Read(payload)
	if err != nil {
		log.Printf("Failed to receive on websocket: %s", ws.Request().RemoteAddr)
		return err
	}
	payloadType := frameReader.PayloadType()

	if payloadType == websocket.ContinuationFrame {
		log.Printf("Received %v-byte continuation frame on websocket %s; payload: %s", length, ws.Request().RemoteAddr, string(payload))
	} else if payloadType == websocket.PingFrame {
		log.Printf("Received %v-byte ping frame on websocket: %s", length, ws.Request().RemoteAddr)
		pong(ws)
	} else if payloadType == websocket.PongFrame {
		log.Printf("Received %v-byte pong frame on websocket: %s", length, ws.Request().RemoteAddr)
	} else if payloadType == websocket.TextFrame {
		log.Printf("Received %v-byte text frame on websocket %s; payload: %s", length, ws.Request().RemoteAddr, string(payload))

		queueUrl := os.Getenv("AWS_SQS_QUEUE_PUBLISH_URL")
		if queueUrl != "" {
			payload = bytes.Trim(payload, "\x00")
			push(queueUrl, string(payload))
		}
	}

	return nil
}

func ping(ws *websocket.Conn) {
	for {
		log.Printf("Attempting to send ping frame on websocket: %s", ws.Request().RemoteAddr)
		frameWriter, err := ws.NewFrameWriter(websocket.PingFrame)
		if err != nil {
			log.Printf("Failed to send ping frame on websocket %s; error: %s", ws.Request().RemoteAddr, err.Error())
			unlink(ws)
			break
		}

		length, err := frameWriter.Write([]byte("ping"))
		if err != nil {
			log.Printf("Failed to send ping frame on websocket %s; error: %s", ws.Request().RemoteAddr, err.Error())
			unlink(ws)
			break
		} else {
			log.Printf("Sent %v-byte ping frame on websocket: %s", length, ws.Request().RemoteAddr)
		}
		time.Sleep(time.Second * 30)
	}
}

func pong(ws *websocket.Conn) {
	log.Printf("Attempting to send pong frame on websocket: %s", ws.Request().RemoteAddr)
	frameWriter, err := ws.NewFrameWriter(websocket.PongFrame)
	if err != nil {
		log.Printf("Failed to send pong frame on websocket %s; error: %s", ws.Request().RemoteAddr, err.Error())
		unlink(ws)
	} else {
		length, err := frameWriter.Write([]byte("pong"))
		if err != nil {
			log.Printf("Failed to send pong frame on websocket %s; error: %s", ws.Request().RemoteAddr, err.Error())
			unlink(ws)
		} else {
			log.Printf("Sent %v-byte pong frame on websocket: %s", length, ws.Request().RemoteAddr)
		}
	}
}

func send(message []byte, ws *websocket.Conn) {
	log.Printf("Attempting to send text frame on websocket: %s", ws.Request().RemoteAddr)
	frameWriter, err := ws.NewFrameWriter(websocket.TextFrame)
	if err != nil {
		log.Printf("Failed to send text frame on websocket %s; error: %s", ws.Request().RemoteAddr, err.Error())
		unlink(ws)
	} else {
		length, err := frameWriter.Write(message)
		if err != nil {
			log.Printf("Failed to send text frame on websocket %s; error: %s", ws.Request().RemoteAddr, err.Error())
			unlink(ws)
		} else {
			log.Printf("Sent %v-byte text frame on websocket: %s", length, ws.Request().RemoteAddr)
		}
	}
}

func push(queueUrl string, payload string) {
	go func(url string) {
		svc := sqs.New(session.New())
		params := &sqs.SendMessageInput{
			QueueUrl:	aws.String(url),
			MessageBody:	aws.String(payload),
		}

		response, err := svc.SendMessage(params)
		if err != nil {
			log.Println("Encountered error while sending to SQS queue; " + err.Error())
		} else {
			log.Printf("Sent message to SQS queue; message id: %s", &response.MessageId)
		}
	}(queueUrl)
}

func link(ws *websocket.Conn) {
	sockets = append(sockets, ws)
	log.Printf("Linked websocket connection: %s", ws.Request().RemoteAddr)
}

func unlink(staleConn *websocket.Conn) {
	var index = -1
	for i, ws := range sockets {
		if ws == staleConn {
			index = i
			break
		}
	}
	if index != -1 {
		sockets = append(sockets[:index], sockets[index + 1:]...)
		log.Printf("Unlinked websocket connection: %s", staleConn.Request().RemoteAddr)
	}
}

func SqsQueueHandler(message *ktsqs.Message) {
	go func() {
		for _, ws := range sockets {
			bytes, err := json.Marshal(&message)
			if err != nil {
				log.Printf("Failed to marshal message received from SQS queue to JSON; error: %s", err.Error())
			} else {
				send(bytes, ws)
			}
		}
	}()
}
