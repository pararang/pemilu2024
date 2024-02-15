package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"pemilu2024/kpu"
	"runtime"
	"sync"
)

type Location struct {
	Name  string `json:"nama"`
	ID    int64  `json:"id"`
	Code  string `json:"kode"`
	Level int64  `json:"tingkat"`
}

type Locations []Location

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
	baseURL = "https://sirekap-obj-data.kpu.go.id"
)


// https://sirekap-obj-2024.kpu.go.id/json-public/wilayah/pemilu/ppwp/73/7371/737114.json
// https://sirekap-obj-data.kpu.go.id/wilayah/pemilu/ppwp/12/1212/121203.json
func (h *Handler) fetchLocations(client *http.Client, dest *Locations, dynamicPaths ...string) error {
	basePathLocation, err := url.JoinPath(baseURL, "wilayah/pemilu/ppwp")
	if err != nil {
		return fmt.Errorf("error on build base path locations: %w", err)
	}

	source, err := url.JoinPath(basePathLocation, dynamicPaths...)
	if err != nil {
		return fmt.Errorf("error on build full URL: %w", err)
	}

	resp, err := client.Get(fmt.Sprintf("%s.%s", source, "json"))
	if err != nil {
		return fmt.Errorf("error on http get: %w", err)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(dest)
	if err != nil {
		return fmt.Errorf("error on decode response: %w", err)
	}

	return nil
}

func (h *Handler) getByProvince(client *http.Client, province Location) (provTree provinceTree, err error) {
	provTree.Location = province

	var cities Locations
	err = h.fetchLocations(client, &cities, province.Code)
	if err != nil {
		return provTree, fmt.Errorf("getCities %s: %w", province.Name, err)
	}

	provTree.Cities = make([]cityTree, len(cities))
	for idxCity := 0; idxCity < len(cities); idxCity++ {
		provTree.Cities[idxCity].Location = cities[idxCity]

		var districts Locations
		err = h.fetchLocations(client, &districts, province.Code, cities[idxCity].Code)
		if err != nil {
			return provTree, fmt.Errorf("getDistricts %s: %w", cities[idxCity].Name, err)
		}

		provTree.Cities[idxCity].Districts = make([]districtTree, len(districts))
		for idxDist := 0; idxDist < len(districts); idxDist++ {
			provTree.Cities[idxCity].Districts[idxDist].Location = districts[idxDist]

			var subdistricts Locations
			err = h.fetchLocations(client, &subdistricts, province.Code, cities[idxCity].Code, districts[idxDist].Code)
			if err != nil {
				return provTree, fmt.Errorf("getSubdistricts %s: %w", cities[idxCity].Name, err)
			}

			provTree.Cities[idxCity].Districts[idxDist].Subdistrict = make([]Location, len(subdistricts))
			for idxSubdist := 0; idxSubdist < len(subdistricts); idxSubdist++ {
				provTree.Cities[idxCity].Districts[idxDist].Subdistrict[idxSubdist] = subdistricts[idxSubdist]
			}
		}

	}

	return provTree, nil
}

type Handler struct {
	sirekap *kpu.Sirekap
}

func (h *Handler) GetVotes(w http.ResponseWriter, r *http.Request) {
	data, err := h.sirekap.GetVotesByTPS(r.URL.Query().Get("tps"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mapCand := map[string]string{
		"100025": "AMIN",
		"100026": "PAGI",
		"100027": "GAMA",
	}

	var response = struct {
		Votes map[string]interface{} `json:"votes"`
		Docs  []string               `json:"docs"`
	}{
		Votes: make(map[string]interface{}),
	}
	for code, votes := range data.Chart {
		cand, ok := mapCand[code]
		if !ok {
			continue
		}

		response.Votes[cand] = votes
	}

	response.Docs = data.Images

	// Set the content type header
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response to the client
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetLocations(w http.ResponseWriter, r *http.Request) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Create a new HTTP client with the custom transport
	client := &http.Client{
		Transport: &loggingTransport{
			Transport: transport,
		},
	}

	var provinces Locations
	err := h.fetchLocations(client, &provinces, "0")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var (
		locations    = make([]provinceTree, len(provinces))
		maxGoroutine = runtime.NumCPU()
		sem          = make(chan struct{}, maxGoroutine)
		wg           sync.WaitGroup
	)

	for idxProv := 0; idxProv < len(provinces); idxProv++ {
		sem <- struct{}{} // Acquire semaphore

		wg.Add(1)
		go func(idx int) {
			defer func() {
				<-sem // Release semaphore
				wg.Done()
			}()

			var err error
			locations[idx], err = h.getByProvince(client, provinces[idx])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}(idxProv)
	}

	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locations)
}