package glpk

import (
	"reflect"
	"testing"

	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
)

func Test_newRequestPayload(t *testing.T) {
	aMatrix := [][]int{{1, 2}, {3, 4}}
	bVector := []int{5, 6}
	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)
	variableIDs := []string{"x", "y"}
	objective := Objective(map[string]int{"x": 2, "y": 4})

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
		Direction:  "maximize",
	}

	if !reflect.DeepEqual(want, newRequestPayload(polyhedron, variableIDs, objective)) {
		t.Errorf("want %v, got %v", want, newRequestPayload(polyhedron, variableIDs, objective))
	}
}
