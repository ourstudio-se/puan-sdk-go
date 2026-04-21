package puan

import (
	"math/rand"
	"testing"
	"time"

	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
)

func Test_RulesetCreator_newTimeBoundVariable_givenTimeEnabled_andValidPeriod(t *testing.T) {
	id := uuid.New().String()
	from := newTestTime("2024-01-01T00:00:00Z")
	to := newTestTime("2024-01-31T00:00:00Z")

	creator := NewRulesetCreator()
	err := creator.EnableTime(from, to)
	assert.NoError(t, err)

	variable, err := creator.newTimeBoundVariable(
		id,
		from,
		to,
	)

	assert.NoError(t, err)
	assert.Equal(t, id, variable.variable)
	assert.Equal(t, from, variable.period.From())
	assert.Equal(t, to, variable.period.To())
}

func Test_RulesetCreator_newTimeBoundVariable_givenTimeNotEnabled_shouldReturnError(
	t *testing.T,
) {
	id := uuid.New().String()
	from := newTestTime("2024-01-01T00:00:00Z")
	to := newTestTime("2024-01-31T00:00:00Z")

	creator := NewRulesetCreator()
	_, err := creator.newTimeBoundVariable(id, from, to)
	assert.Error(t, err)
}

func Test_RulesetCreator_newTimeBoundVariable_givenTimeEnabled_andInvalidPeriod_shouldReturnError(
	t *testing.T,
) {
	id := uuid.New().String()
	from := newTestTime("2024-01-31T00:00:00Z")
	to := newTestTime("2024-01-01T00:00:00Z")

	creator := NewRulesetCreator()
	err := creator.EnableTime(
		newTestTime("2024-01-01T00:00:00Z"),
		newTestTime("2024-02-01T00:00:00Z"),
	)
	assert.NoError(t, err)

	_, err = creator.newTimeBoundVariable(id, from, to)
	assert.Error(t, err)
}

// nolint:lll
func Test_RulesetCreator_newTimeBoundVariable_givenTimeEnabled_andAssumedPeriodOutsideOfEnabledPeriod_shouldReturnError(
	t *testing.T,
) {
	id := uuid.New().String()
	from := newTestTime("2024-01-01T00:00:00Z")
	to := newTestTime("2024-01-31T00:00:00Z")

	creator := NewRulesetCreator()
	err := creator.EnableTime(
		from.Add(1*24*time.Hour),
		to.Add(-1*24*time.Hour),
	)
	assert.NoError(t, err)

	_, err = creator.newTimeBoundVariable(id, from, to)
	assert.Error(t, err)
}

// Test_Create_givenDifferentModelingOrder_shouldReturnSamePolyhedron
// This test ensures that the order in which primitives and rules
// are added does not affect the resulting polyhedron
func Test_Create_givenDifferentModelingOrder_shouldReturnSamePolyhedron(
	t *testing.T,
) {
	primitives := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 10
			oo.RandomMaxSliceSize = 15
		},
	)

	creatorOne := NewRulesetCreator()
	_ = creatorOne.AddPrimitives(primitives...)

	id1One, _ := creatorOne.SetImply(primitives[0], primitives[1])
	id2One, _ := creatorOne.SetOr(primitives[2], primitives[3])
	id3One, _ := creatorOne.SetAnd(primitives[4], primitives[5])
	id4One, _ := creatorOne.SetXor(primitives[6], primitives[7], primitives[8])
	id5One, _ := creatorOne.SetImply(id3One, id2One)
	id6One, _ := creatorOne.SetNot(primitives[9])

	rootOne, _ := creatorOne.SetAnd(id1One, id2One, id3One, id4One, id5One, id6One)
	_ = creatorOne.Assume(rootOne)
	rulesetOne, _ := creatorOne.Create()

	shuffledPrimitives := append([]string(nil), primitives...)
	rand.Shuffle(len(shuffledPrimitives), func(i, j int) {
		shuffledPrimitives[i], shuffledPrimitives[j] = shuffledPrimitives[j], shuffledPrimitives[i]
	})

	// Create a second creator with the same primitives
	// and rules but in a different order.
	creatorTwo := NewRulesetCreator()
	_ = creatorTwo.AddPrimitives(shuffledPrimitives...)
	id1Two, _ := creatorTwo.SetAnd(primitives[4], primitives[5])
	id2Two, _ := creatorTwo.SetImply(primitives[0], primitives[1])
	id3Two, _ := creatorTwo.SetOr(primitives[2], primitives[3])
	id4Two, _ := creatorTwo.SetNot(primitives[9])
	id5Two, _ := creatorTwo.SetXor(primitives[6], primitives[7], primitives[8])
	id6Two, _ := creatorTwo.SetImply(id1Two, id3Two)

	rootTwo, _ := creatorTwo.SetAnd(id1Two, id2Two, id3Two, id4Two, id5Two, id6Two)
	_ = creatorTwo.Assume(rootTwo)
	rulesetTwo, _ := creatorTwo.Create()

	assert.Equalf(
		t,
		rulesetOne.dependentVariables,
		rulesetTwo.dependentVariables,
		"dependentVariables are not equal",
	)
	assert.Equalf(t, rulesetOne.polyhedron.A(), rulesetTwo.polyhedron.A(), "A matrices are not equal")
	assert.Equalf(t, rulesetOne.polyhedron.B(), rulesetTwo.polyhedron.B(), "B vectors are not equal")
}

