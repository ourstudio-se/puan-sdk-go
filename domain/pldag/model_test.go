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
		c    coefficientValues
		want coefficientValues
	}{
		{
			name: "should negate all values",
			c: coefficientValues{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			want: coefficientValues{
				"a": -1,
				"b": -2,
				"c": -3,
			},
		},
		{
			name: "empty coefficientValues should return empty",
			c:    coefficientValues{},
			want: coefficientValues{},
		},
		{
			name: "nil coefficientValues should return empty",
			c:    nil,
			want: coefficientValues{},
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
		c    coefficientValues
		want int
	}{
		{
			name: "given only positive values",
			c: coefficientValues{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			want: 6,
		},
		{
			name: "given only negative values",
			c: coefficientValues{
				"a": -1,
				"b": -2,
				"c": -3,
			},
			want: 6,
		},
		{
			name: "given mixed signed values",
			c: coefficientValues{
				"a": -1,
				"b": 2,
				"c": -3,
			},
			want: 4,
		},
		{
			name: "empty values",
			c:    coefficientValues{},
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

func Test_newAtLeastConstraint(t *testing.T) {
	tests := []struct {
		name      string
		variables []string
		amount    int
		want      Constraint
		wantErr   bool
	}{
		{
			name:      "should create constraint",
			variables: []string{"a", "b", "c"},
			amount:    2,
			want: Constraint{
				id: "id",
				coefficients: coefficientValues{
					"a": -1,
					"b": -1,
					"c": -1,
				},
				bias: Bias(-2),
			},
			wantErr: false,
		},
		{
			name:      "amount larger than number of variables should return error",
			variables: []string{"a"},
			amount:    2,
			want:      Constraint{},
			wantErr:   true,
		},
		{
			name:      "amount equal to the number of variables should return constraint",
			variables: []string{"a", "b"},
			amount:    2,
			want: Constraint{
				id: "id",
				coefficients: coefficientValues{
					"a": -1,
					"b": -1,
				},
				bias: Bias(-2),
			},
			wantErr: false,
		},
		{
			name:      "no variables should return constraint",
			variables: []string{},
			amount:    0,
			want: Constraint{
				id:           "id",
				coefficients: coefficientValues{},
				bias:         Bias(0),
			},
			wantErr: false,
		},
		{
			name:      "nil variables should return constraint",
			variables: nil,
			amount:    0,
			want: Constraint{
				id:           "id",
				coefficients: coefficientValues{},
				bias:         Bias(0),
			},
			wantErr: false,
		},
		{
			name:      "negative amount should return error",
			variables: nil,
			amount:    -1,
			want:      Constraint{},
			wantErr:   true,
		},
		{
			name:      "duplicated variables should return error",
			variables: []string{"a", "a"},
			amount:    2,
			want:      Constraint{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newAtLeastConstraint(tt.variables, tt.amount)
			if tt.wantErr && err == nil {
				t.Errorf("newAtLeastConstraint() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("newAtLeastConstraint() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want.bias, got.bias, "Bias should match")
			assert.Equal(t, tt.want.coefficients, got.coefficients, "Coefficients should match")
		})
	}
}

func Test_newAtMostConstraint(t *testing.T) {
	tests := []struct {
		name      string
		variables []string
		amount    int
		want      Constraint
		wantErr   bool
	}{
		{
			name:      "should create constraint",
			variables: []string{"a", "b", "c"},
			amount:    2,
			want: Constraint{
				id: "id",
				coefficients: coefficientValues{
					"a": 1,
					"b": 1,
					"c": 1,
				},
				bias: Bias(2),
			},
			wantErr: false,
		},
		{
			name:      "amount larger than number of variables should return error",
			variables: []string{"a"},
			amount:    2,
			want:      Constraint{},
			wantErr:   true,
		},
		{
			name:      "amount equal to the number of variables should return constraint",
			variables: []string{"a", "b"},
			amount:    2,
			want: Constraint{
				id: "id",
				coefficients: coefficientValues{
					"a": 1,
					"b": 1,
				},
				bias: Bias(2),
			},
			wantErr: false,
		},
		{
			name:      "no variables should return constraint",
			variables: []string{},
			amount:    0,
			want: Constraint{
				id:           "id",
				coefficients: coefficientValues{},
				bias:         Bias(0),
			},
			wantErr: false,
		},
		{
			name:      "nil variables should return constraint",
			variables: nil,
			amount:    0,
			want: Constraint{
				id:           "id",
				coefficients: coefficientValues{},
				bias:         Bias(0),
			},
			wantErr: false,
		},
		{
			name:      "negative amount should return error",
			variables: nil,
			amount:    -1,
			want:      Constraint{},
			wantErr:   true,
		},
		{
			name:      "duplicated variables should return error",
			variables: []string{"a", "a"},
			amount:    2,
			want:      Constraint{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newAtMostConstraint(tt.variables, tt.amount)
			if tt.wantErr && err == nil {
				t.Errorf("newAtMostConstraint() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("newAtMostConstraint() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want.bias, got.bias, "Bias should match")
			assert.Equal(t, tt.want.coefficients, got.coefficients, "Coefficients should match")
		})
	}
}

func Test_newConstraintID(t *testing.T) {
	tests := []struct {
		name         string
		coefficients coefficientValues
		bias         Bias
		want         string
	}{
		{
			name: "should create id",
			coefficients: coefficientValues{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			bias: 1,
			want: "faba03a1732d697d527760d2c395b1ef6b842115",
		},
		{
			name: "should create id",
			coefficients: coefficientValues{
				"c": 3,
				"b": 2,
				"a": 1,
			},
			bias: 1,
			want: "faba03a1732d697d527760d2c395b1ef6b842115",
		},
		{
			name: "should create id",
			coefficients: coefficientValues{
				"x": 3,
				"y": 2,
				"z": 10,
				"a": 5,
			},
			bias: 20,
			want: "46e3905695e1a101bb46ff5580774c5eb92601a1",
		},
		{
			name: "should create id",
			coefficients: coefficientValues{
				"z": 10,
				"x": 3,
				"a": 5,
				"y": 2,
			},
			bias: 20,
			want: "46e3905695e1a101bb46ff5580774c5eb92601a1",
		},
		{
			name:         "empty coefficients should create id",
			coefficients: coefficientValues{},
			bias:         0,
			want:         "b6589fc6ab0dc82cf12099d1c2d40ab994e8410c",
		},
		{
			name:         "nil coefficients should create id",
			coefficients: nil,
			bias:         0,
			want:         "b6589fc6ab0dc82cf12099d1c2d40ab994e8410c",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(
				t,
				tt.want,
				newConstraintID(tt.coefficients, tt.bias),
				"newConstraintID(%v, %v)",
				tt.coefficients,
				tt.bias,
			)
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
					coefficients: coefficientValues{
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
				coefficients: coefficientValues{
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
