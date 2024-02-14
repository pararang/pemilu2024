package pilpres

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
)

func GetVotes(w http.ResponseWriter, r *http.Request) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Fetch data from the URL
	url := "https://sirekap-obj-2024.kpu.go.id/json-public/pemilu/hhcw/ppwp/12/1209/120909/1209091001/1209091001019.json"
	resp, err := client.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Decode the JSON response
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the content type header
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response to the client
	json.NewEncoder(w).Encode(data)
}
