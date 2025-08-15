package pldag

import (
	"crypto/sha1"
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/go-errors/errors"
)

// implies, and, or, xor, not
// hash id for each variable

type LinearSystem struct {
	aMatrix [][]int
	bVector []int
}

func (l LinearSystem) A() [][]int {
	return l.aMatrix
}

func (l LinearSystem) B() []int {
	return l.bVector
}

type coefficientValues map[string]int

func (c coefficientValues) negate() coefficientValues {
	negated := make(map[string]int, len(c))
	for key, value := range c {
		negated[key] = -value
	}

	return negated
}

func (c coefficientValues) calculateMaxAbsInnerBound() int {
	sumNegatives, sumPositives := 0, 0
	for _, value := range c {
		if value < 0 {
			sumNegatives += value
		}
		if value > 0 {
			sumPositives += value
		}
	}

	absSumNegatives := math.Abs(float64(sumNegatives))
	absSumPositives := math.Abs(float64(sumPositives))
	maxValue := math.Max(absSumNegatives, absSumPositives)

	return int(maxValue)
}

type Bias int

func (b Bias) negate() int {
	return int(-b + 1)
}

type (
	Model struct {
		variables   []string
		constraints []Constraint
	}

	Constraint struct {
		id           string
		coefficients coefficientValues
		bias         Bias
	}
)

func New() *Model {
	return &Model{
		variables:   []string{},
		constraints: []Constraint{},
	}
}

func (m *Model) SetPrimitives(primitives ...string) {
	m.variables = append(m.variables, primitives...)
}

func (m *Model) SetAnd(variables ...string) (string, error) {
	id, err := m.setAtLeast(variables, len(variables))
	if err != nil {
		return "", err
	}

	return id, nil
}

func (m *Model) GenerateSystem() LinearSystem {
	var aMatrix [][]int
	var bVector []int

	for _, c := range m.constraints {
		row, b := createSupportImpliesConstraint(c, m.variables)
		aMatrix = append(aMatrix, row)
		bVector = append(bVector, b)

		row, b = createConstraintImpliesSupport(c, m.variables)
		aMatrix = append(aMatrix, row)
		bVector = append(bVector, b)
	}

	return LinearSystem{aMatrix, bVector}
}

func (m *Model) setAtLeast(variables []string, amount int) (string, error) {
	constraint, err := newAtLeastConstraint(variables, amount)
	if err != nil {
		return "", err
	}

	m.setConstraint(constraint)

	return constraint.id, nil
}

func (m *Model) setAtMost(variables []string, amount int) (string, error) {
	constraint, err := newAtMostConstraint(variables, amount)
	if err != nil {
		return "", err
	}
	m.setConstraint(constraint)

	return constraint.id, nil
}

func newAtLeastConstraint(variables []string, amount int) (Constraint, error) {
	if amount > len(variables) {
		return Constraint{}, errors.New("amount cannot be greater than number of variables")
	}

	if amount < 0 {
		return Constraint{}, errors.New("amount cannot be negative")
	}

	coefficients := make(coefficientValues)
	for _, v := range variables {
		coefficients[v] = 1
	}

	bias := Bias(amount)

	constraint := newConstraint(coefficients, bias)

	return constraint, nil
}

func newAtMostConstraint(variables []string, amount int) (Constraint, error) {
	if amount > len(variables) {
		return Constraint{}, errors.New("amount cannot be greater than number of variables")
	}

	if amount < 0 {
		return Constraint{}, errors.New("amount cannot be negative")
	}

	coefficients := make(coefficientValues)
	for _, v := range variables {
		coefficients[v] = -1
	}

	bias := Bias(-amount)

	constraint := newConstraint(coefficients, bias)

	return constraint, nil
}

func newConstraint(coefficients coefficientValues, bias Bias) Constraint {
	id := newConstraintID(coefficients, bias)
	constraint := Constraint{
		id:           id,
		coefficients: coefficients,
		bias:         bias,
	}
	return constraint
}

func (m *Model) setConstraint(c Constraint) {
	if m.idAlreadyExists(c.id) {
		return
	}

	m.variables = append(m.variables, c.id)
	m.constraints = append(m.constraints, c)
}

func (m *Model) idAlreadyExists(id string) bool {
	return slices.Contains(m.variables, id)
}

func newConstraintID(coefficients coefficientValues, bias Bias) string {
	keys := slices.Sorted(maps.Keys(coefficients))

	h := sha1.New()
	for _, key := range keys {
		h.Write([]byte(key))
		h.Write([]byte(fmt.Sprintf("%d", coefficients[key])))
	}
	h.Write([]byte(fmt.Sprintf("%d", bias)))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func createConstraintImpliesSupport(c Constraint, variables []string) ([]int, int) {
	coefficients := c.coefficients.negate()
	innerBound := coefficients.calculateMaxAbsInnerBound()
	negatedBias := c.bias.negate()

	constraintRow := make([]int, len(variables))
	for i, v := range variables {
		if v == c.id {
			constraintRow[i] = innerBound + negatedBias
			continue
		}

		if value, ok := coefficients[v]; ok {
			constraintRow[i] = value
		} else {
			constraintRow[i] = 0
		}
	}

	return constraintRow, negatedBias
}

func createSupportImpliesConstraint(c Constraint, variables []string) ([]int, int) {
	innerBound := c.coefficients.calculateMaxAbsInnerBound()
	b := int(c.bias) - innerBound

	constraintRow := make([]int, len(variables))
	for i, v := range variables {
		if v == c.id {
			constraintRow[i] = -innerBound
			continue
		}

		if value, ok := c.coefficients[v]; ok {
			constraintRow[i] = value
		} else {
			constraintRow[i] = 0
		}
	}

	return constraintRow, b
}
