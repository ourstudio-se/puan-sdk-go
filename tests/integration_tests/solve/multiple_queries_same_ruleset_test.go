package solve

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/puan"
)

func Test_multipleQueries_shouldNotChangeRuleset(t *testing.T) {
	ruleset, startTime := multipleQueriesWithSameRulesetSetup()
	initialShape := ruleset.Polyhedron().SparseMatrix().Shape()

	// query 1
	from := startTime.Add(45 * time.Minute)
	_, err := solutionCreator.Create(nil, ruleset, &from)
	assert.NoError(t, err)
	// passed periods should not change the initial ruleset
	// as it introduce additional constraints.
	assert.Equal(t, initialShape, ruleset.Polyhedron().SparseMatrix().Shape())

	// query 2
	selection := puan.NewSelectionBuilder("x").
		WithSubSelectionID("y").
		Build()

	_, err = solutionCreator.Create(puan.Selections{selection}, ruleset, nil)
	assert.NoError(t, err)
	// selections should not change the initial ruleset
	// as composite selections introduce additional constraints.
	assert.Equal(t, initialShape, ruleset.Polyhedron().SparseMatrix().Shape())
}

func multipleQueriesWithSameRulesetSetup() (puan.Ruleset, time.Time) {
	creator := puan.NewRulesetCreator()
	_ = creator.AddPrimitives("x", "y", "z")
	xorID, _ := creator.SetXor("y", "z")
	implyID, _ := creator.SetImply("x", xorID)
	_ = creator.Assume(implyID)

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)
	_ = creator.AssumeInPeriod(
		"x",
		startTime.Add(30*time.Minute),
		endTime)

	ruleset, _ := creator.Create()

	return ruleset, startTime
}
