package main

import (
	"crypto/tls"
	"log"
	"net/http"

	"github.com/pararang/pemilu2024/handler"
	"github.com/pararang/pemilu2024/kpu"
)

type loggingTransport struct {
	Transport http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	log.Printf("Request: %s %s\n", req.Method, req.URL.String())

	// Execute the request
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	log.Printf("Response: %s\n", resp.Status)
	return resp, nil
}

func main() {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Create a new HTTP client with the custom transport
	client := &http.Client{
		Transport: &loggingTransport{
			Transport: transport,
		},
	}

	handler := handler.NewHandler(kpu.NewSirekap(client))
	// Define the endpoint handler
	http.HandleFunc("/fetch-votes", handler.GetVotes)
	http.HandleFunc("/fetch-locations", handler.GetLocations)

	// Start the HTTP server
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
