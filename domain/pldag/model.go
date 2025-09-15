package pldag

import (
	"slices"

	"github.com/go-errors/errors"

	"github.com/ourstudio-se/puan-sdk-go/utils"
)

type Model struct {
	variables         []string
	constraints       Constraints
	assumeConstraints AuxiliaryConstraints
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
	dedupedVariables := utils.Dedupe(variables)
	id, err := m.setAtLeast(dedupedVariables, len(dedupedVariables))
	if err != nil {
		return "", err
	}

	return id, nil
}

func (m *Model) SetOr(variables ...string) (string, error) {
	dedupedVariables := utils.Dedupe(variables)
	id, err := m.setAtLeast(dedupedVariables, 1)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (m *Model) SetNot(variables ...string) (string, error) {
	dedupedVariables := utils.Dedupe(variables)
	id, err := m.setAtMost(dedupedVariables, 0)
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
	dedupedVariables := utils.Dedupe(variables)
	return m.setAtMost(dedupedVariables, 1)
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
	dedupedVariables := utils.Dedupe(variables)
	err := m.validateAssumedVariables(dedupedVariables...)
	if err != nil {
		return err
	}

	constraints := m.newAssumedConstraints(dedupedVariables...)
	m.assumeConstraints = append(m.assumeConstraints, constraints...)

	return nil
}

func (m *Model) NewPolyhedron() *Polyhedron {
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

func (m *Model) PrimitiveVariables() []string {
	constraintIDs := make([]string, len(m.constraints))
	for i := range m.constraints {
		constraintIDs[i] = m.constraints[i].id
	}

	primitiveIDs := utils.Without(m.variables, constraintIDs)

	return primitiveIDs
}

func (m *Model) Variables() []string {
	return m.variables
}

func (m *Model) newAssumedConstraints(variables ...string) AuxiliaryConstraints {
	negativeCoefficient := m.newNegativeAssumedConstraint(variables...)
	positiveCoefficient := m.newPositiveAssumedConstraint(variables...)

	return AuxiliaryConstraints{negativeCoefficient, positiveCoefficient}
}

func (m *Model) newNegativeAssumedConstraint(variables ...string) AuxiliaryConstraint {
	coefficients := make(Coefficients, len(variables))
	for _, id := range variables {
		coefficients[id] = -1
	}

	bias := Bias(-len(variables))
	constraint := newAuxiliaryConstraint(coefficients, bias)

	return constraint
}

func (m *Model) newPositiveAssumedConstraint(variables ...string) AuxiliaryConstraint {
	coefficients := make(Coefficients, len(variables))
	for _, id := range variables {
		coefficients[id] = 1
	}

	bias := Bias(len(variables))
	constraint := newAuxiliaryConstraint(coefficients, bias)

	return constraint
}

func (m *Model) validateAssumedVariables(assumedVariables ...string) error {
	existingAssumedVariables := m.assumeConstraints.coefficientIDs()
	for _, v := range assumedVariables {
		if slices.Contains(existingAssumedVariables, v) {
			return errors.New("variable already assumed")
		}
		if !slices.Contains(m.variables, v) {
			return errors.New("variable not in model")
		}
	}

	return nil
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

func (m *Model) setAtLeast(variables []string, amount int) (string, error) {
	constraint, err := NewAtLeastConstraint(variables, amount)
	if err != nil {
		return "", err
	}

	m.setConstraint(constraint)

	return constraint.id, nil
}

func (m *Model) setAtMost(variables []string, amount int) (string, error) {
	constraint, err := NewAtMostConstraint(variables, amount)
	if err != nil {
		return "", err
	}
	m.setConstraint(constraint)

	return constraint.id, nil
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
