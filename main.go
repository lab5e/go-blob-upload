// Package main demonstrates how to upload a blob from an internet-connected device.
package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
)

// Certificate file is a PEM-encoded file that contains the certificate followed by intermediates and the CA
const certFile = "clientcert.crt"
const intermediates = "span-cert-chain.crt"

// The private key file is a PEM-encoded private key
const privateKeyFile = "private.key"

// This server is a regular HTTP server that
// 1) only accept POST requests and
// 2) authenticates via client certificates
const blobEndpoint = "https://data.lab5e.com/"

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("Usage %s [file to upload]\n", os.Args[0])
		return
	}
	fileToUpload := os.Args[1]

	// Read the file we'll upload. File sizes are currently limited to 16MB
	buf, err := os.ReadFile(fileToUpload)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Guess the content-type header from the contents of the file.
	contentType := http.DetectContentType(buf)
	fmt.Println("Detected content-type: ", contentType)

	// Set up the certificates. We'll do a slight shortcut here and just lump the intermediates and the root
	// certificate in the same pool.
	certBuf, err := os.ReadFile(intermediates)
	if err != nil {
		fmt.Println("Error reading intermediates: ", err)
		return
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(certBuf) {
		fmt.Println("Could not append certificates to pool")
		return
	}

	// Load the client certificate and private key
	cert, err := tls.LoadX509KeyPair(certFile, privateKeyFile)
	if err != nil {
		fmt.Println("Error loading certificate: ", err)
		return
	}

	// The client is nothing out of the extraordinary, just the TLS config for the client.
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				RootCAs:            caPool,
				ClientCAs:          caPool,
				InsecureSkipVerify: false,
			},
		},
	}

	// Finally - post the file to the service
	resp, err := client.Post(blobEndpoint+"some/random/path", contentType, bytes.NewReader(buf))
	if err != nil {
		fmt.Println("Error POSTing file:", err)
		return
	}
	fmt.Println("Status from server is ", resp.Status)

	// At this point the file should show up in the Span frontend at https://span.lab5e.com/
}
