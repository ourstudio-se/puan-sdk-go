package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/pldag"
	"github.com/ourstudio-se/puan-sdk-go/weights"
)

const url = "http://127.0.0.1:9000"

func Test_package_variant_will_change_when_selecting_another_xor_component(t *testing.T) {
	/*
		Given package (a) -> (and(x,y,z), xor(n,m)), reversed package rules
		(and(x,y,z,n) -> (a), and(x,y,z,m) -> (a)) and with preferred
		on variant (a,n), we test that if variant (a,x,y,z,m) is preselected
		and we select single variable (n), then we will change into the other
		package variant (a,x,y,z,n) (and not select single (n))
	*/

	// Different selections rule [[a,x], [b]] [a,x] means in the same module, [[a],[x]] means that x is chosen in a different module

}

func Test_select_exactly_one_constrainted_component_with_additional_requirements(t *testing.T) {
	/*
		Exactly one of (a), (b) or (c) must be select. (b) requires another
		variable x. Now, (a) is preselected and we select (b). We expect (b,x) as result.
		(a) is compulsoryPreferreds
	*/

	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "packageC", "item1")

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB", "packageC")
	packageB, _ := model.SetImply("packageB", "item1")

	compulsoryPreferreds := weights.CompulsoryPreferreds{
		{
			PrimitiveID: "packageA",
			PreferredID: "packageA",
		},
	}

	optionalPreferreds := weights.OptionalPreferreds{}

	root, _ := model.SetAnd(exactlyOnePackage, packageB)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
}

func Test_select_same_not_constrainted_selected_component(t *testing.T) {
	/*
		(a) requires (b). (b) has been preselected and we select (b)
		again. We now expect the empty set as the result.
	*/

	model := pldag.New()
	model.SetPrimitives("packageA", "packageB")

	packageARequiredPackageB, _ := model.SetImply("packageA", "packageB")

	_ = model.Assume(packageARequiredPackageB)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)
	selections := weights.Selections{
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
		{
			ID:     "packageB",
			Action: weights.REMOVE,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, nil)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
}

func Test_select_same_selected_exactly_one_constrainted_component(t *testing.T) {
	/*
		Exactly one of (a), (b) or (c) must be select but (a) is preferred.
		(b) has been preselected but is selected again. We now expect (a) to be selected.
	*/

	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "packageC")

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB", "packageC")

	compulsoryPreferreds := weights.CompulsoryPreferreds{
		{
			PrimitiveID: "packageA",
			PreferredID: "packageA",
		},
	}

	optionalPreferreds := weights.OptionalPreferreds{}

	root, _ := model.SetAnd(exactlyOnePackage)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)
	selections := weights.Selections{
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
		{
			ID:     "packageB",
			Action: weights.REMOVE,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
}

func Test_select_package_when_xor_between_packages_and_larger_package_is_selected(t *testing.T) {
	/*
		Two packages (a) and (b) exists, with (b) being the larger one
		and exactly one of them has to be selected.
		(b) has been preselected and we select (a). We know expect
		(a) to be selected without nothing left from (b).
	*/
	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "item1", "item2", "item3")

	includedItemsInA, _ := model.SetAnd("item1", "item2")
	includedItemsInB, _ := model.SetAnd("item1", "item2", "item3")

	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := model.SetImply("packageB", includedItemsInB)

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB")

	anyOfThePackages, _ := model.SetOr("packageA", "packageB")
	itemsInAllPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)
	itemsInPackageB, _ := model.SetImply(includedItemsInB, "packageB")

	compulsoryPreferreds := weights.CompulsoryPreferreds{
		{
			PrimitiveID: "packageA",
			PreferredID: "packageA",
		},
	}

	optionalPreferreds := weights.OptionalPreferreds{}

	root, _ := model.SetAnd(exactlyOnePackage, packageARequiredItems, packageBRequiredItems, itemsInAllPackages, itemsInPackageB)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)
	selections := weights.Selections{
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
}

func Test_select_package_when_xor_between_packages(t *testing.T) {
	/*
		Two packages, (a) and (b), exists with (b) being the larger
		one. They both share a subset of variables and exactly one
		of (a) and (b) must be selected, but with (a) as preferred.
		With nothing being preselected, we select (a) and expects to get (a).
	*/
	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "item1", "item2", "item3")

	includedItemsInA, _ := model.SetAnd("item1", "item2")
	includedItemsInB, _ := model.SetAnd("item1", "item2", "item3")

	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := model.SetImply("packageB", includedItemsInB)

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB")

	anyOfThePackages, _ := model.SetOr("packageA", "packageB")
	itemsInAllPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)
	itemsInPackageB, _ := model.SetImply(includedItemsInB, "packageB")

	compulsoryPreferreds := weights.CompulsoryPreferreds{
		{
			PrimitiveID: "packageA",
			PreferredID: "packageA",
		},
	}

	optionalPreferreds := weights.OptionalPreferreds{}

	root, _ := model.SetAnd(exactlyOnePackage, packageARequiredItems, packageBRequiredItems, itemsInAllPackages, itemsInPackageB)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)
	selections := weights.Selections{}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
}

