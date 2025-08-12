package pldag

import "github.com/google/uuid"

type Model struct {
	primities  *map[string]any
	composites *map[string]Operation
}

type OperationType string

const (
	OperationAnd OperationType = "AND"
	OperationOr  OperationType = "OR"
)

type Operation struct {
	variables []string
	operation OperationType
	bias      int
}

type LinearSystem struct {
	matrix          [][]int
	rightHandVector []int
}

func New() *Model {
	return &Model{
		primities:  &map[string]any{},
		composites: &map[string]Operation{},
	}
}

func (m *Model) SetPrimities(primities ...string) {
	for _, primitive := range primities {
		(*m.primities)[primitive] = nil
	}
}

func (m *Model) SetAnd(variables ...string) string {
	id := uuid.New().String()

	bias := -1 * len(variables)
	(*m.composites)[id] = Operation{
		variables: variables,
		operation: OperationAnd,
		bias:      bias,
	}

	return id
}

func (m *Model) NewLinearSystem() LinearSystem {
	return LinearSystem{
		matrix:          [][]int{},
		rightHandVector: []int{},
	}
}
