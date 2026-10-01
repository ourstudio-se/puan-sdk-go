package puan

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type nextSolutionsCreatorSuite struct {
	suite.Suite

	ruleset    Ruleset
	primitives []string
	client     *mockSolverClient
	creator    *nextSolutionsCreator
}

func Test_SolutionCreator_Suite(t *testing.T) {
	suite.Run(t, new(nextSolutionsCreatorSuite))
}

// SetupSuite builds a saturatable ruleset once, as it is shared and never
// mutated by the tests.
func (s *nextSolutionsCreatorSuite) SetupSuite() {
	primitives := make([]string, 100)
	for i := range primitives {
		primitives[i] = fmt.Sprintf("p%d", i)
	}

	creator := NewRulesetCreator()
	_ = creator.AddPrimitives(primitives...)

	orID, err := creator.SetOr(primitives...)
	s.Require().NoError(err)
	s.Require().NoError(creator.Assume(orID))

	ruleset, err := creator.Create()
	s.Require().NoError(err)

	s.ruleset = ruleset
	s.primitives = primitives

	s.client = &mockSolverClient{}
	queryCreator := newSolverQueryCreator()
	singleSolutionCreator := newSingleSolutionCreator(s.client, queryCreator)
	s.creator = newNextSolutionsCreator(s.client, queryCreator, singleSolutionCreator)
}

func (s *nextSolutionsCreatorSuite) TearDownTest() {
	s.client.multiSolveCalls = 0
	s.client.solveCalls = 0
}

func (
	s *nextSolutionsCreatorSuite,
) Test_calculateNextDependentSolutions_givenSmallWeights_shouldSolveInOneBatch() {
	currentSelections := selectionsFor(s.T(), s.primitives[:3])
	nextSelections := selectionsFor(s.T(), s.primitives[3:5])
	query, err := NewNextSolutionsQuery(
		currentSelections,
		nextSelections,
		s.ruleset,
		nil,
		nil,
	)
	s.Require().NoError(err)

	actual, err := s.creator.calculateNextDependentSolutions(query)

	s.Require().NoError(err)
	assertSolutionExistsForEachSelection(s.T(), actual, nextSelections)
	s.Assert().Equal(1, s.client.multiSolveCalls)
	s.Assert().Zero(s.client.solveCalls)
}

func (
	s *nextSolutionsCreatorSuite,
) Test_calculateNextDependentSolutions_givenOversizedWeights_shouldSolveOneByOne() {
	// Select many for current to enforce saturation.
	currentSelections := selectionsFor(s.T(), s.primitives[:25])
	nextSelections := selectionsFor(s.T(), s.primitives[25:27])
	query, err := NewNextSolutionsQuery(
		currentSelections,
		nextSelections,
		s.ruleset,
		nil,
		nil,
	)
	s.Require().NoError(err)

	actual, err := s.creator.calculateNextDependentSolutions(query)

	s.Require().NoError(err)
	assertSolutionExistsForEachSelection(s.T(), actual, nextSelections)
	s.Assert().Zero(s.client.multiSolveCalls)
	// each selection is split once, 2 calls per next selection, therefore 4 in total.
	s.Assert().Equal(4, s.client.solveCalls)
}

func (
	s *nextSolutionsCreatorSuite,
) Test_CreateNextSolutions_givenNoNextSelections_shouldReturnEmptyEnvelope() {
	query, err := NewNextSolutionsQuery(
		selectionsFor(s.T(), s.primitives[:3]),
		nil,
		s.ruleset,
		nil,
		nil,
	)
	s.Require().NoError(err)

	solutions, err := s.creator.createNextSolutions(query)

	s.Require().NoError(err)
	s.Assert().Empty(solutions)
}

func selectionsFor(t *testing.T, primitives []string) Selections {
	t.Helper()
	selections := make(Selections, len(primitives))
	for i, primitive := range primitives {
		selections[i] = NewSelectionBuilder(primitive).Build()
	}

	return selections
}

func assertSolutionExistsForEachSelection(
	t *testing.T,
	solutions []SolutionBySelection,
	selections Selections,
) {
	t.Helper()

	require.Len(t, solutions, len(selections))

	envelope, err := NewSolutionsBySelectionEnvelope(solutions)
	require.NoError(t, err)

	for _, selection := range selections {
		_, err := envelope.GetSolutionBySelection(selection)
		assert.NoError(t, err, "no solution for selection %s", selection.ID())
	}
}

type mockSolverClient struct {
	solveCalls      int
	multiSolveCalls int
}

func (c *mockSolverClient) Solve(_ *SolverQuery) (Solution, error) {
	c.solveCalls++

	return Solution{}, nil
}

func (c *mockSolverClient) SolveWithManyWeights(
	query *MultiWeightSolverQuery,
) ([]Solution, error) {
	c.multiSolveCalls++

	solutions := make([]Solution, len(query.WeightGroups()))
	for i := range solutions {
		solutions[i] = Solution{}
	}

	return solutions, nil
}
