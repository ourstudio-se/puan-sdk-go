package weights

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
)

func Test_invalidAction_givenInvalidAction_shouldReturnError(t *testing.T) {
	isInvalid := invalidAction(Action(fake.New[string]()))
	assert.True(t, isInvalid)
}

func Test_invalidAction_givenValidRemoveAction(t *testing.T) {
	isInvalid := invalidAction(REMOVE)
	assert.False(t, isInvalid)
}

func Test_invalidAction_givenValidAddAction(t *testing.T) {
	isInvalid := invalidAction(ADD)
	assert.False(t, isInvalid)
}
