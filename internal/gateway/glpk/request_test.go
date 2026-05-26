package glpk

import (
	"testing"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
	"github.com/ourstudio-se/puan-sdk-go/puan"
	"github.com/stretchr/testify/assert"
)

func Test_newSolveRequest(t *testing.T) {
	aMatrix := [][]int{{1, 2}, {3, 4}}
	bVector := fake.New[[]int]()
	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)
	variableIDs := []string{"x", "y"}
	objectives := fake.New[[]Objective]()

	want := SolveRequest{
		Polyhedron: Polyhedron{
			A: SparseMatrix{
				Rows:  []int{0, 0, 1, 1},
				Cols:  []int{0, 1, 0, 1},
				Vals:  []int{1, 2, 3, 4},
				Shape: Shape{Nrows: 2, Ncols: 2},
			},
			B: bVector,
			Variables: []Variable{
				{ID: "x", Bound: [2]int{0, 1}},
				{ID: "y", Bound: [2]int{0, 1}},
			},
		},
		Objectives: objectives,
		Direction:  DefaultDirection,
	}

	got := newSolveRequest(polyhedron, variableIDs, objectives...)

	assert.Equal(t, want, got)
}

func Test_newSolveRequestFromQuery_givenQuery_shouldReturnSolveRequest(t *testing.T) {
	aMatrix := [][]int{{1, 2}, {3, 4}}
	bVector := fake.New[[]int]()
	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)
	variableIDs := []string{"x", "y"}
	objective := map[string]int{"x": 2, "y": 4}

	query := puan.NewQuery(polyhedron, variableIDs, objective)

	want := SolveRequest{
		Polyhedron: Polyhedron{
			A: SparseMatrix{
				Rows:  []int{0, 0, 1, 1},
				Cols:  []int{0, 1, 0, 1},
				Vals:  []int{1, 2, 3, 4},
				Shape: Shape{Nrows: 2, Ncols: 2},
			},
			B: bVector,
			Variables: []Variable{
				{
					ID:    "x",
					Bound: [2]int{0, 1},
				},
				{
					ID:    "y",
					Bound: [2]int{0, 1},
				},
			},
		},
		Objectives: []Objective{objective},
		Direction:  DefaultDirection,
	}

	got := newSolveRequestFromQuery(query)

	assert.Equal(t, want, got)
}

func Test_newSolveRequestFromMultiQuery(t *testing.T) {
	aMatrix := [][]int{{1, 2}, {3, 4}}
	bVector := fake.New[[]int]()
	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)
	variableIDs := []string{"x", "y"}
	queryWeights := []weights.Weights{
		{"x": 3, "y": 4},
		{"x": 1, "y": 2},
	}

	query := puan.NewMultiWeightQuery(polyhedron, variableIDs, queryWeights)

	want := SolveRequest{
		Polyhedron: Polyhedron{
			A: SparseMatrix{
				Rows:  []int{0, 0, 1, 1},
				Cols:  []int{0, 1, 0, 1},
				Vals:  []int{1, 2, 3, 4},
				Shape: Shape{Nrows: 2, Ncols: 2},
			},
			B: bVector,
			Variables: []Variable{
				{ID: "x", Bound: [2]int{0, 1}},
				{ID: "y", Bound: [2]int{0, 1}},
			},
		},
		Objectives: []Objective{
			{"x": 3, "y": 4},
			{"x": 1, "y": 2},
		},
		Direction: DefaultDirection,
	}

	got := newSolveRequestFromMultiQuery(query)

	assert.Equal(t, want, got)
}
