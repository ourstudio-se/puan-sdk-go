// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/weights"
)

const url = "http://127.0.0.1:9000"

// TODO: Fix this test
// func Test_package_variant_will_change_when_selecting_another_xor_component(t *testing.T) {
//	/*
//		Given package (a) -> (and(item1,item2,item3), xor(item4,item5)), reversed package rules
//		(and(item1,item2,item3,item4) -> (a), and(item1,item2,item3,item5) -> (a)) and with preferred
//		on variant (a,item4), we test that if variant (a,item1,item2,item3,item5) is preselected
//		and we select single variable (item4), then we will change into the other
//		package variant (a,item1,item2,item3,item4) (and not select single (item4))
//	*/
//	model := pldag.New()
//	model.SetPrimitives("packageA", "item1", "item2", "item3", "item4", "item5")
//
//	variantOneIncludes, _ := model.SetAnd("item1", "item2", "item3", "item4")
//	variantTwoIncludes, _ := model.SetAnd("item1", "item2", "item3", "item5")
//	variantOne, _ := model.SetAnd("packageA", variantOneIncludes)
//	variantTwo, _ := model.SetAnd("packageA", variantTwoIncludes)
//
//	exactlyOneVariant, _ := model.SetXor(variantOne, variantTwo)
//	packageVariants, _ := model.SetImply("packageA", exactlyOneVariant)
//	reversedVariantOne, _ := model.SetImply(variantOneIncludes, "packageA")
//	reversedVariantTwo, _ := model.SetImply(variantTwoIncludes, "packageA")
//
//	root, _ := model.SetAnd(packageVariants, reversedVariantOne, reversedVariantTwo)
//	_ = model.Assume(root)
//
//	xorWithPreference := weights.XORWithPreference{
//		XORID:       exactlyOneVariant,
//		PreferredID: variantOne,
//	}
//
//	polyhedron := model.GeneratePolyhedron()
//	client := glpk.NewClient(url)
//
//	selections := weights.Selections{
//		{
//			ID:     "packageA",
//			Action: weights.ADD,
//		},
//		{
//			ID:     "item5",
//			Action: weights.ADD,
//		},
//		{
//			ID:     "item4",
//			Action: weights.ADD,
//		},
//	}
//	selectionsIDs := selections.ExtractActiveSelectionIDS()
//	objective := weights.CalculateObjective(
//		model.PrimitiveVariables(),
//		selectionsIDs,
//		[]weights.XORWithPreference{xorWithPreference},
//	)
//	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
//	//	assert.Equal(t, 1, solution["packageA"])
//	assert.Equal(t, 1, solution["item1"])
//	assert.Equal(t, 1, solution["item2"])
//	assert.Equal(t, 1, solution["item3"])
//	assert.Equal(t, 1, solution["item4"])
//	assert.Equal(t, 0, solution["item5"])
//}

func Test_select_exactly_one_constrainted_component_with_additional_requirements(t *testing.T) {
	/*
		Exactly one of (a), (b) or (c) must be select. (b) requires another
		variable x. Now, (a) is preselected and we select (b). We expect (b,item1) as result.
	*/

	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "packageC", "item1")
	exactlyOnePackage, _ := model.SetXor("packageA", "packageB", "packageC")
	packageB, _ := model.SetEquivalent("packageB", "item1")

	preferred, _ := model.SetAnd(exactlyOnePackage, "packageA")
	xorWithPreference := weights.XORWithPreference{
		XORID:       exactlyOnePackage,
		PreferredID: preferred,
	}

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
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 0, solution["packageA"])
	assert.Equal(t, 1, solution["packageB"])
	assert.Equal(t, 0, solution["packageC"])
	assert.Equal(t, 1, solution["item1"])
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
	objective := weights.CalculateObjective(model.PrimitiveVariables(), selectionsIDs, nil)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 0, solution["packageA"])
	assert.Equal(t, 0, solution["packageB"])
}