func Test_RulesetCreator_AssumeInPeriod_givenSamePeriod_shouldUseAssume(t *testing.T) {
	creator := NewRulesetCreator()
	from, to := newTestTime("2024-01-01T00:00:00Z"), newTestTime("2024-01-31T23:59:59Z")
	err := creator.EnableTime(
		from,
		to,
	)
	assert.NoError(t, err)

	_ = creator.AddPrimitives("itemX")
	err = creator.AssumeInPeriod("itemX", from, to)
	assert.NoError(t, err)

	assert.Contains(t, creator.assumedVariables, "itemX")
	assert.NotContains(t, creator.timeBoundAssumedVariables.ids(), "itemX")
}

// nolint:lll
func Test_RulesetCreator_AssumeInPeriod_givenDifferentPeriod_shouldAddTimeBoundVariable(t *testing.T) {
	creator := NewRulesetCreator()
	from, to := newTestTime("2024-01-01T00:00:00Z"), newTestTime("2024-01-31T23:59:59Z")
	err := creator.EnableTime(
		from,
		to,
	)
	assert.NoError(t, err)

	newFrom := from.Add(time.Hour)
	_ = creator.AddPrimitives("itemX")
	err = creator.AssumeInPeriod("itemX", newFrom, to)
	assert.NoError(t, err)

	assert.NotContains(t, creator.assumedVariables, "itemX")
	assert.Contains(t, creator.timeBoundAssumedVariables.ids(), "itemX")
}

func Test_RulesetCreator_ForbidPeriod_givenValidPeriod_shouldAddPeriod(t *testing.T) {
	creator := RulesetCreator{
		period: &Period{
			from: newTestTime("2024-01-01"),
			to:   newTestTime("2024-01-31"),
		},
	}

	forbiddenFrom := newTestTime("2024-01-10")
	forbiddenTo := newTestTime("2024-01-15")

	err := creator.ForbidPeriod(forbiddenFrom, forbiddenTo)

	assert.NoError(t, err)
	assert.Equal(t, []Period{
		{
			from: forbiddenFrom,
			to:   forbiddenTo,
		},
	}, creator.forbiddenPeriods)
}

func Test_RulesetCreator_validateForbiddenPeriod_givenErrorCases_shouldReturnError(
	t *testing.T,
) {
	type testCase struct {
		name             string
		from             time.Time
		to               time.Time
		forbiddenPeriods []Period
	}

	cases := []testCase{
		{
			name: "outside of enabled period",
			from: newTestTime("2023-12-20"),
			to:   newTestTime("2023-12-25"),
		},
		{
			name: "overlaps with existing forbidden period",
			from: newTestTime("2024-01-14"),
			to:   newTestTime("2024-01-20"),
			forbiddenPeriods: []Period{
				{
					from: newTestTime("2024-01-10"),
					to:   newTestTime("2024-01-15"),
				},
			},
		},
		{
			name: "same as enabled period",
			from: newTestTime("2024-01-01"),
			to:   newTestTime("2024-01-31"),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			creator := RulesetCreator{
				period: &Period{
					from: newTestTime("2024-01-01"),
					to:   newTestTime("2024-01-31"),
				},
				forbiddenPeriods: tt.forbiddenPeriods,
			}

			period := Period{
				from: tt.from,
				to:   tt.to,
			}
			err := creator.validateForbiddenPeriod(period)

			assert.Error(t, err)
		})
	}
}

