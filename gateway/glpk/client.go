package glpk

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
	"github.com/ourstudio-se/puan-sdk-go/domain/puan"
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
) (puan.Solution, error) {
	payload := newRequestPayload(polyhedron, variables, objective)

	request, err := c.newRequest(payload)
	if err != nil {
		return puan.Solution{}, err
	}

	resp, err := c.Do(request)
	if err != nil {
		return puan.Solution{}, errors.Wrap(err, 0)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return puan.Solution{},
			errors.Errorf(
				"body failed with status %d: %s", resp.StatusCode,
				string(body),
			)
	}

	var solveResp SolutionResponse
	if err = json.NewDecoder(resp.Body).Decode(&solveResp); err != nil {
		return puan.Solution{}, errors.Wrap(err, 0)
	}

	return solveResp.getSolutionEntity()
}

func newRequestPayload(
	polyhedron pldag.Polyhedron,
	variableIDs []string,
	objective Objective,
) SolveRequest {
	A := toSparseMatrix(polyhedron.SparseMatrix())
	b := polyhedron.B()
	variables := toBooleanVariables(variableIDs)
	objectives := []Objective{objective}

	request := SolveRequest{
		Polyhedron: Polyhedron{
			A:         A,
			B:         b,
			Variables: variables,
		},
		Objectives: objectives,
		Direction:  "maximize",
	}

	return request
}

func (c *Client) newRequest(body SolveRequest) (*http.Request, error) {
	buffer, err := body.asBufferedBytes()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/solve", buffer)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