func Test_select_same_selected_exactly_one_constrainted_component(t *testing.T) {
	/*
		Exactly one of (a), (b) or (c) must be select but (a) is preferred.
		(b) has been preselected but is selected again. We now expect (a) to be selected.
	*/

	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "packageC")

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB", "packageC")
	preferred, _ := model.SetAnd("packageA", exactlyOnePackage)
	xorWithPreferred := weights.XORWithPreference{
		XORID:       exactlyOnePackage,
		PreferredID: preferred,
	}

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

	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreferred},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, solution["packageA"])
	assert.Equal(t, 0, solution["packageB"])
	assert.Equal(t, 0, solution["packageC"])
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
	packageBRequiredItems, _ := model.SetEquivalent("packageB", includedItemsInB)

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB")

	anyOfThePackages, _ := model.SetOr("packageA", "packageB")
	itemsInAllPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)
	reversedPackageB, _ := model.SetImply(includedItemsInB, "packageB")

	preferred, _ := model.SetAnd("packageA", exactlyOnePackage)
	xorWithPreferred := weights.XORWithPreference{
		XORID:       exactlyOnePackage,
		PreferredID: preferred,
	}

	root, _ := model.SetAnd(
		exactlyOnePackage,
		packageARequiredItems,
		packageBRequiredItems,
		itemsInAllPackages,
		reversedPackageB,
	)
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
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreferred},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, solution["packageA"])
	assert.Equal(t, 0, solution["packageB"])
	assert.Equal(t, 1, solution["item1"])
	assert.Equal(t, 1, solution["item2"])
	assert.Equal(t, 0, solution["item3"])
}

func Test_select_package_when_xor_between_packages(t *testing.T) {
	/*
		Two packages, (a) and (b), exists with (b) being the larger
		one. They both share a subset of variables and exactly one
		of (a) and (b) must be selected, but with (a) as preferred.
		With nothing being preselected, we select (b) and expects to get (b).
	*/
	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "item1", "item2", "item3")

	includedItemsInA, _ := model.SetAnd("item1", "item2")
	includedItemsInB, _ := model.SetAnd("item1", "item2", "item3")

	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := model.SetEquivalent("packageB", includedItemsInB)

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB")

	anyOfThePackages, _ := model.SetOr("packageA", "packageB")
	itemsInAllPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)
	reversedPackageB, _ := model.SetImply(includedItemsInB, "packageB")

	preferred, _ := model.SetAnd("packageA", exactlyOnePackage)
	xorWithPreferred := weights.XORWithPreference{
		XORID:       exactlyOnePackage,
		PreferredID: preferred,
	}

	root, _ := model.SetAnd(
		exactlyOnePackage,
		packageARequiredItems,
		packageBRequiredItems,
		itemsInAllPackages,
		reversedPackageB,
	)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)
	selections := weights.Selections{
		{
			ID:     "packageB",
			Action: weights.ADD,
		},
	}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreferred},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 0, solution["packageA"])
	assert.Equal(t, 1, solution["packageB"])
	assert.Equal(t, 1, solution["item1"])
	assert.Equal(t, 1, solution["item2"])
	assert.Equal(t, 1, solution["item3"])
}

func Test_upgrade_package_when_xor_between_multiple_packages_case1(t *testing.T) {
	/*
		Here are three packages, (a), (b) and (c), exists. (c) is larger
		than (b) and (b) is larger than (a). We will do several test going from
		nothing preselected to from (b) preselected while selecting (c).
		This tests that we can select larger packages when smaller is already
		selected.
	*/

	model, xorWithPreference := change_package_when_xor_between_multiple_packages_setup()
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
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, solution["packageA"])
	assert.Equal(t, 0, solution["packageB"])
	assert.Equal(t, 0, solution["packageC"])
	assert.Equal(t, 1, solution["item1"])
	assert.Equal(t, 1, solution["item2"])
	assert.Equal(t, 0, solution["item3"])
	assert.Equal(t, 0, solution["item4"])
}