func Test_RulesetCreator_ForbidPeriod_givenTimeNotEnabled_shouldReturnError(
	t *testing.T,
) {
	creator := NewRulesetCreator()
	err := creator.ForbidPeriod(
		fake.New[time.Time](),
		fake.New[time.Time](),
	)
	assert.ErrorIs(t, err, puanerror.InvalidOperation)
}

func Test_RulesetCreator_ForbidPeriod_givenFromAfterTo_shouldReturnError(
	t *testing.T,
) {
	creator := RulesetCreator{
		period: &Period{
			from: newTestTime("2024-01-01"),
			to:   newTestTime("2024-01-31"),
		},
	}
	from := newTestTime("2024-01-10")
	to := newTestTime("2024-01-05")

	err := creator.ForbidPeriod(from, to)

	assert.ErrorIs(t, err, puanerror.InvalidArgument)
}

func Test_RulesetCreator_calculateAllowedPartitionedPeriods(
	t *testing.T,
) {
	allowed := Period{
		from: newTestTime("2024-01-01"),
		to:   newTestTime("2024-01-31"),
	}
	forbidden := Period{
		from: newTestTime("2024-01-10"),
		to:   newTestTime("2024-01-20"),
	}
	creator := NewRulesetCreator()
	_ = creator.EnableTime(
		allowed.From(),
		allowed.To(),
	)
	_ = creator.ForbidPeriod(
		forbidden.From(),
		forbidden.To(),
	)

	got := creator.calculateAllowedPartitionedPeriods()

	want := []Period{
		{
			from: allowed.From(),
			to:   forbidden.From(),
		},
		{
			from: forbidden.To(),
			to:   allowed.To(),
		},
	}
	assert.Equal(t, want, got)
}

func Test_RulesetCreator_setSingleOrOR_givenNoIDs_shouldReturnError(t *testing.T) {
	creator := NewRulesetCreator()
	_, err := creator.setSingleOrOR([]string{}...)

	assert.ErrorIs(t, err, puanerror.InvalidArgument)
}

func Test_RulesetCreator_setSingleOrOR_givenDuplicatedIDs_shouldReturnID(t *testing.T) {
	code := fake.New[string]()
	creator := NewRulesetCreator()
	got, err := creator.setSingleOrOR(code, code)

	assert.NoError(t, err)
	assert.Equal(t, code, got)
}

func Test_RulesetCreator_setSingleOrXOR_givenNoIDs_shouldReturnError(t *testing.T) {
	creator := NewRulesetCreator()
	_, err := creator.setSingleOrXOR([]string{}...)

	assert.ErrorIs(t, err, puanerror.InvalidArgument)
}

func Test_RulesetCreator_setSingleOrXOR_givenDuplicatedIDs_shouldReturnID(t *testing.T) {
	code := fake.New[string]()
	creator := NewRulesetCreator()
	got, err := creator.setSingleOrXOR(code, code)

	assert.NoError(t, err)
	assert.Equal(t, code, got)
}

func Test_RulesetCreator_setSingleOrAND_givenNoIDs_shouldReturnError(t *testing.T) {
	creator := NewRulesetCreator()
	_, err := creator.setSingleOrAnd([]string{}...)

	assert.ErrorIs(t, err, puanerror.InvalidArgument)
}

func Test_RulesetCreator_setSingleOrAND_givenDuplicatedIDs_shouldReturnID(t *testing.T) {
	code := fake.New[string]()
	creator := NewRulesetCreator()
	got, err := creator.setSingleOrAnd(code, code)

	assert.NoError(t, err)
	assert.Equal(t, code, got)
}

func Test_AddPrimitives_givenPrimitiveWithPeriodPrefix_shouldReturnError(t *testing.T) {
	creator := NewRulesetCreator()
	err := creator.AddPrimitives("period_")
	assert.ErrorIs(t, err, puanerror.InvalidArgument)
}

func Test_AddPrimitives_givenPrimitiveWithoutPeriodPrefix_shouldReturnNoError(t *testing.T) {
	creator := NewRulesetCreator()
	err := creator.AddPrimitives(fake.New[string]())
	assert.NoError(t, err)
}

