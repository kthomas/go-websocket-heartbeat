package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kthomas/go-aws-config"
	"github.com/kthomas/go-sqs-consumer"
)

func main() {
	router := Router()

	certificatePath := os.Getenv("SSL_CERTIFICATE_PATH")
	privateKeyPath := os.Getenv("SSL_PRIVATE_KEY_PATH")
	listenAddr := buildListenAddr()

	if shouldConsumeSqsQueue() {
		consumeSqsQueue()
	}

	if shouldServeTLS(certificatePath, privateKeyPath) {
		log.Fatal(http.ListenAndServeTLS(listenAddr, certificatePath, privateKeyPath, router))
	} else {
		log.Fatal(http.ListenAndServe(listenAddr, router))
	}
}

func buildListenAddr() (string) {
	var listenPort string
	listenPort = os.Getenv("WEBSOCKET_PORT")
	if listenPort == "" {
		listenPort = "8080"
	}
	return fmt.Sprintf("0.0.0.0:%s", listenPort)
}

func consumeSqsQueue() {
	config := awsconf.GetConfig()
	if config.DefaultSqsQueueUrl != nil {
		sqs.Consume(*config.DefaultSqsQueueUrl, SqsQueueHandler)
	}
}

func shouldConsumeSqsQueue() (bool) {
	return awsconf.GetConfig().DefaultSqsQueueUrl != nil
}

func shouldServeTLS(certificatePath string, privateKeyPath string) (bool) {
	var tls = false
	if _, err := os.Stat(certificatePath); err == nil {
		if _, err := os.Stat(privateKeyPath); err == nil {
			tls = true
		}
	}
	return tls
}
