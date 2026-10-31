package kev

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRankSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/systemone" {
			t.Errorf("path = %s", r.URL.Path)
		}
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode: %v", err)
		}
		q, ok := req.Questions["ranking"]
		if !ok || q.Type != "choice" || len(q.Criteria) != 2 {
			t.Errorf("unexpected questions: %+v", req.Questions)
		}
		_ = json.NewEncoder(w).Encode(response{Answers: map[string]choiceAnswer{
			"ranking": {
				Choice:        "t-b",
				Confidence:    floatPtr(0.91),
				Probabilities: map[string]float64{"t-a": 0.2, "t-b": 0.8},
			},
		}})
	}))
	defer server.Close()

	client := New(server.URL, 5*time.Second)
	ranking, err := client.Rank(context.Background(), "log", map[string]string{"t-a": "first", "t-b": "second"}, "rank worst to best")
	if err != nil {
		t.Fatalf("Rank: %v", err)
	}
	if ranking.Choice != "t-b" {
		t.Errorf("choice = %s", ranking.Choice)
	}
	if ranking.Confidence == nil || *ranking.Confidence != 0.91 {
		t.Errorf("confidence = %v", ranking.Confidence)
	}
	if ranking.Probabilities["t-a"] != 0.2 || ranking.Probabilities["t-b"] != 0.8 {
		t.Errorf("probabilities = %v", ranking.Probabilities)
	}
}

func TestRankRejectsBadResponses(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   response
	}{
		{"server error", http.StatusInternalServerError, response{}},
		{"missing answer", http.StatusOK, response{Answers: map[string]choiceAnswer{}}},
		{"choice outside criteria", http.StatusOK, response{Answers: map[string]choiceAnswer{
			"ranking": {Choice: "t-x", Probabilities: map[string]float64{"t-x": 1}},
		}}},
		{"empty probabilities", http.StatusOK, response{Answers: map[string]choiceAnswer{
			"ranking": {Choice: "t-a", Probabilities: map[string]float64{}},
		}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_ = json.NewEncoder(w).Encode(tc.body)
			}))
			defer server.Close()
			client := New(server.URL, 5*time.Second)
			if _, err := client.Rank(context.Background(), "log", map[string]string{"t-a": "only"}, "rank"); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestRankEmptyCriteria(t *testing.T) {
	client := New("http://127.0.0.1:0", time.Second)
	if _, err := client.Rank(context.Background(), "log", nil, "rank"); err == nil {
		t.Fatal("expected error")
	}
}

func floatPtr(v float64) *float64 {
	return &v
}
