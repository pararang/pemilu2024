package handler

import (
	"encoding/json"
	"fmt"
	"github.com/pararang/pemilu2024/kpu"
	"net/http"
	"runtime"
	"sync"
)

type districtTree struct {
	kpu.Location
	Subdistrict kpu.Locations `json:"desa_kelurahan"`
}

type cityTree struct {
	kpu.Location
	Districts []districtTree `json:"kecamatan"`
}

type provinceTree struct {
	kpu.Location
	Cities []cityTree `json:"kota_kabupaten"`
}

type Handler struct {
	sirekap *kpu.Sirekap
}

func NewHandler(sirekap *kpu.Sirekap) *Handler {
	return &Handler{
		sirekap: sirekap,
	}
}

func (h *Handler) getByProvince(province kpu.Location) (provTree provinceTree, err error) {
	provTree.Location = province

	var cities kpu.Locations
	err = h.sirekap.FetchLocations(&cities, province.Code)
	if err != nil {
		return provTree, fmt.Errorf("getCities %s: %w", province.Name, err)
	}

	provTree.Cities = make([]cityTree, len(cities))
	for idxCity := 0; idxCity < len(cities); idxCity++ {
		provTree.Cities[idxCity].Location = cities[idxCity]

		var districts kpu.Locations
		err = h.sirekap.FetchLocations(&districts, province.Code, cities[idxCity].Code)
		if err != nil {
			return provTree, fmt.Errorf("getDistricts %s: %w", cities[idxCity].Name, err)
		}

		provTree.Cities[idxCity].Districts = make([]districtTree, len(districts))
		for idxDist := 0; idxDist < len(districts); idxDist++ {
			provTree.Cities[idxCity].Districts[idxDist].Location = districts[idxDist]

			var subdistricts kpu.Locations
			err = h.sirekap.FetchLocations(&subdistricts, province.Code, cities[idxCity].Code, districts[idxDist].Code)
			if err != nil {
				return provTree, fmt.Errorf("getSubdistricts %s: %w", cities[idxCity].Name, err)
			}

			provTree.Cities[idxCity].Districts[idxDist].Subdistrict = make([]kpu.Location, len(subdistricts))
			for idxSubdist := 0; idxSubdist < len(subdistricts); idxSubdist++ {
				provTree.Cities[idxCity].Districts[idxDist].Subdistrict[idxSubdist] = subdistricts[idxSubdist]
			}
		}

	}

	return provTree, nil
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
	var provinces kpu.Locations
	err := h.sirekap.FetchLocations(&provinces, "0")
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
			locations[idx], err = h.getByProvince(provinces[idx])
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
