package puan

import (
	"testing"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NextSolutionsQuery_hasEmptyNextSelections_givenEmpty_shouldReturnTrue(
	t *testing.T,
) {
	query := NextSolutionsQuery{}
	assert.True(t, query.hasEmptyNextSelections())
}

func Test_NextSolutionsQuery_hasEmptyNextSelections_givenSelections_shouldReturnFalse(
	t *testing.T,
) {
	query := NextSolutionsQuery{
		nextSelections: fake.New[Selections](),
	}
	assert.False(t, query.hasEmptyNextSelections())
}

func Test_NextSolutionsQuery_batchableAndNonBatchableQuery(t *testing.T) {
	creator := NewRulesetCreator()
	require.NoError(t, creator.AddPrimitives("a", "b", "c"))

	ruleset, err := creator.Create()
	require.NoError(t, err)

	a := NewSelectionBuilder("a").Build()
	b := NewSelectionBuilder("b").Build()
	c := NewSelectionBuilder("c").Build()

	query, err := NewNextSolutionsQuery(nil, Selections{a, b, c}, ruleset, nil, nil)
	require.NoError(t, err)

	withinLimit := weights.Weights{"underLimit": weights.WEIGHTS_SATURATION_LIMIT - 1}
	atLimit := weights.Weights{"atLimit": weights.WEIGHTS_SATURATION_LIMIT}
	tooLarge := weights.Weights{"tooLarge": weights.WEIGHTS_SATURATION_LIMIT + 1}

	theories := []struct {
		name             string
		weightGroups     []weights.Weights
		wantBatchable    Selections
		wantNonBatchable Selections
	}{
		{
			name:          "all within the limit should be batchable",
			weightGroups:  []weights.Weights{withinLimit, withinLimit, withinLimit},
			wantBatchable: Selections{a, b, c},
		},
		{
			name:             "all above the limit should be non-batchable",
			weightGroups:     []weights.Weights{tooLarge, tooLarge, tooLarge},
			wantNonBatchable: Selections{a, b, c},
		},
		{
			name:             "mixed",
			weightGroups:     []weights.Weights{withinLimit, tooLarge, withinLimit},
			wantBatchable:    Selections{a, c},
			wantNonBatchable: Selections{b},
		},
		{
			name:          "exactly at the limit should be batchable",
			weightGroups:  []weights.Weights{atLimit, atLimit, atLimit},
			wantBatchable: Selections{a, b, c},
		},
		{
			name:         "no weight groups should give empty queries",
			weightGroups: nil,
		},
	}

	for _, theory := range theories {
		t.Run(theory.name, func(t *testing.T) {
			batchable, err := query.batchableQuery(theory.weightGroups)
			require.NoError(t, err)

			nonBatchable, err := query.nonBatchableQuery(theory.weightGroups)
			require.NoError(t, err)

			assert.Equal(t, theory.wantBatchable, batchable.nextSelections)
			assert.Equal(t, theory.wantNonBatchable, nonBatchable.nextSelections)
		})
	}
}
