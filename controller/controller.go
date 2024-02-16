package controller

import (
	"fmt"
	"runtime"

	"github.com/pararang/pemilu2024/kpu"
	"golang.org/x/sync/errgroup"
)

type DistrictTree struct {
	kpu.Location
	Subdistrict kpu.Locations `json:"desa_kelurahan"`
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

func (h *Controller) GetLocations() ([]ProvinceTree, error) {
	var provinces kpu.Locations
	err := h.sirekap.FetchLocations(&provinces, "0")
	if err != nil {
		return nil, fmt.Errorf("error FetchLocations province: %w", err)
	}

	var (
		locations    = make([]ProvinceTree, len(provinces))
		maxGoroutine = runtime.NumCPU()
		sem          = make(chan struct{}, maxGoroutine)
	)

	// Create an error group
	var eg errgroup.Group

	for idxProv := 0; idxProv < len(provinces); idxProv++ {
		sem <- struct{}{} // Acquire semaphore

		idx := idxProv

		eg.Go(func() error {
			defer func() {
				<-sem // Release semaphore
			}()

			var err error
			locations[idx], err = h.getByProvince(provinces[idx])
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

func (h *Controller) getByProvince(province kpu.Location) (provTree ProvinceTree, err error) {
	provTree.Location = province

	var cities kpu.Locations
	err = h.sirekap.FetchLocations(&cities, province.Code)
	if err != nil {
		return provTree, fmt.Errorf("getCities %s: %w", province.Name, err)
	}

	provTree.Cities = make([]CityTree, len(cities))
	for idxCity := 0; idxCity < len(cities); idxCity++ {
		provTree.Cities[idxCity].Location = cities[idxCity]

		var districts kpu.Locations
		err = h.sirekap.FetchLocations(&districts, province.Code, cities[idxCity].Code)
		if err != nil {
			return provTree, fmt.Errorf("getDistricts %s: %w", cities[idxCity].Name, err)
		}

		provTree.Cities[idxCity].Districts = make([]DistrictTree, len(districts))
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