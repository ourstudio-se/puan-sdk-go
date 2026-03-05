package solver

import (
	"net/http"

	glpk "github.com/ourstudio-se/puan-sdk-go/internal/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

func NewClient(
	baseURL string,
	apiKey string,
	client *http.Client,
) puan.SolverClient {
	return glpk.NewClient(
		baseURL,
		apiKey,
		client,
	)
}
