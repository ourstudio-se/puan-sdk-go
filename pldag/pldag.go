package pldag

import "github.com/google/uuid"

type Model struct {
	variables  *[]string
	composites *map[string]Operation
}

type OperationType string

const (
	OperationAnd OperationType = "AND"
	OperationOr  OperationType = "OR"
)

type Operation struct {
	variables   map[string]any
	operation   OperationType
	bias        int
	auxiliaryID string
}

type LinearSystem struct {
	matrix          [][]int
	rightHandVector []int
}

func New() *Model {
	return &Model{
		variables:  &[]string{},
		composites: &map[string]Operation{},
	}
}

func (m *Model) SetPrimities(primities ...string) {
	for _, primitive := range primities {
		*m.variables = append(*m.variables, primitive)
	}
}

func (m *Model) SetAnd(variables ...string) string {
	id := uuid.New().String()

	auxiliaryID := uuid.New().String()
	*m.variables = append(*m.variables, auxiliaryID)

	variablesByID := map[string]any{}
	for _, variable := range variables {
		variablesByID[variable] = nil
	}
	variablesByID[auxiliaryID] = nil

	bias := -1 * len(variables)
	(*m.composites)[id] = Operation{
		variables:   variablesByID,
		operation:   OperationAnd,
		bias:        bias,
		auxiliaryID: auxiliaryID,
	}

	return id
}

func (m *Model) NewLinearSystem() LinearSystem {
	matrix := [][]int{}
	rightHandVector := []int{}

	for _, composite := range *m.composites {
		row1 := []int{}

		for _, variable := range *m.variables {
			if _, ok := composite.variables[variable]; ok {
				if variable == composite.auxiliaryID {
					row1 = append(row1, composite.bias)
				} else {
					row1 = append(row1, 1)
				}
			} else {
				row1 = append(row1, 0)
			}
		}
		matrix = append(matrix, row1)

		rightHandVectorValue1 := len(composite.variables) - 1 + composite.bias
		rightHandVector = append(rightHandVector, rightHandVectorValue1)

		row2 := []int{}
		for _, variable := range *m.variables {
			if _, ok := composite.variables[variable]; ok {
				if variable == composite.auxiliaryID {
					row2 = append(row2, 1)
				} else {
					row2 = append(row2, -1)
				}
			} else {
				row2 = append(row2, 0)
			}
		}
		matrix = append(matrix, row2)

		rightHandVectorValue2 := 1 - len(composite.variables) + 1
		rightHandVector = append(rightHandVector, rightHandVectorValue2)
	}

	return LinearSystem{
		matrix:          matrix,
		rightHandVector: rightHandVector,
	}
}
