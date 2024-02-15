package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"pemilu2024/kpu"
)
// loggingTransport is a custom transport that logs each HTTP request and response
type loggingTransport struct {
	Transport http.RoundTripper
}

// RoundTrip executes a single HTTP transaction and logs the request and response
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Log the request
	log.Printf("Request: %s %s\n", req.Method, req.URL.String())

	// Execute the request
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Log the response
	log.Printf("Response: %s\n", resp.Status)

	// Return the response
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

	handler := &Handler{
		sirekap: kpu.NewSirekap(client),
	}
	// Define the endpoint handler
	http.HandleFunc("/fetch-votes", handler.GetVotes)
	http.HandleFunc("/fetch-locations", handler.GetLocations)

	// Start the HTTP server
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
