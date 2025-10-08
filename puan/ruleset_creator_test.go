package puan

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_RuleSetCreator_newTimeBoundVariable_givenTimeEnabled_andValidPeriod(t *testing.T) {
	id := uuid.New().String()
	from := newTestTime("2024-01-01T00:00:00Z")
	to := newTestTime("2024-01-31T00:00:00Z")

	creator := NewRuleSetCreator()
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

func Test_RuleSetCreator_newTimeBoundVariable_givenTimeNotEnabled_shouldReturnError(
	t *testing.T,
) {
	id := uuid.New().String()
	from := newTestTime("2024-01-01T00:00:00Z")
	to := newTestTime("2024-01-31T00:00:00Z")

	creator := NewRuleSetCreator()
	_, err := creator.newTimeBoundVariable(id, from, to)
	assert.Error(t, err)
}

func Test_RuleSetCreator_newTimeBoundVariable_givenTimeEnabled_andInvalidPeriod_shouldReturnError(
	t *testing.T,
) {
	id := uuid.New().String()
	from := newTestTime("2024-01-31T00:00:00Z")
	to := newTestTime("2024-01-01T00:00:00Z")

	creator := NewRuleSetCreator()
	err := creator.EnableTime(
		newTestTime("2024-01-01T00:00:00Z"),
		newTestTime("2024-02-01T00:00:00Z"),
	)
	assert.NoError(t, err)

	_, err = creator.newTimeBoundVariable(id, from, to)
	assert.Error(t, err)
}

// nolint:lll
func Test_RuleSetCreator_newTimeBoundVariable_givenTimeEnabled_andAssumedPeriodOutsideOfEnabledPeriod_shouldReturnError(
	t *testing.T,
) {
	id := uuid.New().String()
	from := newTestTime("2024-01-01T00:00:00Z")
	to := newTestTime("2024-01-31T00:00:00Z")

	creator := NewRuleSetCreator()
	err := creator.EnableTime(
		from.Add(1*24*time.Hour),
		to.Add(-1*24*time.Hour),
	)
	assert.NoError(t, err)

	_, err = creator.newTimeBoundVariable(id, from, to)
	assert.Error(t, err)
}
