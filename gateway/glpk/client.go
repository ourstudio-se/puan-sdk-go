package glpk

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	sparseMatrix := polyhedron.SparseMatrix()
	b := polyhedron.B()

	var tmpVariables []Variable
	for _, v := range variables {
		tmpVariables = append(tmpVariables, Variable{
			ID:    v,
			Bound: [2]int{0, 1},
		})
	}

	tmpObjective := Objective{}
	for _, v := range variables {
		tmpObjective[v] = 1
	}

	request := &SolveRequest{
		Polyhedron: Polyhedron{
			A: SparseMatrix{
				Rows: sparseMatrix.Row,
				Cols: sparseMatrix.Column,
				Vals: sparseMatrix.Value,
				Shape: Shape{
					Nrows: polyhedron.Shape().NrOfColumns(),
					Ncols: polyhedron.Shape().NrOfRows(),
				},
			},
			B:         b,
			Variables: tmpVariables,
		},

		Objectives: []Objective{objective},
		Direction:  "maximize",
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return SolveResponse{}, errors.Errorf("failed to marshal request: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/solve", bytes.NewBuffer(jsonData))
	if err != nil {
		return SolveResponse{}, errors.Errorf("failed to create request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return SolveResponse{}, errors.Errorf("failed to make request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return SolveResponse{}, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var solveResp SolveResponse
	if err = json.NewDecoder(resp.Body).Decode(&solveResp); err != nil {
		return SolveResponse{}, errors.Errorf("failed to decode response: %w", err)
	}

	return solveResp, nil
}
