package glpk

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/pldag"
)

type Client struct {
	BaseURL string

	http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		Client:  http.Client{},
	}
}

func (c *Client) Solve(
	polyhedron pldag.Polyhedron,
	variables []string,
	objective map[string]int,
) (SolveResponse, error) {
	request := newSolveRequest(polyhedron, variables, objective)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return SolveResponse{}, errors.Errorf("failed to marshal request: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/solve", bytes.NewBuffer(jsonData))
	if err != nil {
		return SolveResponse{}, errors.Errorf("failed to create request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return SolveResponse{}, errors.Errorf("failed to make request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return SolveResponse{},
			errors.Errorf(
				"request failed with status %d: %s", resp.StatusCode,
				string(body),
			)
	}

	var solveResp SolveResponse
	if err = json.NewDecoder(resp.Body).Decode(&solveResp); err != nil {
		return SolveResponse{}, errors.Errorf("failed to decode response: %w", err)
	}

	return solveResp, nil
}