func Test_upgrade_package_when_xor_between_multiple_packages_case1(t *testing.T) {
	/*
		Here are three packages, (a), (b) and (c), exists. (c) is larger
		than (b) and (b) is larger than (a). We will do several test going from
		nothing preselected to from (b) preselected while selecting (c).
		This tests that we can select larger packages when smaller is already
		selected.
	*/

	model, compulsoryPreferreds, optionalPreferreds := change_package_when_xor_between_multiple_packages_setup()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	// Case 1: No preselected packages, select package A
	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])
}

func Test_upgrade_package_when_xor_between_multiple_packages_case2(t *testing.T) {
	/*
		Here are three packages, (a), (b) and (c), exists. (c) is larger
		than (b) and (b) is larger than (a). We will do several test going from
		nothing preselected to from (b) preselected while selecting (c).
		This tests that we can select larger packages when smaller is already
		selected.
	*/

	model, compulsoryPreferreds, optionalPreferreds := change_package_when_xor_between_multiple_packages_setup()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])
}

func Test_upgrade_package_when_xor_between_multiple_packages_case3(t *testing.T) {
	/*
		Here are three packages, (a), (b) and (c), exists. (c) is larger
		than (b) and (b) is larger than (a). We will do several test going from
		nothing preselected to from (b) preselected while selecting (c).
		This tests that we can select larger packages when smaller is already
		selected.
	*/

	model, compulsoryPreferreds, optionalPreferreds := change_package_when_xor_between_multiple_packages_setup()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "packageC",
			Action: weights.ADD,
		},
	}
	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item4"])
}

func Test_upgrade_package_when_xor_between_multiple_packages_case4(t *testing.T) {
	/*
		Here are three packages, (a), (b) and (c), exists. (c) is larger
		than (b) and (b) is larger than (a). We will do several test going from
		nothing preselected to from (b) preselected while selecting (c).
		This tests that we can select larger packages when smaller is already
		selected.
	*/

	model, compulsoryPreferreds, optionalPreferreds := change_package_when_xor_between_multiple_packages_setup()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
		{
			ID:     "packageC",
			Action: weights.ADD,
		},
	}
	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item4"])
}

func Test_downgrade_package_when_xor_between_multiple_packages_case1(t *testing.T) {
	/*
	   Here we have three packages, (a), (b) and (c), with (c) being largest
	   and (a) being smallest. We will try and select (b) when (c) is preselected,
	   try select (a) and (b) is selected and try (a) when (c) is selected. All
	   tests should result in the selected package.
	*/
	model, compulsoryPreferreds, optionalPreferreds := change_package_when_xor_between_multiple_packages_setup()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{
		{
			ID:     "packageC",
			Action: weights.ADD,
		},
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
	}
	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])
}

func Test_downgrade_package_when_xor_between_multiple_packages_case2(t *testing.T) {
	/*
	   Here we have three packages, (a), (b) and (c), with (c) being largest
	   and (a) being smallest. We will try and select (b) when (c) is preselected,
	   try select (a) and (b) is selected and try (a) when (c) is selected. All
	   tests should result in the selected package.
	*/
	model, compulsoryPreferreds, optionalPreferreds := change_package_when_xor_between_multiple_packages_setup()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
	}
	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])
}

func Test_downgrade_package_when_xor_between_multiple_packages_case3(t *testing.T) {
	/*
	   Here we have three packages, (a), (b) and (c), with (c) being largest
	   and (a) being smallest. We will try and select (b) when (c) is preselected,
	   try select (a) and (b) is selected and try (a) when (c) is selected. All
	   tests should result in the selected package.
	*/
	model, compulsoryPreferreds, optionalPreferreds := change_package_when_xor_between_multiple_packages_setup()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{
		{
			ID:     "packageC",
			Action: weights.ADD,
		},
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageB"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageC"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])
}

