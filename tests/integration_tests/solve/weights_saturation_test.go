package solve

import (
	"testing"

	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

func Test_WeightsTooLarge_givenManySelections_weightsShouldBeTooLarge(t *testing.T) {
	ruleset, primitives := rulesetWithPrimitivesForSaturationTests()

	selections := puan.Selections{}
	for _, primitive := range primitives {
		selections = append(
			selections,
			puan.NewSelectionBuilder(primitive).Build(),
		)
	}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	weightsTooLarge := envelope.WeightsTooLarge()
	assert.True(
		t,
		weightsTooLarge,
	)
}

func Test_WeightsTooLarge_givenFewSelections_weightsShouldNotBeTooLarge(t *testing.T) {
	ruleset, primitives := rulesetWithPrimitivesForSaturationTests()

	selections := puan.Selections{}
	for _, primitive := range primitives[:10] {
		selections = append(
			selections,
			puan.NewSelectionBuilder(primitive).Build(),
		)
	}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	weightsTooLarge := envelope.WeightsTooLarge()
	assert.False(
		t,
		weightsTooLarge,
	)
}

func rulesetWithPrimitivesForSaturationTests() (puan.Ruleset, []string) {
	creator := puan.NewRulesetCreator()

	primitives := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 100
			oo.RandomMaxSliceSize = 100
		},
	)
	_ = creator.AddPrimitives(primitives...)

	// Needed to set some rule to avoid having all variables as independent
	andID, _ := creator.SetAnd(primitives...)
	_ = creator.Assume(andID)

	ruleset, _ := creator.Create()

	return ruleset, primitives
}
