package controller

import (
	"fmt"
	"log"
	"runtime"

	"github.com/pararang/pemilu2024/kpu"
	"golang.org/x/sync/errgroup"
)

type DistrictTree struct {
	kpu.Location
	Subdistrict []kpu.Location `json:"desa_kelurahan"`
}

type CityTree struct {
	kpu.Location
	Districts []DistrictTree `json:"kecamatan"`
}

type ProvinceTree struct {
	kpu.Location
	Cities []CityTree `json:"kota_kabupaten"`
}

type Controller struct {
	sirekap *kpu.Sirekap
}

func NewController(sirekap *kpu.Sirekap) *Controller {
	return &Controller{
		sirekap: sirekap,
	}
}

func (c *Controller) maxGoroutine() uint64 {
	memStats := new(runtime.MemStats)
	runtime.ReadMemStats(memStats)
	availableMemory := memStats.Sys // Total available memory in bytes
	log.Println(fmt.Sprintf("availableMemory: %v", availableMemory))


	numCPU := runtime.NumCPU() // Number of CPU cores
	log.Println(fmt.Sprintf("numCPU: %v", numCPU))


	optimalMaxGoroutines := availableMemory / (2 * 1024 * 1024) // Assume each goroutine consumes 128 MB
	log.Println(fmt.Sprintf("optimalMaxGoroutines: %v", optimalMaxGoroutines))

	if optimalMaxGoroutines > uint64(numCPU) {
		optimalMaxGoroutines = uint64(numCPU)
	}

	return optimalMaxGoroutines
}

func (c *Controller) GetLocations(maxLoop uint) ([]ProvinceTree, error) {
	var provinces kpu.Locations
	err := c.sirekap.FetchLocations(&provinces, "0")
	if err != nil {
		return nil, fmt.Errorf("error FetchLocations province: %w", err)
	}

	var (
		locations    = make([]ProvinceTree, len(provinces))
		maxGoroutine = len(provinces) //20 //c.maxGoroutine() //runtime.NumCPU()
		sem          = make(chan struct{}, maxGoroutine)
	)

	// Create an error group
	var eg errgroup.Group

	for idxProv := 0; idxProv < len(provinces); idxProv++ {
		if maxLoop > 0 && maxLoop == uint(idxProv) {
			locations = locations[0:maxLoop]
			break
		}

		sem <- struct{}{} // Acquire semaphore

		idx := idxProv

		eg.Go(func() error {
			defer func() {
				<-sem // Release semaphore
			}()

			var err error
			locations[idx], err = c.getByProvince(provinces[idx])
			if err != nil {
				return fmt.Errorf("error FetchLocations province %s (%s): %w",provinces[idx].Name, provinces[idx].Code, err)
			}
			return nil
		})
	}

	// Wait for all goroutines to finish
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return locations, nil
}

func (c *Controller) getByProvince(province kpu.Location) (provTree ProvinceTree, err error) {
	provTree.Location = province

	var cities kpu.Locations
	err = c.sirekap.FetchLocations(&cities, province.Code)
	if err != nil {
		return provTree, fmt.Errorf("getCities %s: %w", province.Name, err)
	}

	provTree.Cities = make([]CityTree, len(cities))
	for idxCity := 0; idxCity < len(cities); idxCity++ {
		provTree.Cities[idxCity].Location = cities[idxCity]

		var districts kpu.Locations
		err = c.sirekap.FetchLocations(&districts, province.Code, cities[idxCity].Code)
		if err != nil {
			return provTree, fmt.Errorf("getDistricts %s: %w", cities[idxCity].Name, err)
		}

		provTree.Cities[idxCity].Districts = make([]DistrictTree, len(districts))
		for idxDist := 0; idxDist < len(districts); idxDist++ {
			provTree.Cities[idxCity].Districts[idxDist].Location = districts[idxDist]

			var subdistricts kpu.Locations
			err = c.sirekap.FetchLocations(&subdistricts, province.Code, cities[idxCity].Code, districts[idxDist].Code)
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

type Votes struct {
	Votes map[string]interface{} `json:"votes"`
	Docs  []string               `json:"docs"`
}

func (c *Controller) GetVotes(codeTPS string) (Votes, error) {
	data, err := c.sirekap.GetVotesByTPS(codeTPS)
	if err != nil {
		return Votes{}, fmt.Errorf("error on GetVotesByTPS: %w", err)
	}

	mapCand := map[string]string{
		"100025": "AMIN",
		"100026": "PAGI",
		"100027": "GAMA",
	}

	response := Votes{
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

	return response, nil
}

type DataNationwide struct {
	Votes   []Vote  `json:"votes"`
	Progres Progres `json:"progres"`
}

type Progres struct {
	Total   int64 `json:"total"`
	Progres int64 `json:"progres"`
}

type Vote struct {
	LocationName   string  `json:"location_name"`
	LocationLevel  int64   `json:"location_level"`
	PSU            string  `json:"psu"`
	Amin           int64   `json:"amin"`
	Pagi           int64   `json:"pagi"`
	Gama           int64   `json:"gama"`
	Persen         float64 `json:"persen"`
	StatusProgress bool    `json:"status_progress"`
}

func (c *Controller) GetVotesNationwide() (kpu.ResponseDataPresidentialNationwide, error) {
	data, err := c.sirekap.GetVotesPresidentialNationwide()
	if err != nil {
		return kpu.ResponseDataPresidentialNationwide{}, fmt.Errorf("error on GetVotesByTPS: %w", err)
	}
	//TODO: transform to DataNationwide
	return data, nil
}