func change_package_when_xor_between_multiple_packages_setup() (*pldag.Model, weights.CompulsoryPreferreds, weights.OptionalPreferreds) {
	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "packageC", "item1", "item2", "item3", "item4")

	includedItemsInA, _ := model.SetAnd("item1", "item2")
	includedItemsInB, _ := model.SetAnd("item1", "item2", "item3")
	includedItemsInC, _ := model.SetAnd("item1", "item2", "item3", "item4")

	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := model.SetImply("packageB", includedItemsInB)
	packageCRequiredItems, _ := model.SetImply("packageC", includedItemsInC)

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB", "packageC")
	anyOfThePackages, _ := model.SetOr("packageA", "packageB", "packageC")
	packageBOrC, _ := model.SetOr("packageB", "packageC")

	itemsInAllPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)
	itemsInPackageBOrC, _ := model.SetImply(includedItemsInB, packageBOrC)
	itemsInPackageC, _ := model.SetImply(includedItemsInC, "packageC")

	compulsoryPreferreds := weights.CompulsoryPreferreds{
		{
			PrimitiveID: "packageA",
			PreferredID: "packageA",
		},
	}

	optionalPreferreds := weights.OptionalPreferreds{}

	root, _ := model.SetAnd(exactlyOnePackage, packageARequiredItems, packageBRequiredItems, packageCRequiredItems, itemsInAllPackages, itemsInPackageBOrC, itemsInPackageC)
	_ = model.Assume(root)

	return model, compulsoryPreferreds, optionalPreferreds
}

func Test_default_component_in_package_when_part_in_multiple_xors(t *testing.T) {
	/*
		Package (a) has two variants: (a,x,y,n) and (a,x,y,m,o) with
		preferred on the former. Nothing is preselected and we expect
		(a,x,y,n) as our result configuration.
	*/

	model := pldag.New()
	model.SetPrimitives("packageA", "item1", "item2", "item3", "item4", "item5")

	sharedItemsInPackage, _ := model.SetAnd("item1", "item2")
	includedItemsInVariantOne, _ := model.SetAnd("item1", "item2", "item3")
	includedItemsInVariantTwo, _ := model.SetAnd("item1", "item2", "item4", "item5")

	packageVariantOne, _ := model.SetAnd("packageA", includedItemsInVariantOne)
	packageVariantTwo, _ := model.SetAnd("packageA", includedItemsInVariantTwo)

	exactlyOnePackage, _ := model.SetXor(packageVariantOne, packageVariantTwo)
	compulsoryPreferreds := weights.CompulsoryPreferreds{
		{
			PrimitiveID: "packageA",
			PreferredID: packageVariantOne,
		},
	}

	optionalPreferreds := weights.OptionalPreferreds{}

	anyOfTheVariants, _ := model.SetOr(packageVariantOne, packageVariantTwo)
	itemsInAllPackages, _ := model.SetImply(sharedItemsInPackage, anyOfTheVariants)
	itemsInPackageVariantOne, _ := model.SetImply(packageVariantOne, includedItemsInVariantOne)
	itemsInPackageVariantTwo, _ := model.SetImply(packageVariantTwo, includedItemsInVariantTwo)

	root, _ := model.SetAnd(exactlyOnePackage, itemsInAllPackages, itemsInPackageVariantOne, itemsInPackageVariantTwo)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item4"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item5"])
}

