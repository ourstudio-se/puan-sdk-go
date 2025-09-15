package puan

type SelectionBuilder struct {
	id              string
	subSelectionIDs []string
	action          Action
}

func NewSelectionBuilder(id string) *SelectionBuilder {
	return &SelectionBuilder{
		id:              id,
		action:          ADD,
		subSelectionIDs: []string{},
	}
}

func (b *SelectionBuilder) WithSubSelectionID(subSelectionID string) *SelectionBuilder {
	b.subSelectionIDs = append(b.subSelectionIDs, subSelectionID)
	return b
}

func (b *SelectionBuilder) WithAction(action Action) *SelectionBuilder {
	b.action = action
	return b
}

func (b *SelectionBuilder) Build() Selection {
	return newSelection(b.action, b.id, b.subSelectionIDs)
}
