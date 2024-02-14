package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"pemilu2024/pilpres"
)

type Location struct {
	Name  string `json:"nama"`
	ID    int64  `json:"id"`
	Code  string `json:"kode"`
	Level int64  `json:"tingkat"`
}

type districtTree struct {
	Location
	Subdistrict []Location `json:"desa_kelurahan"`
}

type cityTree struct {
	Location
	Districts []districtTree `json:"kecamatan"`
}

type provinceTree struct {
	Location
	Cities []cityTree `json:"kota_kabupaten"`
}

const (
	baseURL            = "https://sirekap-obj-data.kpu.go.id"
	pathMasterLocation = "wilayah/pemilu/ppwp"
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

func GetLocations(w http.ResponseWriter, r *http.Request) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Create a new HTTP client with the custom transport
	client := &http.Client{
		Transport: &loggingTransport{
			Transport: transport,
		},
	}

	// Fetch data from the URL
	pathProvinceList, _ := url.JoinPath(baseURL, pathMasterLocation, "0.json")
	resp, err := client.Get(pathProvinceList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Decode the JSON response
	var provinces []Location
	err = json.NewDecoder(resp.Body).Decode(&provinces)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var locations = make([]provinceTree, len(provinces))

	for i := 0; i < len(provinces); i++ {
		locations[i].Location = provinces[i]
		cities, err := getCities(client, provinces[i].Code)
		if err != nil {
			http.Error(w, fmt.Sprintf("getCities %s: %s", provinces[i].Name, err.Error()), http.StatusInternalServerError)
			return
		}

		locations[i].Cities = make([]cityTree, len(cities))
		for ii := 0; ii < len(cities); ii++ {
			locations[i].Cities[ii].Location = cities[ii]
			districts, err := getDistricts(client, provinces[i].Code, cities[ii].Code)
			if err != nil {
				http.Error(w, fmt.Sprintf("getDistricts %s: %s", cities[ii].Name, err.Error()), http.StatusInternalServerError)
				return
			}

			locations[i].Cities[ii].Districts = make([]districtTree, len(districts))
			for iii := 0; iii < len(districts); iii++ {
				locations[i].Cities[ii].Districts[iii].Location = districts[iii]
				subdistricts, err := getSubdistricts(client, provinces[i].Code, cities[ii].Code, districts[iii].Code)
				if err != nil {
					http.Error(w, fmt.Sprintf("getSubdistricts %s: %s", cities[ii].Name, err.Error()), http.StatusInternalServerError)
					return
				}

				locations[i].Cities[ii].Districts[iii].Subdistrict = make([]Location, len(subdistricts))
				for iiii := 0; iiii < len(subdistricts); iiii++ {
					locations[i].Cities[ii].Districts[iii].Subdistrict[iiii] = subdistricts[iiii]
				}
			}

		}

	}

	// Set the content type header
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response to the client
	json.NewEncoder(w).Encode(locations)
}

func getCities(client *http.Client, provinceCode string) ([]Location, error) {
	// Fetch data from the URL
	log.Printf("\ngetCities for %s", provinceCode)
	fullPath, err := url.JoinPath(baseURL, pathMasterLocation, provinceCode+".json")
	if err != nil {
		return nil, err
	}

	resp, err := client.Get(fullPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the JSON response
	var cities []Location
	err = json.NewDecoder(resp.Body).Decode(&cities)
	if err != nil {
		return nil, err
	}

	return cities, nil
}

func getDistricts(client *http.Client, provinceCode, cityCode string) ([]Location, error) {
	// Fetch data from the URL
	log.Printf("\ngetDistricts for %s:%s", provinceCode, cityCode)
	fullPath, err := url.JoinPath(baseURL, pathMasterLocation, provinceCode, cityCode+".json")
	if err != nil {
		return nil, err
	}

	resp, err := client.Get(fullPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the JSON response
	var districts []Location
	err = json.NewDecoder(resp.Body).Decode(&districts)
	if err != nil {
		return nil, err
	}

	return districts, nil
}

// https://sirekap-obj-2024.kpu.go.id/json-public/wilayah/pemilu/ppwp/73/7371/737114.json
func getSubdistricts(client *http.Client, provinceCode, cityCode, districtCode string) ([]Location, error) {
	// Fetch data from the URL
	log.Printf("\ngetSubdistricts for %s:%s:%s", provinceCode, cityCode, districtCode)
	fullPath, err := url.JoinPath(baseURL, pathMasterLocation, provinceCode, cityCode, districtCode+".json")
	if err != nil {
		return nil, err
	}

	resp, err := client.Get(fullPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the JSON response
	var districts []Location
	err = json.NewDecoder(resp.Body).Decode(&districts)
	if err != nil {
		return nil, err
	}

	return districts, nil
}

func main() {
	// Define the endpoint handler
	http.HandleFunc("/fetch-data", pilpres.GetVotes)
	http.HandleFunc("/fetch-locations", GetLocations)

	// Start the HTTP server
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
