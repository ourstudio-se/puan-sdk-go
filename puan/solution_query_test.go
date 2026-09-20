package puan

import (
	"testing"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
)

func Test_nextSolutionQueryPartitioner_givenMixedWeightGroups_shouldSplit(
	t *testing.T,
) {
	polyhedron := pldag.Polyhedron{}
	addX := NewSelectionBuilder("x").Build()
	addY := NewSelectionBuilder("y").Build()
	partitioner := nextSolutionQueryPartitioner{
		initialQuery: NextSolutionsQuery{
			ruleset: Ruleset{
				polyhedron:          &polyhedron,
				selectableVariables: []string{"x", "y"},
			},
			nextSelections: Selections{
				addX,
				addY,
			},
		},
		weightGroups: []weights.Weights{
			{"addX": 1},
			{"addY": weights.WEIGHTS_SATURATION_LIMIT + 1},
		},
	}

	batchable, err := partitioner.combinable()
	require.NoError(t, err)
	require.Len(t, batchable.nextSelections, 1)
	assert.Equal(
		t,
		batchable.nextSelections[0],
		addX,
	)

	saturated, err := partitioner.oversized()
	require.NoError(t, err)
	require.Len(t, saturated.nextSelections, 1)
	assert.Equal(
		t,
		saturated.nextSelections[0],
		addY,
	)
}

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
