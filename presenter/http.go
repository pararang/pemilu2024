package presenter

import (
	"encoding/json"
	"net/http"

	"github.com/pararang/pemilu2024/controller"
	"github.com/pararang/pemilu2024/kpu"
)

type Handler struct {
	sirekap *kpu.Sirekap
	control     *controller.Controller
}

func NewPresenterHTTP(sirekap *kpu.Sirekap) *Handler {
	return &Handler{
		sirekap: sirekap,
		control: controller.NewController(sirekap),
	}
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
	locations, err := h.control.GetLocations()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locations)
}