func Test_upgrade_package_when_xor_between_multiple_packages_case2(t *testing.T) {
	/*
		Here are three packages, (a), (b) and (c), exists. (c) is larger
		than (b) and (b) is larger than (a). We will do several test going from
		nothing preselected to from (b) preselected while selecting (c).
		This tests that we can select larger packages when smaller is already
		selected.
	*/

	model, xorWithPreference := change_package_when_xor_between_multiple_packages_setup()
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
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 0, solution["packageA"])
	assert.Equal(t, 1, solution["packageB"])
	assert.Equal(t, 0, solution["packageC"])
	assert.Equal(t, 1, solution["item1"])
	assert.Equal(t, 1, solution["item2"])
	assert.Equal(t, 1, solution["item3"])
	assert.Equal(t, 0, solution["item4"])
}

func Test_upgrade_package_when_xor_between_multiple_packages_case3(t *testing.T) {
	/*
		Here are three packages, (a), (b) and (c), exists. (c) is larger
		than (b) and (b) is larger than (a). We will do several test going from
		nothing preselected to from (b) preselected while selecting (c).
		This tests that we can select larger packages when smaller is already
		selected.
	*/

	model, xorWithPreference := change_package_when_xor_between_multiple_packages_setup()
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
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 0, solution["packageA"])
	assert.Equal(t, 0, solution["packageB"])
	assert.Equal(t, 1, solution["packageC"])
	assert.Equal(t, 1, solution["item1"])
	assert.Equal(t, 1, solution["item2"])
	assert.Equal(t, 1, solution["item3"])
	assert.Equal(t, 1, solution["item4"])
}

func Test_upgrade_package_when_xor_between_multiple_packages_case4(t *testing.T) {
	/*
		Here are three packages, (a), (b) and (c), exists. (c) is larger
		than (b) and (b) is larger than (a). We will do several test going from
		nothing preselected to from (b) preselected while selecting (c).
		This tests that we can select larger packages when smaller is already
		selected.
	*/

	model, xorWithPreference := change_package_when_xor_between_multiple_packages_setup()
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
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 0, solution["packageA"])
	assert.Equal(t, 0, solution["packageB"])
	assert.Equal(t, 1, solution["packageC"])
	assert.Equal(t, 1, solution["item1"])
	assert.Equal(t, 1, solution["item2"])
	assert.Equal(t, 1, solution["item3"])
	assert.Equal(t, 1, solution["item4"])
}

func Test_downgrade_package_when_xor_between_multiple_packages_case1(t *testing.T) {
	/*
	   Here we have three packages, (a), (b) and (c), with (c) being largest
	   and (a) being smallest. We will try and select (b) when (c) is preselected,
	   try select (a) and (b) is selected and try (a) when (c) is selected. All
	   tests should result in the selected package.
	*/
	model, xorWithPreference := change_package_when_xor_between_multiple_packages_setup()
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
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, solution["packageA"])
	assert.Equal(t, 0, solution["packageB"])
	assert.Equal(t, 0, solution["packageC"])
	assert.Equal(t, 1, solution["item1"])
	assert.Equal(t, 1, solution["item2"])
	assert.Equal(t, 0, solution["item3"])
	assert.Equal(t, 0, solution["item4"])
}

func Test_downgrade_package_when_xor_between_multiple_packages_case2(t *testing.T) {
	/*
	   Here we have three packages, (a), (b) and (c), with (c) being largest
	   and (a) being smallest. We will try and select (b) when (c) is preselected,
	   try select (a) and (b) is selected and try (a) when (c) is selected. All
	   tests should result in the selected package.
	*/
	model, xorWithPreference := change_package_when_xor_between_multiple_packages_setup()
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
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, solution["packageA"])
	assert.Equal(t, 0, solution["packageB"])
	assert.Equal(t, 0, solution["packageC"])
	assert.Equal(t, 1, solution["item1"])
	assert.Equal(t, 1, solution["item2"])
	assert.Equal(t, 0, solution["item3"])
	assert.Equal(t, 0, solution["item4"])
}

