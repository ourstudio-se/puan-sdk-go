package glpk

import (
	"github.com/ourstudio-se/puan-sdk-go/pldag"
)

func newSolveRequest(
	polyhedron pldag.Polyhedron,
	variables []string,
	objective ...Objective,
) *SolveRequest {
	sparseMatrix := polyhedron.SparseMatrix()
	b := polyhedron.B()

	var tmpVariables []Variable
	for _, v := range variables {
		tmpVariables = append(tmpVariables, Variable{
			ID:    v,
			Bound: [2]int{0, 1},
		})
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

		Objectives: append([]Objective{}, objective...),
		Direction:  "maximize",
	}

	return request
}

type SolveRequest struct {
	Polyhedron Polyhedron  `json:"polyhedron"`
	Objectives []Objective `json:"objectives"`
	Direction  string      `json:"direction"`
}

type Polyhedron struct {
	A         SparseMatrix `json:"A"`
	B         []int        `json:"b"`
	Variables []Variable   `json:"variables"`
}
type SparseMatrix struct {
	Rows  []int `json:"rows"`
	Cols  []int `json:"cols"`
	Vals  []int `json:"vals"`
	Shape Shape `json:"shape"`
}

type Shape struct {
	Nrows int `json:"nrows"`
	Ncols int `json:"ncols"`
}

type Variable struct {
	ID    string `json:"id"`
	Bound [2]int `json:"bound"`
}

type Objective map[string]int

type SolveResponse struct {
	Solutions []Solution `json:"solutions"`
}

type Solution struct {
	Error     *string        `json:"error"`
	Objective float64        `json:"objective"`
	Solution  SolutionValues `json:"solution"`
	Status    string         `json:"status"`
}

type SolutionValues map[string]int
