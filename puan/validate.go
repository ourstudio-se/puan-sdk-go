package puan

import (
	"time"

	"github.com/go-errors/errors"
	"github.com/ourstudio-se/puan-sdk-go/internal/utils"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
)

func validateRuleset(ruleset Ruleset) error {
	if ruleset.polyhedron == nil {
		return errors.Errorf("%w: ruleset is required", puanerror.InvalidArgument)
	}
	return nil
}

func validateTimestamps(from *time.Time, to *time.Time) error {
	if from != nil && to != nil {
		if from.After(*to) {
			return errors.Errorf(
				"%w: from '%s' must be before to '%s'",
				puanerror.InvalidArgument,
				from,
				to,
			)
		}
	}
	return nil
}

func validateSelections(
	ruleset Ruleset,
	selections Selections,
) error {
	for _, selection := range selections {
		if !utils.ContainsAll(ruleset.selectableVariables, selection.IDs()) {
			return errors.Errorf(
				"%w: selection contains non-selectable variables: %v",
				puanerror.InvalidArgument,
				selection,
			)
		}

		hasSubSelection := len(selection.subSelectionIDs) > 0
		if hasSubSelection {
			if utils.ContainsAny(selection.IDs(), ruleset.independentVariables) {
				return errors.Errorf(
					"%w: independent variables cannot be part of a composite selections: %v",
					puanerror.InvalidArgument,
					selection,
				)
			}
		}
	}

	return nil
}