func Test_downgrade_package_when_xor_between_multiple_packages_case3(t *testing.T) {
	/*
	   Here we have three packages, (a), (b) and (c), with (c) being largest
	   and (a) being smallest. We will try and select (b) when (c) is preselected,
	   try select (a) and (b) is selected and try (a) when (c) is selected. All
	   tests should result in the selected package.
	*/
	model, xorWithPreference := change_package_when_xor_between_multiple_packages_setup()
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
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 0, solution["packageA"])
	assert.Equal(t, 1, solution["packageB"])
	assert.Equal(t, 0, solution["packageC"])
	assert.Equal(t, 1, solution["item1"])
	assert.Equal(t, 1, solution["item2"])
	assert.Equal(t, 1, solution["item3"])
	assert.Equal(t, 0, solution["item4"])
}

func Test_downgrade_package_when_xor_between_multiple_packages_case4(t *testing.T) {
	/*
	   Here we have three packages, (a), (b) and (c), with (c) being largest
	   and (a) being smallest. We will try and select (b) when (c) is preselected,
	   try select (a) and (b) is selected and try (a) when (c) is selected. All
	   tests should result in the selected package.
	*/
	model, xorWithPreference := change_package_when_xor_between_multiple_packages_setup()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, solution["packageA"])
	assert.Equal(t, 0, solution["packageB"])
	assert.Equal(t, 0, solution["packageC"])
	assert.Equal(t, 1, solution["item1"])
	assert.Equal(t, 1, solution["item2"])
	assert.Equal(t, 0, solution["item3"])
	assert.Equal(t, 0, solution["item4"])
}

func change_package_when_xor_between_multiple_packages_setup() (*pldag.Model, weights.XORWithPreference) {
	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "packageC", "item1", "item2", "item3", "item4")

	includedItemsInA, _ := model.SetAnd("item1", "item2")
	includedItemsInB, _ := model.SetAnd("item1", "item2", "item3")
	includedItemsInC, _ := model.SetAnd("item1", "item2", "item3", "item4")

	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := model.SetImply("packageB", includedItemsInB)
	packageCRequiredItems, _ := model.SetEquivalent("packageC", includedItemsInC)

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB", "packageC")
	anyOfThePackages, _ := model.SetOr("packageA", "packageB", "packageC")
	packageBOrC, _ := model.SetOr("packageB", "packageC")

	itemsInAllPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)
	itemsInPackageBOrC, _ := model.SetImply(includedItemsInB, packageBOrC)
	reversedPackageC, _ := model.SetImply(includedItemsInC, "packageC")

	preferred, _ := model.SetAnd("packageA", exactlyOnePackage)
	xorWithPreference := weights.XORWithPreference{
		XORID:       exactlyOnePackage,
		PreferredID: preferred,
	}

	root, _ := model.SetAnd(
		exactlyOnePackage,
		packageARequiredItems,
		packageBRequiredItems,
		packageCRequiredItems,
		itemsInAllPackages,
		itemsInPackageBOrC,
		reversedPackageC,
	)
	_ = model.Assume(root)

	return model, xorWithPreference
}

func Test_default_component_in_package_when_part_in_multiple_xors(t *testing.T) {
	/*
		Package (a) has two variants: (a,itemX,itemY,itemN) and (a,itemX,itemY,itemM,itemO) with
		preferred on the former. Nothing is preselected and we expect
		(a,itemX,itemY,itemN) as our result configuration.
	*/

	model := pldag.New()
	model.SetPrimitives("packageA", "itemX", "itemY", "itemM", "itemN", "itemO")

	includedItemsInVariantOne, _ := model.SetAnd("itemX", "itemY", "itemN")
	includedItemsInVariantTwo, _ := model.SetAnd("itemX", "itemY", "itemM", "itemO")

	packageVariantOne, _ := model.SetAnd("packageA", includedItemsInVariantOne)
	packageVariantTwo, _ := model.SetAnd("packageA", includedItemsInVariantTwo)

	exactlyOnePackage, _ := model.SetXor(packageVariantOne, packageVariantTwo)

	reversedPackageVariantOne, _ := model.SetImply(includedItemsInVariantOne, "packageA")
	reversedPackageVariantTwo, _ := model.SetImply(includedItemsInVariantTwo, "packageA")

	xorWithPreference := weights.XORWithPreference{
		XORID:       exactlyOnePackage,
		PreferredID: packageVariantOne,
	}

	root, _ := model.SetAnd(exactlyOnePackage, reversedPackageVariantOne, reversedPackageVariantTwo)
	_ = model.Assume(root)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := weights.Selections{}

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, solution["packageA"])
	assert.Equal(t, 1, solution["itemX"])
	assert.Equal(t, 1, solution["itemY"])
	assert.Equal(t, 0, solution["itemM"])
	assert.Equal(t, 1, solution["itemN"])
	assert.Equal(t, 0, solution["itemO"])
}

