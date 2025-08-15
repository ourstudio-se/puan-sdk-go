package pldag

import (
	"reflect"
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
			name: "nil value",
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
		want int
	}{
		{
			name: "should negate bias",
			b:    1,
			want: 0,
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
			assert.Equalf(t, tt.want, newConstraintID(tt.coefficients, tt.bias), "newConstraintID(%v, %v)", tt.coefficients, tt.bias)
		})
	}
}
