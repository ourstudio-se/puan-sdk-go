//nolint:lll
package solve

import (
	"testing"

	"github.com/ourstudio-se/puan-sdk-go/puan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateNextSolutions_shouldCreateSolutionsForAllPrimitives(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

	primitives := []string{"optA", "optB", "optC", "optD", "optE", "optF"}
	_ = creator.AddPrimitives(primitives...)

	orID, _ := creator.SetOr("optA", "optB")

	_ = creator.Assume(orID)

	cImpliesD, _ := creator.SetImply("optC", "optD")

	_ = creator.Assume(cImpliesD)

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("optA").Build(),
		puan.NewSelectionBuilder("optE").Build(),
		puan.NewSelectionBuilder("optC").Build(),
	}
	query := puan.NewSolutionQueryBuilder().
		WithSelections(selections).
		WithRuleset(ruleset).
		Build()
	envelope, _ := solutionCreator.CreateNextSolutions(query)

	assert.Len(t, envelope.SolutionsBySelection(), len(primitives))

	_, err := envelope.GetSolutionBySelection(
		puan.NewSelectionBuilder("optA").
			WithAction(puan.REMOVE).
			Build(),
	)
	require.NoError(t, err)
	_, err = envelope.GetSolutionBySelection(
		puan.NewSelectionBuilder("optB").
			Build(),
	)
	require.NoError(t, err)
	_, err = envelope.GetSolutionBySelection(
		puan.NewSelectionBuilder("optC").
			WithAction(puan.REMOVE).
			Build(),
	)
	require.NoError(t, err)
	_, err = envelope.GetSolutionBySelection(
		puan.NewSelectionBuilder("optD").
			Build(),
	)
	require.NoError(t, err)
	_, err = envelope.GetSolutionBySelection(
		puan.NewSelectionBuilder("optE").
			WithAction(puan.REMOVE).
			Build(),
	)
	require.NoError(t, err)
	_, err = envelope.GetSolutionBySelection(
		puan.NewSelectionBuilder("optF").
			Build(),
	)
	require.NoError(t, err)
}

func Test_CreateNextSolutions_shouldRemoveExistingSelections(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

	primitives := []string{"optA", "optB", "optC", "optD", "optE", "optF"}
	_ = creator.AddPrimitives(primitives...)

	orID, _ := creator.SetOr("optA", "optB")

	_ = creator.Assume(orID)

	cImpliesD, _ := creator.SetImply("optC", "optD")

	_ = creator.Assume(cImpliesD)

	ruleset, err := creator.Create()
	assert.NoError(t, err)

	selections := puan.Selections{
		puan.NewSelectionBuilder("optA").Build(),
		puan.NewSelectionBuilder("optE").Build(),
		puan.NewSelectionBuilder("optC").Build(),
	}
	query := puan.NewSolutionQueryBuilder().
		WithSelections(selections).
		WithRuleset(ruleset).
		Build()
	envelope, err := solutionCreator.CreateNextSolutions(query)
	require.NoError(t, err)

	removeA := puan.NewSelectionBuilder("optA").WithAction(puan.REMOVE).Build()
	solution, err := envelope.GetSolutionBySelection(removeA)
	require.NoError(t, err)

	asserter := newSolutionAsserter(solution.Solution())
	asserter.assertInactive(t, "optA")
	asserter.assertActive(t, "optB")
	asserter.assertActive(t, "optC")
	asserter.assertActive(t, "optD")
	asserter.assertActive(t, "optE")
	asserter.assertInactive(t, "optF")
}

func Test_CreateNextSolutions_shouldAddNewSelections(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

	primitives := []string{"optA", "optB", "optC", "optD", "optE", "optF"}
	_ = creator.AddPrimitives(primitives...)

	orID, _ := creator.SetOr("optA", "optB")

	_ = creator.Assume(orID)

	cImpliesD, _ := creator.SetImply("optC", "optD")

	_ = creator.Assume(cImpliesD)

	ruleset, err := creator.Create()
	assert.NoError(t, err)

	selections := puan.Selections{
		puan.NewSelectionBuilder("optA").Build(),
		puan.NewSelectionBuilder("optE").Build(),
		puan.NewSelectionBuilder("optC").Build(),
	}
	query := puan.NewSolutionQueryBuilder().
		WithSelections(selections).
		WithRuleset(ruleset).
		Build()
	envelope, err := solutionCreator.CreateNextSolutions(query)
	require.NoError(t, err)

	addB := puan.NewSelectionBuilder("optB").Build()
	solution, err := envelope.GetSolutionBySelection(addB)
	require.NoError(t, err)

	asserter := newSolutionAsserter(solution.Solution())
	asserter.assertActive(t, "optA")
	asserter.assertActive(t, "optB")
	asserter.assertActive(t, "optC")
	asserter.assertActive(t, "optD")
	asserter.assertActive(t, "optE")
	asserter.assertInactive(t, "optF")
}
