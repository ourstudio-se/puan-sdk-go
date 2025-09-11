package pldag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateConstraintInput(t *testing.T) {
	tests := []struct {
		name      string
		variables []string
		amount    int
		wantErr   bool
	}{
		{
			name:      "valid input",
			variables: []string{"a", "b", "c"},
			amount:    2,
			wantErr:   false,
		},
		{
			name:      "amount larger than number of variables should return error",
			variables: []string{"a"},
			amount:    2,
			wantErr:   true,
		},
		{
			name:      "negative amount should return error",
			variables: nil,
			amount:    -1,
			wantErr:   true,
		},
		{
			name:      "duplicated variables should return error",
			variables: []string{"a", "a"},
			amount:    2,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConstraintInput(tt.variables, tt.amount)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
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
	}{
		{
			name:      "should create constraint",
			variables: []string{"a", "b", "c"},
			amount:    2,
			want: Constraint{
				id: "id",
				coefficients: Coefficients{
					"a": -1,
					"b": -1,
					"c": -1,
				},
				bias: Bias(-2),
			},
		},
		{
			name:      "amount equal to the number of variables should return constraint",
			variables: []string{"a", "b"},
			amount:    2,
			want: Constraint{
				id: "id",
				coefficients: Coefficients{
					"a": -1,
					"b": -1,
				},
				bias: Bias(-2),
			},
		},
		{
			name:      "no variables should return constraint",
			variables: []string{},
			amount:    0,
			want: Constraint{
				id:           "id",
				coefficients: Coefficients{},
				bias:         Bias(0),
			},
		},
		{
			name:      "nil variables should return constraint",
			variables: nil,
			amount:    0,
			want: Constraint{
				id:           "id",
				coefficients: Coefficients{},
				bias:         Bias(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAtLeastConstraint(tt.variables, tt.amount)

			assert.NoError(t, err)
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
	}{
		{
			name:      "should create constraint",
			variables: []string{"a", "b", "c"},
			amount:    2,
			want: Constraint{
				id: "id",
				coefficients: Coefficients{
					"a": 1,
					"b": 1,
					"c": 1,
				},
				bias: Bias(2),
			},
		},
		{
			name:      "amount equal to the number of variables should return constraint",
			variables: []string{"a", "b"},
			amount:    2,
			want: Constraint{
				id: "id",
				coefficients: Coefficients{
					"a": 1,
					"b": 1,
				},
				bias: Bias(2),
			},
		},
		{
			name:      "no variables should return constraint",
			variables: []string{},
			amount:    0,
			want: Constraint{
				id:           "id",
				coefficients: Coefficients{},
				bias:         Bias(0),
			},
		},
		{
			name:      "nil variables should return constraint",
			variables: nil,
			amount:    0,
			want: Constraint{
				id:           "id",
				coefficients: Coefficients{},
				bias:         Bias(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAtMostConstraint(tt.variables, tt.amount)

			assert.NoError(t, err)
			assert.Equal(t, tt.want.bias, got.bias, "Bias should match")
			assert.Equal(t, tt.want.coefficients, got.coefficients, "Coefficients should match")
		})
	}
}

func Test_newConstraintID(t *testing.T) {
	tests := []struct {
		name         string
		coefficients Coefficients
		bias         Bias
		want         string
	}{
		{
			name: "should create id",
			coefficients: Coefficients{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			bias: 1,
			want: "faba03a1732d697d527760d2c395b1ef6b842115",
		},
		{
			name: "should create id",
			coefficients: Coefficients{
				"c": 3,
				"b": 2,
				"a": 1,
			},
			bias: 1,
			want: "faba03a1732d697d527760d2c395b1ef6b842115",
		},
		{
			name: "should create id",
			coefficients: Coefficients{
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
			coefficients: Coefficients{
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
			coefficients: Coefficients{},
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
