package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	router := Router()

	certificatePath := os.Getenv("SSL_CERTIFICATE_PATH")
	privateKeyPath := os.Getenv("SSL_PRIVATE_KEY_PATH")

	var listenPort string
	listenPort = os.Getenv("WEBSOCKET_PORT")
	if listenPort == "" {
		listenPort = "8080"
	}
	listenAddr := fmt.Sprintf("0.0.0.0:%s", listenPort)

	if shouldServeTLS(certificatePath, privateKeyPath) {
		log.Fatal(http.ListenAndServeTLS(listenAddr, certificatePath, privateKeyPath, router))
	} else {
		log.Fatal(http.ListenAndServe(listenAddr, router))
	}
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
