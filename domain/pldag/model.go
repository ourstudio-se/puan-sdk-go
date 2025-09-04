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

type Bias int

func (b Bias) negate() Bias {
	return -b - 1
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

type (
	Model struct {
		variables         []string
		constraints       Constraints
		assumeConstraints AuxiliaryConstraints
	}

	Constraint struct {
		id           string
		coefficients CoefficientValues
		bias         Bias
	}

	AuxiliaryConstraint struct {
		coefficients CoefficientValues
		bias         Bias
	}

	Constraints          []Constraint
	AuxiliaryConstraints []AuxiliaryConstraint
)

func (c AuxiliaryConstraint) Coefficients() CoefficientValues {
	return c.coefficients
}

func (c AuxiliaryConstraint) Bias() Bias {
	return c.bias
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

func (m *Model) PrimitiveVariables() []string {
	constraintIDs := make([]string, len(m.constraints))
	for i := range m.constraints {
		constraintIDs[i] = m.constraints[i].id
	}

	primitiveIDs := utils.Without(m.variables, constraintIDs)

	return primitiveIDs
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

func (m *Model) Variables() []string {
	return m.variables
}

func New() *Model {
	return &Model{
		variables:         []string{},
		constraints:       Constraints{},
		assumeConstraints: AuxiliaryConstraints{},
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

func (m *Model) SetOr(variables ...string) (string, error) {
	id, err := m.setAtLeast(variables, 1)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (m *Model) SetNot(variables ...string) (string, error) {
	id, err := m.setAtMost(variables, 0)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (m *Model) SetImply(condition, consequence string) (string, error) {
	notID, err := m.SetNot(condition)
	if err != nil {
		return "", err
	}

	return m.SetOr([]string{notID, consequence}...)
}

func (m *Model) SetXor(variables ...string) (string, error) {
	atLeastID, err := m.setAtLeast(variables, 1)
	if err != nil {
		return "", err
	}
	atMostID, err := m.setAtMost(variables, 1)
	if err != nil {
		return "", err
	}

	return m.SetAnd([]string{atLeastID, atMostID}...)
}

func (m *Model) SetOneOrNone(variables ...string) (string, error) {
	return m.setAtMost(variables, 1)
}

func (m *Model) SetEquivalent(variableOne, variableTwo string) (string, error) {
	andID, err := m.SetAnd(variableOne, variableTwo)
	if err != nil {
		return "", err
	}
	notID, err := m.SetNot(variableOne, variableTwo)
	if err != nil {
		return "", err
	}

	return m.SetOr(andID, notID)
}

func (m *Model) Assume(variables ...string) error {
	err := m.validateAssumedVariables(variables...)
	if err != nil {
		return err
	}

	constraint := m.newAssumedConstraint(variables...)
	m.assumeConstraints = append(m.assumeConstraints, constraint)

	return nil
}

func (m *Model) newAssumedConstraint(variables ...string) AuxiliaryConstraint {
	coefficients := make(CoefficientValues, len(variables))
	for _, id := range variables {
		coefficients[id] = -1
	}

	bias := Bias(-len(variables))

	return newAuxiliaryConstraint(coefficients, bias)
}

func (m *Model) validateAssumedVariables(assumedVariables ...string) error {
	existingAssumedVariables := m.assumeConstraints.coefficientIDs()
	seen := make(map[string]any)
	for _, v := range assumedVariables {
		if _, ok := seen[v]; ok {
			return errors.New("duplicated variable")
		}
		seen[v] = nil

		if slices.Contains(existingAssumedVariables, v) {
			return errors.New("variable already assumed")
		}
		if !slices.Contains(m.variables, v) {
			return errors.New("variable not in model")
		}
	}

	return nil
}

func (m *Model) GeneratePolyhedron() Polyhedron {
	var aMatrix [][]int
	var bVector []int

	constraintsWithSupport := m.toAuxiliaryConstraintsWithSupport()
	var constraintsInMatrix AuxiliaryConstraints
	constraintsInMatrix = append(constraintsInMatrix, constraintsWithSupport...)
	constraintsInMatrix = append(constraintsInMatrix, m.assumeConstraints...)
	for _, c := range constraintsInMatrix {
		row := c.asMatrixRow(m.variables)
		bias := int(c.bias)
		aMatrix = append(aMatrix, row)
		bVector = append(bVector, bias)
	}

	return NewPolyhedron(aMatrix, bVector)
}

func (m *Model) toAuxiliaryConstraintsWithSupport() AuxiliaryConstraints {
	var constraints AuxiliaryConstraints
	for _, c := range m.constraints {
		supportImpliesConstraint, constraintImpliesSupport := c.ToAuxiliaryConstraintsWithSupport()
		constraints = append(constraints, supportImpliesConstraint)
		constraints = append(constraints, constraintImpliesSupport)
	}

	return constraints
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

func (m *Model) setAtLeast(variables []string, amount int) (string, error) {
	constraint, err := NewAtLeastConstraint(variables, amount)
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

func newAtMostConstraint(variables []string, amount int) (Constraint, error) {
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

func newConstraint(coefficients CoefficientValues, bias Bias) Constraint {
	id := newConstraintID(coefficients, bias)
	constraint := Constraint{
		id:           id,
		coefficients: coefficients,
		bias:         bias,
	}
	return constraint
}

func newAuxiliaryConstraint(coefficients CoefficientValues, bias Bias) AuxiliaryConstraint {
	constraint := AuxiliaryConstraint{
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

func newConstraintID(coefficients CoefficientValues, bias Bias) string {
	keys := slices.Sorted(maps.Keys(coefficients))

	h := sha1.New()
	for _, key := range keys {
		h.Write([]byte(key))
		fmt.Fprintf(h, "%d", coefficients[key])
	}
	fmt.Fprintf(h, "%d", bias)

	return fmt.Sprintf("%x", h.Sum(nil))
}
