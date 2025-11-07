package weights

import "github.com/go-errors/errors"

const (
	ADD    Action = "ADD"
	REMOVE Action = "REMOVE"
)

var ErrInvalidAction = errors.New("invalid action")

type Action string

type Selection struct {
	id     string
	action Action
}

func NewSelection(id string, action Action) (Selection, error) {
	if invalidAction(action) {
		return Selection{}, errors.Errorf("%w: %s",
			ErrInvalidAction,
			action,
		)
	}

	return Selection{
		id:     id,
		action: action,
	}, nil
}

func invalidAction(action Action) bool {
	return action != ADD && action != REMOVE
}

type Selections []Selection

func (s Selections) ids() []string {
	ids := make([]string, len(s))
	for i, selection := range s {
		ids[i] = selection.id
	}

	return ids
}
