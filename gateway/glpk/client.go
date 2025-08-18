package glpk

import "github.com/ourstudio-se/puan-sdk-go/pldag"

type Client struct {
	BaseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
	}
}

func (c *Client) Solve(
	system pldag.Polyhedron,
	variables []string,
	objective map[string]int,
) (string, error) {
	panic("not implemented")
}
