package pldag

import (
	"crypto/sha1"
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/utils"
)

type Constraint struct {
	id           string
	coefficients CoefficientValues
	bias         Bias
}
type Constraints []Constraint

func NewAtLeastConstraint(variables []string, amount int) (Constraint, error) {
	if utils.ContainsDuplicates(variables) {
		return Constraint{}, errors.New("duplicated variables")
	}

	if amount > len(variables) {
		return Constraint{}, errors.New("amount cannot be greater than number of variables")
	}

	if amount < 0 {
		return Constraint{}, errors.New("amount cannot be negative")
	}

	coefficients := make(CoefficientValues)
	for _, v := range variables {
		coefficients[v] = -1
	}

	bias := Bias(-amount)

	constraint := newConstraint(coefficients, bias)

	return constraint, nil
}

func NewAtMostConstraint(variables []string, amount int) (Constraint, error) {
	if utils.ContainsDuplicates(variables) {
		return Constraint{}, errors.New("duplicated variables")
	}

	if amount > len(variables) {
		return Constraint{}, errors.New("amount cannot be greater than number of variables")
	}

	if amount < 0 {
		return Constraint{}, errors.New("amount cannot be negative")
	}

	coefficients := make(CoefficientValues)
	for _, v := range variables {
		coefficients[v] = 1
	}

	bias := Bias(amount)

	constraint := newConstraint(coefficients, bias)

	return constraint, nil
}

func newConstraint(coefficients CoefficientValues, bias Bias) Constraint {
	id := newConstraintID(coefficients, bias)
	constraint := Constraint{
		id:           id,
		coefficients: coefficients,
		bias:         bias,
	}
	return constraint
}

func (c Constraint) ID() string {
	return c.id
}

func (c Constraint) Bias() Bias {
	return c.bias
}

func (c Constraint) Coefficients() CoefficientValues {
	return c.coefficients
}

func (c Constraint) ToAuxiliaryConstraintsWithSupport() (AuxiliaryConstraint, AuxiliaryConstraint) {
	supportImpliesConstraint := c.newSupportImpliesConstraint()
	constraintImpliesSupport := c.newConstraintImpliesSupport()

	return supportImpliesConstraint, constraintImpliesSupport
}

func (c Constraint) newConstraintImpliesSupport() AuxiliaryConstraint {
	negatedCoefficients := c.coefficients.negate()
	innerBound := negatedCoefficients.calculateMaxAbsInnerBound()
	negatedBias := c.bias.negate()

	newCoefficients := make(CoefficientValues, len(c.coefficients)+1)
	for coefficientID, value := range negatedCoefficients {
		newCoefficients[coefficientID] = value
	}

	newCoefficients[c.id] = int(negatedBias) - innerBound

	return AuxiliaryConstraint{
		coefficients: newCoefficients,
		bias:         negatedBias,
	}
}

func (c Constraint) newSupportImpliesConstraint() AuxiliaryConstraint {
	innerBound := c.coefficients.calculateMaxAbsInnerBound()
	bias := Bias(int(c.bias) + innerBound)

	newCoefficients := make(CoefficientValues, len(c.coefficients)+1)
	for coefficientID, value := range c.coefficients {
		newCoefficients[coefficientID] = value
	}

	newCoefficients[c.id] = innerBound

	return AuxiliaryConstraint{
		coefficients: newCoefficients,
		bias:         bias,
	}
}

type AuxiliaryConstraint struct {
	coefficients CoefficientValues
	bias         Bias
}
type AuxiliaryConstraints []AuxiliaryConstraint

func NewAuxiliaryConstraint(coefficients CoefficientValues, bias Bias) AuxiliaryConstraint {
	constraint := AuxiliaryConstraint{
		coefficients: coefficients,
		bias:         bias,
	}
	return constraint
}

func (c AuxiliaryConstraint) Coefficients() CoefficientValues {
	return c.coefficients
}

func (c AuxiliaryConstraint) Bias() Bias {
	return c.bias
}

func (c AuxiliaryConstraints) coefficientIDs() []string {
	idMap := make(map[string]any)
	for _, constraint := range c {
		for coefficientID := range constraint.coefficients {
			idMap[coefficientID] = nil
		}
	}

	ids := make([]string, len(idMap))
	for id := range idMap {
		ids = append(ids, id)
	}

	return ids
}

func (c AuxiliaryConstraint) asMatrixRow(variables []string) []int {
	row := make([]int, len(variables))
	for i, id := range variables {
		if value, ok := c.coefficients[id]; ok {
			row[i] = value
		} else {
			row[i] = 0
		}
	}

	return row
}

func newConstraintID(coefficients CoefficientValues, bias Bias) string {
	keys := slices.Sorted(maps.Keys(coefficients))

	h := sha1.New()
	for _, key := range keys {
		h.Write([]byte(key))
		_, _ = fmt.Fprintf(h, "%d", coefficients[key])
	}
	_, _ = fmt.Fprintf(h, "%d", bias)

	return fmt.Sprintf("%x", h.Sum(nil))
}

type CoefficientValues map[string]int

func (c CoefficientValues) negate() CoefficientValues {
	negated := make(map[string]int, len(c))
	for key, value := range c {
		negated[key] = -value
	}

	return negated
}

func (c CoefficientValues) calculateMaxAbsInnerBound() int {
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

func (b Bias) negate() Bias {
	return -b - 1
}
