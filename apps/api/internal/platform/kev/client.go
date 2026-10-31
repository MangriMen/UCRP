package kev

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

type Ranking struct {
	Choice        string
	Confidence    *float64
	Probabilities map[string]float64
}

type question struct {
	Type         string            `json:"type"`
	Instructions string            `json:"instructions"`
	Criteria     map[string]string `json:"criteria"`
}

type request struct {
	State     string              `json:"state"`
	Questions map[string]question `json:"questions"`
}

type choiceAnswer struct {
	Choice        string             `json:"choice"`
	Confidence    *float64           `json:"confidence"`
	Probabilities map[string]float64 `json:"probabilities"`
}

type response struct {
	Answers map[string]choiceAnswer `json:"answers"`
}

const questionKey = "ranking"

func (c *Client) Rank(ctx context.Context, state string, criteria map[string]string, instructions string) (*Ranking, error) {
	if len(criteria) == 0 {
		return nil, fmt.Errorf("kev: criteria must not be empty")
	}
	body, err := json.Marshal(request{
		State: state,
		Questions: map[string]question{
			questionKey: {Type: "choice", Instructions: instructions, Criteria: criteria},
		},
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/systemone", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kev: unexpected status %d", resp.StatusCode)
	}
	var out response
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	answer, ok := out.Answers[questionKey]
	if !ok {
		return nil, fmt.Errorf("kev: answer %q missing", questionKey)
	}
	if answer.Choice == "" {
		return nil, fmt.Errorf("kev: empty choice")
	}
	if _, ok := criteria[answer.Choice]; !ok {
		return nil, fmt.Errorf("kev: choice %q outside criteria", answer.Choice)
	}
	if len(answer.Probabilities) == 0 {
		return nil, fmt.Errorf("kev: empty probabilities")
	}
	return &Ranking{
		Choice:        answer.Choice,
		Confidence:    answer.Confidence,
		Probabilities: answer.Probabilities,
	}, nil
}
