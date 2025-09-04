package puan

type SelectionBuilder struct {
	id             string
	subSelectionID *string
	action         Action
}

func NewSelectionBuilder(id string) *SelectionBuilder {
	return &SelectionBuilder{
		id:     id,
		action: ADD,
	}
}

func (b *SelectionBuilder) WithSubSelectionID(subSelectionID string) *SelectionBuilder {
	b.subSelectionID = &subSelectionID
	return b
}

func (b *SelectionBuilder) WithAction(action Action) *SelectionBuilder {
	b.action = action
	return b
}

func (b *SelectionBuilder) Build() Selection {
	return newSelection(b.action, b.id, b.subSelectionID)
}
