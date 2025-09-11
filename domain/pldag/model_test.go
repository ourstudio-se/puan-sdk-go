package pldag

import (
	"reflect"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_coefficientValues_negate(t *testing.T) {
	tests := []struct {
		name string
		c    CoefficientValues
		want CoefficientValues
	}{
		{
			name: "should negate all values",
			c: CoefficientValues{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			want: CoefficientValues{
				"a": -1,
				"b": -2,
				"c": -3,
			},
		},
		{
			name: "empty CoefficientValues should return empty",
			c:    CoefficientValues{},
			want: CoefficientValues{},
		},
		{
			name: "nil CoefficientValues should return empty",
			c:    nil,
			want: CoefficientValues{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.negate(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("negate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_coefficientValues_calculateMaxAbsInnerBound(t *testing.T) {
	tests := []struct {
		name string
		c    CoefficientValues
		want int
	}{
		{
			name: "given only positive values",
			c: CoefficientValues{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			want: 6,
		},
		{
			name: "given only negative values",
			c: CoefficientValues{
				"a": -1,
				"b": -2,
				"c": -3,
			},
			want: 6,
		},
		{
			name: "given mixed signed values",
			c: CoefficientValues{
				"a": -1,
				"b": 2,
				"c": -3,
			},
			want: 4,
		},
		{
			name: "empty values",
			c:    CoefficientValues{},
			want: 0,
		},
		{
			name: "nil values",
			c:    nil,
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.calculateMaxAbsInnerBound(); got != tt.want {
				t.Errorf("calculateMaxAbsInnerBound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBias_negate(t *testing.T) {
	tests := []struct {
		name string
		b    Bias
		want Bias
	}{
		{
			name: "should negate bias",
			b:    1,
			want: -2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.negate(); got != tt.want {
				t.Errorf("negate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_GeneratePolyhedron(t *testing.T) {
	model := New()
	model.SetPrimitives([]string{"x", "y", "z", "k", "w"}...)

	andID, _ := model.SetAnd([]string{"x", "y"}...)
	notID, _ := model.SetNot([]string{"k"}...)
	orID, _ := model.SetOr([]string{"y", "z"}...)

	xorID, _ := model.SetXor([]string{andID, notID, orID}...)
	implyID, _ := model.SetImply("w", xorID)
	_ = model.Assume(implyID)

	lp := model.GeneratePolyhedron()

	expectedVector := []int{0, 1, 1, 2, 4, 0, 1, 1, 1, -1, 0, 0, -2, 1, -1, 0, -1}
	expectedMatrix := [][]int{
		{-1, -1, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0},
		{0, -1, -1, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, -1, -1, -1, 3, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 1, 1, 1, 0, 3, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, -1, -1, 2, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, -1, -1, 2},
		{1, 1, 0, 0, 0, -1, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, -1, 0, 0, -2, 0, 0, 0, 0, 0, 0},
		{0, 1, 1, 0, 0, 0, 0, -2, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 1, 1, 1, -3, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, -1, -1, -1, 0, -5, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, -1, 0, 0},
		{0, 0, 0, 0, -1, 0, 0, 0, 0, 0, 0, -2, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, -2},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, -1},
	}

	assertEqual(t, expectedMatrix, lp.aMatrix, expectedVector, lp.bVector)
}

func assertEqual(
	t *testing.T,
	expectedMatrix,
	actualMatrix [][]int,
	expectedVector,
	actualVector []int,
) {
	sortedActualMatrix := [][]int{}
	sortedActualVector := []int{}
	for _, row := range expectedMatrix {
		found := false
		for j, actualRow := range actualMatrix {
			if slices.Equal(row, actualRow) {
				found = true
				sortedActualVector = append(sortedActualVector, actualVector[j])
				sortedActualMatrix = append(sortedActualMatrix, actualMatrix[j])
			}
		}
		if !found {
			t.Errorf("Expected nrOfRows %v not found in actual matrix", row)
		}
	}

	assert.Equal(t, expectedMatrix, sortedActualMatrix)
	assert.Equal(t, expectedVector, sortedActualVector)
}

func TestValidateAssumedVariables(t *testing.T) {
	tests := []struct {
		name                       string
		existingAssumedConstraints AuxiliaryConstraints
		existingVariables          []string
		assumedVariables           []string
		wantErr                    bool
	}{
		{
			name:                       "valid model",
			existingAssumedConstraints: AuxiliaryConstraints{},
			existingVariables:          []string{"a", "b", "c"},
			assumedVariables:           []string{"a", "b"},
			wantErr:                    false,
		},
		{
			name: "invalid assumed variable again",
			existingAssumedConstraints: AuxiliaryConstraints{
				{
					coefficients: CoefficientValues{
						"a": 1,
						"b": 1,
					},
					bias: Bias(2),
				},
			},
			existingVariables: []string{"a", "b", "c"},
			assumedVariables:  []string{"a", "b"},
			wantErr:           true,
		},
		{
			name:              "invalid assumed non-existing variable",
			existingVariables: []string{"a", "b", "c"},
			assumedVariables:  []string{"x"},
			wantErr:           true,
		},
		{
			name:              "duplicated assumed variable",
			existingVariables: []string{"a", "b", "c"},
			assumedVariables:  []string{"a", "a"},
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{
				variables:         tt.existingVariables,
				assumeConstraints: tt.existingAssumedConstraints,
			}

			err := m.Assume(tt.assumedVariables...)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestModel_newAssumedConstraint(t *testing.T) {
	tests := []struct {
		name      string
		variables []string
		want      AuxiliaryConstraint
	}{
		{
			name:      "valid constraint",
			variables: []string{"a", "b"},
			want: AuxiliaryConstraint{
				coefficients: CoefficientValues{
					"a": -1,
					"b": -1,
				},
				bias: Bias(-2),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{}
			constraint := m.newAssumedConstraint(tt.variables...)
			assert.Equal(t, tt.want, constraint, "Constraint should match")
		})
	}
}
