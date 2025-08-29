package puan

import (
	"reflect"
	"testing"
)

func TestSelections_extractSelectionsIDs(t *testing.T) {
	tests := []struct {
		name string
		s    Selections
		want []string
	}{
		{
			"empty selections",
			Selections{},
			[]string{},
		},
		{
			"one selection",
			Selections{{
				ID:     "x",
				Action: ADD,
			}},
			[]string{"x"},
		},
		{
			"two selections",
			Selections{
				{
					ID:     "x",
					Action: ADD,
				},
				{
					ID:     "y",
					Action: ADD,
				},
			},
			[]string{"x", "y"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.ids(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ids() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelections_extractActiveSelections(t *testing.T) {
	tests := []struct {
		name       string
		selections Selections
		want       Selections
	}{
		{
			"with select and unselected selections",
			Selections{
				{
					ID:     "x",
					Action: ADD,
				},
				{
					ID:     "y",
					Action: ADD,
				},
				{
					ID:     "x",
					Action: REMOVE,
				},
			},
			Selections{
				{
					ID:     "y",
					Action: ADD,
				},
			},
		},
		{
			"no redundant selections",
			Selections{
				{
					ID:     "x",
					Action: ADD,
				},
				{
					ID:     "y",
					Action: ADD,
				},
				{
					ID:     "z",
					Action: ADD,
				},
			},
			Selections{
				{
					ID:     "x",
					Action: ADD,
				},
				{
					ID:     "y",
					Action: ADD,
				},
				{
					ID:     "z",
					Action: ADD,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.selections.removeRedundantSelections(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeRedundantSelections() = %v, want %v", got, tt.want)
			}
		})
	}
}
