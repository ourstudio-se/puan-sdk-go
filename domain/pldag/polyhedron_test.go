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
			name: "populated polyhedron",
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
			name:    "empty rows in polyhedron",
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
			name: "single row",
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
			name: "many rows",
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

func Test_AddEmptyColumn_givenSingleRow(t *testing.T) {
	theories := []struct {
		name    string
		aMatrix [][]int
		want    [][]int
	}{
		{
			name: "add to single row",
			aMatrix: [][]int{
				{1, 1},
			},
			want: [][]int{
				{1, 1, 0},
			},
		},
		{
			name: "add to multiple rows",
			aMatrix: [][]int{
				{1, 1},
				{2, 2},
				{3, 3},
			},
			want: [][]int{
				{1, 1, 0},
				{2, 2, 0},
				{3, 3, 0},
			},
		},
		{
			name:    "add to empty matrix",
			aMatrix: [][]int{},
			want:    [][]int{},
		},
	}

	for _, tt := range theories {
		t.Run(tt.name, func(t *testing.T) {
			p := Polyhedron{
				aMatrix: tt.aMatrix,
			}
			p.AddEmptyColumn()
			assert.Equal(t, tt.want, p.aMatrix)
		})
	}
}

func Test_Extend_extendToPopulatedPolyhedron(t *testing.T) {
	row := []int{0, 0, 1}
	b := Bias(1)

	polyhedron := Polyhedron{
		aMatrix: [][]int{
			{1, 1, 0},
		},
		bVector: []int{1},
	}

	polyhedron.Extend(row, b)
	assert.Equal(t, [][]int{{1, 1, 0}, {0, 0, 1}}, polyhedron.aMatrix)
	assert.Equal(t, []int{1, 1}, polyhedron.bVector)
}

func Test_Extend_extendToEmptyPolyhedron(t *testing.T) {
	row := []int{0, 1}
	b := Bias(1)

	polyhedron := Polyhedron{}

	polyhedron.Extend(row, b)
	assert.Equal(t, [][]int{{0, 1}}, polyhedron.aMatrix)
	assert.Equal(t, []int{1}, polyhedron.bVector)
}
