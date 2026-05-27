package glpk

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/puan"
)

type Client struct {
	*http.Client

	baseURL string
	apiKey  string
}

func NewDefaultClient(baseURL string) *Client {
	return &Client{
		Client:  &http.Client{},
		baseURL: baseURL,
	}
}

func NewClient(
	baseURL string,
	apiKey string,
	client *http.Client,
) *Client {
	return &Client{
		Client:  client,
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

func (c *Client) Solve(
	query *puan.Query,
) (puan.Solution, error) {
	payload := newSolveRequestFromQuery(query)

	request, err := c.newRequest(payload)
	if err != nil {
		return puan.Solution{}, err
	}

	response, err := c.doSolveRequest(request)
	if err != nil {
		return puan.Solution{}, err
	}

	return response.getSingleSolution()
}

func (c *Client) newRequest(body SolveRequest) (*http.Request, error) {
	buffer, err := body.asBufferedBytes()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/solve", buffer)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", c.apiKey)

	return req, nil
}

func (c *Client) doSolveRequest(request *http.Request) (SolutionResponse, error) {
	response, err := c.Do(request)
	if err != nil {
		return SolutionResponse{}, errors.Wrap(err, 0)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return SolutionResponse{},
			errors.Errorf(
				"body failed with status %d: %s", response.StatusCode,
				string(body),
			)
	}

	var solution SolutionResponse
	if err = json.NewDecoder(response.Body).Decode(&solution); err != nil {
		return SolutionResponse{}, errors.Wrap(err, 0)
	}

	return solution, nil
}

func (c *Client) SolveWithManyWeights(
	query *puan.MultiWeightQuery,
) ([]puan.Solution, error) {
	payload := newSolveRequestFromMultiQuery(query)

	request, err := c.newRequest(payload)
	if err != nil {
		return nil, err
	}

	response, err := c.doSolveRequest(request)
	if err != nil {
		return nil, err
	}

	wantCount := len(query.WeightGroups())
	solutions, err := response.getManySolutions(wantCount)
	if err != nil {
		return nil, err
	}

	return solutions, nil
}