func Test_select_component_with_indirect_package_requirement(t *testing.T) {
	/*
		There exists a chain of requirements: (e) -> (f) -> (a) -> (item1,item2,item3).
		We select (e) and expect our result configuration to (e,f,a,item1,item2,item3)
	*/

	model := pldag.New()
	model.SetPrimitives("packageA", "packageE", "packageF", "item1", "item2", "item3")

	includedItemsInA, _ := model.SetAnd("item1", "item2", "item3")
	packageARequiresItems, _ := model.SetImply("packageA", includedItemsInA)

	packageERequiresF, _ := model.SetImply("packageE", "packageF")
	packageFRequiresA, _ := model.SetImply("packageF", "packageA")

	reversedPackageA, _ := model.SetImply(includedItemsInA, "packageA")

	root, _ := model.SetAnd(
		packageERequiresF,
		packageFRequiresA,
		packageARequiresItems,
		reversedPackageA,
	)
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
	objective := weights.CalculateObjective(model.PrimitiveVariables(), selectionsIDs, nil)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, solution["packageA"])
	assert.Equal(t, 1, solution["packageE"])
	assert.Equal(t, 1, solution["packageF"])
	assert.Equal(t, 1, solution["item1"])
	assert.Equal(t, 1, solution["item2"])
	assert.Equal(t, 1, solution["item3"])
}

func Test_select_single_xor_component_when_another_xor_pair_is_preferred(t *testing.T) {
	/*
		Package (a) has two variants: (a,x) and (a,y,z) with the latter
		being preferred. We select (a,x) and expect the result configuration
		(a,x)
	*/

	model, xorWithPreference := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

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
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, solution["packageA"])
	assert.Equal(t, 1, solution["item1"])
	assert.Equal(t, 0, solution["item2"])
	assert.Equal(t, 0, solution["item3"])
}

func Test_select_xor_pair_when_xor_pair_is_preferred(t *testing.T) {
	/*
	   	Package (a) has two variants: (a,x) and (a,y,z) with the latter
	      being preferred. We select (a,y,z) and expect the result configuration
	      (a,y,z). This test is just to make sure that there is no weird behavior
	      such as an empty configuration as result.
	*/

	model, xorWithPreference := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

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
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, solution["packageA"])
	assert.Equal(t, 0, solution["item1"])
	assert.Equal(t, 1, solution["item2"])
	assert.Equal(t, 1, solution["item3"])
}

func Test_deselect_package_when_xor_pair_is_preferred_over_single_xor_component(t *testing.T) {
	/*
		Given rules a -> xor(x,y), a -> xor(x,z). (y,z) is preferred oved (x)
		If a(yz) is already selected, check that we will remove package when deselecting a
	*/

	model, xorWithPreference := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

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
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 0, solution["packageA"])
	assert.Equal(t, 0, solution["item1"])
	assert.Equal(t, 0, solution["item2"])
	assert.Equal(t, 0, solution["item3"])
}

func Test_select_single_xor_component_when_xor_pair_is_already_selected(t *testing.T) {
	/*
		Given rules a -> xor(x,y), a -> xor(x,z). (y,z) is preferred oved (x)
		If a(yz) is already selected, check that we will select a(x) variant when selecting x
	*/

	model, xorWithPreference := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

	// TODO: How should packages be handled as selections as unit or sequential?
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
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, solution["packageA"])
	assert.Equal(t, 1, solution["item1"])
	assert.Equal(t, 0, solution["item2"])
	assert.Equal(t, 0, solution["item3"])
}

