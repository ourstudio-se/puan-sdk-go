package pldag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolyhedron_Shape(t *testing.T) {
	tests := []struct {
		name    string
		aMatrix [][]int
		want    Shape
	}{
		{
			name: "valid polyhedron",
			aMatrix: [][]int{
				{1, 1},
			},
			want: Shape{1, 2},
		},
		{
			name:    "nil polyhedron",
			aMatrix: nil,
			want:    Shape{},
		},
		{
			name:    "empty polyhedron",
			aMatrix: [][]int{},
			want:    Shape{},
		},
		{
			name:    "empty polyhedron",
			aMatrix: [][]int{{}, {}},
			want:    Shape{2, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Polyhedron{
				aMatrix: tt.aMatrix,
			}
			assert.Equalf(t, tt.want, p.shape(), "shape()")
		})
	}
}

func TestPolyhedron_SparseMatrix(t *testing.T) {
	tests := []struct {
		name    string
		aMatrix [][]int
		want    SparseMatrix
	}{
		{
			name: "valid polyhedron",
			aMatrix: [][]int{
				{1, 1},
			},
			want: SparseMatrix{
				rows:    []int{0, 0},
				columns: []int{0, 1},
				values:  []int{1, 1},
				shape:   Shape{1, 2},
			},
		},
		{
			name: "valid polyhedron",
			aMatrix: [][]int{
				{1, 1, 2},
				{1, 1, 0},
			},
			want: SparseMatrix{
				rows:    []int{0, 0, 0, 1, 1},
				columns: []int{0, 1, 2, 0, 1},
				values:  []int{1, 1, 2, 1, 1},
				shape:   Shape{2, 3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Polyhedron{
				aMatrix: tt.aMatrix,
			}
			assert.Equalf(t, tt.want, p.SparseMatrix(), "SparseMatrix()")
		})
	}
}