func Test_select_component_with_indirect_package_requirement(t *testing.T) {
	/*
		There exists a chain of requirements: (e) -> (f) -> (a) -> (x,y,z).
		We select (e) and expect our result configuration to (e,f,a,x,y,z)
	*/

	model := pldag.New()
	model.SetPrimitives("packageA", "packageE", "packageF", "item1", "item2", "item3")

	includedItemsInA, _ := model.SetAnd("item1", "item2", "item3")
	packageARequiresItems, _ := model.SetImply("packageA", includedItemsInA)

	packageERequiresF, _ := model.SetImply("packageE", "packageF")
	packageFRequiresA, _ := model.SetImply("packageF", "packageA")

	compulsoryPreferreds := weights.CompulsoryPreferreds{}
	optionalPreferreds := weights.OptionalPreferreds{}

	root, _ := model.SetAnd(packageERequiresF, packageFRequiresA, packageARequiresItems)
	_ = model.Assume(root)

	selections := weights.Selections{
		{
			ID:     "packageE",
			Action: weights.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)
	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageE"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageF"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
}

func Test_select_single_xor_component_when_another_xor_pair_is_preferred(t *testing.T) {
	/*
		Package (a) has two variants: (a,x) and (a,y,z) with the latter
		being preferred. We select (a,x) and expect the result configuration
		(a,x)
	*/

	model, compulsoryPreferreds, optionalPreferreds := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "item1",
			Action: weights.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
}

func Test_select_xor_pair_when_xor_pair_is_preferred(t *testing.T) {
	/*
	   	Package (a) has two variants: (a,x) and (a,y,z) with the latter
	      being preferred. We select (a,y,z) and expect the result configuration
	      (a,y,z). This test is just to make sure that there is no weird behavior
	      such as an empty configuration as result.
	*/

	model, compulsoryPreferreds, optionalPreferreds := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "item2",
			Action: weights.ADD,
		},
		{
			ID:     "item3",
			Action: weights.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
}

func Test_deselect_package_when_xor_pair_is_preferred_over_single_xor_component(t *testing.T) {
	/*
		Given rules a -> xor(x,y), a -> xor(x,z). (y,z) is preferred oved (x)
		If a(yz) is already selected, check that we will remove package when deselecting a
	*/

	model, compulsoryPreferreds, optionalPreferreds := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

	selections := weights.Selections{ // TODO: How should packages be handled as selections?
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "item2",
			Action: weights.ADD,
		},
		{
			ID:     "item3",
			Action: weights.ADD,
		},
		{
			ID:     "packageA",
			Action: weights.REMOVE,
		},
		{
			ID:     "item2",
			Action: weights.REMOVE,
		},
		{
			ID:     "item3",
			Action: weights.REMOVE,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 0, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
}

func Test_select_single_xor_component_when_xor_pair_is_already_selected(t *testing.T) {
	/*
		Given rules a -> xor(x,y), a -> xor(x,z). (y,z) is preferred oved (x)
		If a(yz) is already selected, check that we will select a(x) variant when selecting x
	*/

	model, compulsoryPreferreds, optionalPreferreds := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

	selections := weights.Selections{ // TODO: How should packages be handled as selections as unit or sequential?
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "item2",
			Action: weights.ADD,
		},
		{
			ID:     "item3",
			Action: weights.ADD,
		},
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "item1",
			Action: weights.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item3"])
}

func Test_select_only_package_selected_with_heavy_preferred_in_xor(t *testing.T) {
	/*
		Given rules a -> xor(x,y), a -> xor(x,z). (y,z) is preferred oved (x)
		If a is selected, check that we will get ayz
	*/

	model, compulsoryPreferreds, optionalPreferreds := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
}

func Test_select_one_component_in_xor_pair_when_single_xor_component_is_already_selected(t *testing.T) {
	/*
		Given rules a -> xor(item1,item2), a -> xor(item1,item3). (item2,item3) is preferred oved (item1)
		If a(item1) is already selected, check that we will get a item2 item3 config when selecting item2 (or item3)
	*/

	// TODO: Failing test. In python tests, packageA and item1 are selected as a unit.
	// The test succeeds if packageA is chosen again after item1, should it be in which context it is selected?.
	model, compulsoryPreferreds, optionalPreferreds := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

	// variantOnePreffered  a, item2, item3
	// variantTwo  a, item1
	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "item1",
			Action: weights.ADD,
		},
		{
			ID:     "item2",
			Action: weights.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	compulsoryPreferredIDs := compulsoryPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)
	optionalPreferredIDs := optionalPreferreds.ExtractNonRedundantPreferredIDs(selectionsIDs)

	preferredIDs := append([]string{}, optionalPreferredIDs...)
	preferredIDs = append(preferredIDs, compulsoryPreferredIDs...)

	objective := weights.Create(model.PrimitiveVariables(), selectionsIDs, preferredIDs)

	resp, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, len(resp.Solutions))
	assert.Equal(t, 1, resp.Solutions[0].Solution["packageA"])
	assert.Equal(t, 0, resp.Solutions[0].Solution["item1"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item2"])
	assert.Equal(t, 1, resp.Solutions[0].Solution["item3"])
}

func select_single_xor_component_when_another_xor_pair_is_preferred_setup() (*pldag.Model, weights.CompulsoryPreferreds, weights.OptionalPreferreds) {
	model := pldag.New()
	model.SetPrimitives("packageA", "item1", "item2", "item3")

	includedItemsInVariantONe, _ := model.SetAnd("item2", "item3")
	packageVariantOne, _ := model.SetAnd("packageA", includedItemsInVariantONe)
	packageVariantTwo, _ := model.SetAnd("packageA", "item1")

	exactlyOneVariant, _ := model.SetXor(packageVariantOne, packageVariantTwo)
	packageA, _ := model.SetImply("packageA", exactlyOneVariant)
	reversePackageVariantOne, _ := model.SetImply(includedItemsInVariantONe, "packageA")
	reversePackageVariantTwo, _ := model.SetImply("item1", "packageA")

	preferred, _ := model.SetAnd("packageA", packageVariantOne)
	compulsoryPreferreds := weights.CompulsoryPreferreds{}
	optionalPreferreds := weights.OptionalPreferreds{
		{
			PrimitiveID: "packageA",
			PreferredID: preferred,
		},
	}

	root, _ := model.SetAnd(packageA, reversePackageVariantOne, reversePackageVariantTwo)
	_ = model.Assume(root)
	_ = model.Assume(root)

	return model, compulsoryPreferreds, optionalPreferreds
}
