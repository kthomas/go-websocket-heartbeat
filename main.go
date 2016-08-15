package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	router := Router()

	certificatePath := os.Getenv("SSL_CERTIFICATE_PATH")
	privateKeyPath := os.Getenv("SSL_PRIVATE_KEY_PATH")

	if shouldServeTLS(certificatePath, privateKeyPath) {
		log.Fatal(http.ListenAndServeTLS("0.0.0.0:8080", certificatePath, privateKeyPath, router))
	} else {
		log.Fatal(http.ListenAndServe("0.0.0.0:8080", router))
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
