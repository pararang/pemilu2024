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