func Test_select_only_package_selected_with_heavy_preferred_in_xor(t *testing.T) {
	/*
		Given rules a -> xor(x,y), a -> xor(x,z). (y,z) is preferred oved (x)
		If a is selected, check that we will get ayz
	*/

	model, xorWithPreference := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, solution["packageA"])
	assert.Equal(t, 0, solution["item1"])
	assert.Equal(t, 1, solution["item2"])
	assert.Equal(t, 1, solution["item3"])
}

func Test_select_one_component_in_xor_pair_when_single_xor_component_is_already_selected(t *testing.T) {
	/*
		Given rules a -> xor(item1,item2), a -> xor(item1,item3). (item2,item3) is preferred oved (item1)
		If a(item1) is already selected, check that we will get a item2 item3 config when selecting item2 (or item3)
	*/

	model, xorWithPreference := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

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
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, solution["packageA"])
	assert.Equal(t, 0, solution["item1"])
	assert.Equal(t, 1, solution["item2"])
	assert.Equal(t, 1, solution["item3"])
}

func Test_select_all_components_selected(t *testing.T) {
	/*
		Given rules a -> xor(item1,item2), a -> xor(item1,item3). (item2,item3) is preferred oved (item1)
		If a(item1) is already selected, check that we will get a item2 item3 config when selecting item2 (or item3)
	*/

	model, xorWithPreference := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

	selections := weights.Selections{
		{
			ID:     "packageA",
			Action: weights.ADD,
		},
		{
			ID:     "item3",
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
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 1, solution["packageA"])
	assert.Equal(t, 0, solution["item1"])
	assert.Equal(t, 1, solution["item2"])
	assert.Equal(t, 1, solution["item3"])
}

func Test_select_nothing_in_optional_xor(t *testing.T) {
	/*
		Given rules a -> xor(item1,item2), a -> xor(item1,item3). (item2,item3) is preferred oved (item1)
		If a(item1) is already selected, check that we will get a item2 item3 config when selecting item2 (or item3)
	*/

	model, xorWithPreference := select_single_xor_component_when_another_xor_pair_is_preferred_setup()

	selections := weights.Selections{}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.ExtractActiveSelectionIDS()
	objective := weights.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]weights.XORWithPreference{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	assert.Equal(t, 0, solution["packageA"])
	assert.Equal(t, 0, solution["item1"])
	assert.Equal(t, 0, solution["item2"])
	assert.Equal(t, 0, solution["item3"])
}

// TODO: Updated all tests with package selection when support for that is implemented.
func select_single_xor_component_when_another_xor_pair_is_preferred_setup() (*pldag.Model, weights.XORWithPreference) {
	model := pldag.New()
	model.SetPrimitives("packageA", "item1", "item2", "item3")

	xorItem1Item2, _ := model.SetXor("item1", "item2")
	xorItem1Item3, _ := model.SetXor("item1", "item3")

	packageExactlyOneOfItem1Item2, _ := model.SetImply("packageA", xorItem1Item2)
	packageExactlyOneOfItem1Item3, _ := model.SetImply("packageA", xorItem1Item3)

	includedItemsInVariantOne, _ := model.SetAnd("item2", "item3")
	packageVariantOne, _ := model.SetAnd("packageA", includedItemsInVariantOne)
	packageVariantTwo, _ := model.SetAnd("packageA", "item1")
	exactlyOneVariant, _ := model.SetXor(packageVariantOne, packageVariantTwo)

	packageA, _ := model.SetImply("packageA", exactlyOneVariant)
	reversePackageVariantOne, _ := model.SetImply(includedItemsInVariantOne, "packageA")
	reversePackageVariantTwo, _ := model.SetImply("item1", "packageA")

	xorWithPreference := weights.XORWithPreference{
		XORID:       exactlyOneVariant,
		PreferredID: packageVariantOne,
	}

	root, _ := model.SetAnd(
		packageA,
		packageExactlyOneOfItem1Item2,
		packageExactlyOneOfItem1Item3,
		reversePackageVariantOne,
		reversePackageVariantTwo,
	)
	// TODO: ska man kunna ta item1 med a,item2,item3?
	_ = model.Assume(root)

	return model, xorWithPreference
}
