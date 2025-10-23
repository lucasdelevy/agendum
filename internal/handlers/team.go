package handlers

import (
	"encoding/json"
	"net/http"
)

type Team struct {
	TeamID  string   `json:"team_id"`
	Name    string   `json:"name"`
	Admins  []string `json:"admins"`
	Members []string `json:"members"`
}

func ListTeams(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For local development, return mock data
	teams := []Team{
		{
			TeamID:  "team-1",
			Name:    "Development Team",
			Admins:  []string{"john_doe"},
			Members: []string{"alice_jones", "bob_wilson"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(teams)
}