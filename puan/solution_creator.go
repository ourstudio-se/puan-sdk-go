package puan

import (
	"time"

	"github.com/go-errors/errors"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
)

type SolverClient interface {
	Solve(query *SolverQuery) (Solution, error)
	SolveWithManyWeights(query *MultiWeightSolverQuery) ([]Solution, error)
}

type SolutionCreator struct {
	singleSolutionCreator       *singleSolutionCreator
	solutionsBySelectionCreator *solutionsBySelectionCreator
	nextSolutionsCreator        *nextSolutionsCreator
	manySolutionsCreator        *manySolutionsCreator
}

func NewSolutionCreator(
	client SolverClient,
) *SolutionCreator {
	singleSolutionCreator := newSingleSolutionCreator(
		client,
	)
	solutionsBySelectionCreator := newSolutionsBySelectionCreator(
		client,
		singleSolutionCreator,
	)
	nextSolutionsCreator := newNextSolutionsCreator(
		client,
		singleSolutionCreator,
	)
	manySolutionsCreator := newManySolutionsCreator(
		client,
	)
	return &SolutionCreator{
		singleSolutionCreator:       singleSolutionCreator,
		solutionsBySelectionCreator: solutionsBySelectionCreator,
		nextSolutionsCreator:        nextSolutionsCreator,
		manySolutionsCreator:        manySolutionsCreator,
	}
}

func (c *SolutionCreator) Create(
	query SolutionQuery,
) (SolutionEnvelope, error) {
	solution, err := c.singleSolutionCreator.create(query)
	if err != nil {
		err = updateSolveError(err, query.ruleset, query.from)
		return SolutionEnvelope{}, err
	}

	return SolutionEnvelope{
		solution: solution,
	}, nil
}

func (c *SolutionCreator) CreateSolutionsBySelection(
	query SolutionQuery,
) (SolutionsBySelectionEnvelope, error) {
	solutions, err := c.solutionsBySelectionCreator.create(query)
	if err != nil {
		err = updateSolveError(err, query.ruleset, query.from)
		return SolutionsBySelectionEnvelope{}, err
	}

	return NewSolutionsBySelectionEnvelope(solutions)
}

func (c *SolutionCreator) CreateNextSolutions(
	query NextSolutionsQuery,
) (SolutionsBySelectionEnvelope, error) {
	solutions, err := c.nextSolutionsCreator.create(query)
	if err != nil {
		err = updateSolveError(err, query.ruleset, query.from)
		return SolutionsBySelectionEnvelope{}, err
	}

	return NewSolutionsBySelectionEnvelope(solutions)
}

func (c *SolutionCreator) CreateManySolutions(
	query ManySolutionsQuery,
) (SolutionsBySelectionGroupEnvelope, error) {
	solutions, err := c.manySolutionsCreator.create(query)
	if err != nil {
		err = updateSolveError(err, query.ruleset, query.from)
		return SolutionsBySelectionGroupEnvelope{}, err
	}

	envelope := NewSolutionsBySelectionGroupEnvelope(solutions)
	return envelope, nil
}

func updateSolveError(
	err error,
	ruleset Ruleset,
	from *time.Time,
) error {
	solverFailed := errors.Is(err, puanerror.SolverFailed)
	if solverFailed {
		invalidTime := !ruleset.isValidFromTime(from)
		if invalidTime {
			return errors.Errorf(
				"%w: from '%s' is not valid for the ruleset",
				puanerror.InvalidArgument,
				from,
			)
		}
	}

	return err
}