// nolint:lll
func Test_RulesetCreator_newPeriodVariables_givenOrderedPeriods_shouldReturnPeriodVariables(
	t *testing.T,
) {
	periods := []Period{
		{
			from: newTestTime("2024-01-01"),
			to:   newTestTime("2024-01-05"),
		},
		{
			from: newTestTime("2024-01-10"),
			to:   newTestTime("2024-01-15"),
		},
	}

	creator := RulesetCreator{}
	periodVariables, err := creator.newPeriodVariables(periods)

	assert.NoError(t, err)
	want := TimeBoundVariables{
		{
			variable: "period_0",
			period:   periods[0],
		},
		{
			variable: "period_1",
			period:   periods[1],
		},
	}
	assert.Equal(t, want, periodVariables)
}

func Test_RulesetCreator_newPeriodVariables_givenInvalidPeriods_shouldReturnError(t *testing.T) {
	type testCase struct {
		name    string
		periods []Period
	}

	cases := []testCase{
		{
			name: "given overlaps",
			periods: []Period{
				{
					from: newTestTime("2024-01-01"),
					to:   newTestTime("2024-01-10"),
				},
				{
					from: newTestTime("2024-01-05"),
					to:   newTestTime("2024-01-15"),
				},
			},
		},
		{
			name: "given duplicates",
			periods: []Period{
				{
					from: newTestTime("2024-01-01"),
					to:   newTestTime("2024-01-10"),
				},
				{
					from: newTestTime("2024-01-01"),
					to:   newTestTime("2024-01-10"),
				},
			},
		},
		{
			name: "not sorted",
			periods: []Period{
				{
					from: newTestTime("2024-01-05"),
					to:   newTestTime("2024-01-10"),
				},
				{
					from: newTestTime("2024-01-01"),
					to:   newTestTime("2024-01-05"),
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			creator := RulesetCreator{}

			_, err := creator.newPeriodVariables(tt.periods)

			assert.Error(t, err)
		})
	}
}

func Test_RulesetCreator_Create_givenForbiddenPeriod_shouldNotBeIncludedInRuleset(
	t *testing.T,
) {
	minute0 := time.Now().Truncate(time.Minute)
	minute30 := minute0.Add(30 * time.Minute)
	minute60 := minute0.Add(60 * time.Minute)

	creator := NewRulesetCreator()

	_ = creator.EnableTime(minute0, minute60)

	_ = creator.ForbidPeriod(
		minute30,
		minute60,
	)

	ruleset, err := creator.Create()

	require.NoError(t, err)
	assert.Len(t, ruleset.PeriodVariables(), 1)
	assert.Equal(
		t,
		ruleset.PeriodVariables()[0].Period(),
		Period{
			from: minute0,
			to:   minute30,
		},
	)
}

func Test_RulesetCreator_Create_givenAssumedInForbiddenPeriod_shouldNotBeIncludedInRuleset(
	t *testing.T,
) {
	minute0 := time.Now().Truncate(time.Minute)
	minute30 := minute0.Add(30 * time.Minute)
	minute60 := minute0.Add(60 * time.Minute)

	creator := NewRulesetCreator()

	_ = creator.EnableTime(minute0, minute60)

	_ = creator.ForbidPeriod(
		minute30,
		minute60,
	)

	code := fake.New[string]()
	_ = creator.AddPrimitives(code)
	_ = creator.AssumeInPeriod(code, minute30, minute60)

	ruleset, err := creator.Create()

	require.NoError(t, err)
	assert.Equal(
		t,
		ruleset.independentVariables,
		[]string{code},
	)
}

func Test_RulesetCreator_Create_givenPreferredInForbiddenPeriod_shouldNotBeIncludedInRuleset(
	t *testing.T,
) {
	minute0 := time.Now().Truncate(time.Minute)
	minute30 := minute0.Add(30 * time.Minute)
	minute60 := minute0.Add(60 * time.Minute)

	creator := NewRulesetCreator()

	_ = creator.EnableTime(minute0, minute60)

	_ = creator.ForbidPeriod(
		minute30,
		minute60,
	)

	code := fake.New[string]()
	_ = creator.AddPrimitives(code)
	_ = creator.PreferInPeriod(code, minute30, minute60)

	ruleset, err := creator.Create()

	require.NoError(t, err)
	assert.Equal(
		t,
		ruleset.independentVariables,
		[]string{code},
	)
	assert.Empty(t, ruleset.PreferredVariables())
}
