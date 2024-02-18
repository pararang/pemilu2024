package kpu

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type ResponseDataTPS struct {
	Chart        map[string]int64 `json:"chart"`
	Images       []string         `json:"images"`
	Administrasi interface{}      `json:"administrasi"`
	PSU          interface{}      `json:"psu"`
	Ts           string           `json:"ts"`
	StatusSuara  bool             `json:"status_suara"`
	StatusAdm    bool             `json:"status_adm"`
}

type Location struct {
	Name  string `json:"nama"`
	ID    int64  `json:"id"`
	Code  string `json:"kode"`
	Level int64  `json:"tingkat"`
}

type Locations []Location

type Sirekap struct {
	host string
	http *http.Client
}

func NewSirekap(httpClient *http.Client) *Sirekap {
	return &Sirekap{
		host: "https://sirekap-obj-data.kpu.go.id",
		http: httpClient,
	}
}

func (s *Sirekap) GetVotesByTPS(tpsCode string) (ResponseDataTPS, error) {
	if len(tpsCode) != 13 {
		return ResponseDataTPS{}, errors.New("invalid code TPS, expect 13 chars")
	}

	var votes ResponseDataTPS
	err := s.fetchVotes(&votes, tpsCode[0:2], tpsCode[0:4], tpsCode[0:6], tpsCode[0:10], tpsCode)
	if err != nil {
		return ResponseDataTPS{}, fmt.Errorf("error on fetchVotes: %w", err)
	}

	return votes, nil
}

func (s *Sirekap) FetchLocations(dest *Locations, dynamicPaths ...string) error {
	basePathLocation, err := url.JoinPath(s.host, "wilayah/pemilu/ppwp")
	if err != nil {
		return fmt.Errorf("error on build base path locations: %w", err)
	}

	source, err := url.JoinPath(basePathLocation, dynamicPaths...)
	if err != nil {
		return fmt.Errorf("error on build full URL: %w", err)
	}

	resp, err := s.http.Get(fmt.Sprintf("%s.%s", source, "json"))
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

func (s *Sirekap) fetchVotes(dest any, dynamicPaths ...string) error {
	// "https://sirekap-obj-data.kpu.go.id/pemilu/hhcw/ppwp/73/7371/737114/7371141006/7371141006002.json"
	basePathVote, err := url.JoinPath(s.host, "pemilu/hhcw/ppwp")
	if err != nil {
		return fmt.Errorf("error on build base path votes: %w", err)
	}

	source, err := url.JoinPath(basePathVote, dynamicPaths...)
	if err != nil {
		return fmt.Errorf("error on build full URL: %w", err)
	}

	resp, err := s.http.Get(fmt.Sprintf("%s.%s", source, "json"))
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

type ResponseDataNationwide struct {
	Ts      string             `json:"ts"`
	PSU     PSU                `json:"psu"`
	Mode    string             `json:"mode"`
	Chart   map[string]float64 `json:"chart"`
	Table   map[string]Table   `json:"table"`
	Progres Progres            `json:"progres"`
}

type Progres struct {
	Total   int64 `json:"total"`
	Progres int64 `json:"progres"`
}

type Table struct {
	The100025      *int64  `json:"100025,omitempty"`
	The100026      *int64  `json:"100026,omitempty"`
	The100027      *int64  `json:"100027,omitempty"`
	PSU            PSU     `json:"psu"`
	Persen         float64 `json:"persen"`
	StatusProgress bool    `json:"status_progress"`
}

type PSU string

const (
	Reguler PSU = "Reguler"
)

// https://sirekap-obj-data.kpu.go.id/pemilu/hhcw/ppwp.json
func (s *Sirekap) GetVotesNationwide() (ResponseDataNationwide, error) {
	var votes ResponseDataNationwide
	err := s.fetchVotes(&votes)
	if err != nil {
		return ResponseDataNationwide{}, fmt.Errorf("error on fetchVotes: %w", err)
	}

	return votes, nil
}
