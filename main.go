package main

import (
	"crypto/tls"
	"log"
	"net/http"

	"github.com/pararang/pemilu2024/kpu"
	"github.com/pararang/pemilu2024/presenter"
)

type loggingTransport struct {
	Transport http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	log.Printf("Request: %s %s\n", req.Method, req.URL.String())

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

	client := &http.Client{
		Transport: &loggingTransport{
			Transport: transport,
		},
	}

	presenter := presenter.NewPresenterHTTP(kpu.NewSirekap(client))

	http.HandleFunc("/fetch-votes", presenter.GetVotes)
	http.HandleFunc("/fetch-locations", presenter.GetLocations)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
