package ratcheting

/*
This benchmark suite compares two approaches to ratcheting validation:
1. Validate_Struct1: Runs validation first, then checks for unchanged objects
2. Validate_Struct2: Checks for unchanged objects first, then runs validation

To run all the benchmarks:
  $ go test -bench=. -benchmem ./staging/src/k8s.io/code-generator/cmd/validation-gen/output_tests/ratcheting

To run just Update benchmarks:
  $ go test -bench=Update -benchmem ./staging/src/k8s.io/code-generator/cmd/validation-gen/output_tests/ratcheting

To compare Struct1 vs Struct2 performance:
  $ go test -bench=InvalidButUnchanged -benchmem ./staging/src/k8s.io/code-generator/cmd/validation-gen/output_tests/ratcheting

To export benchmark results for comparison:
  $ go test -bench=Update -benchmem ./staging/src/k8s.io/code-generator/cmd/validation-gen/output_tests/ratcheting > benchmark_results.txt

Expected Results:
- Struct2 should be significantly faster for InvalidButUnchanged and NoChange cases
- Both should have similar performance for other scenarios
*/

import (
	"context"
	"testing"

	// Pretend these imports work as per the prompt

	operation "k8s.io/apimachinery/pkg/api/operation"
)

// createValidStruct1 creates a valid Struct1 instance
func createValidStruct1() *Struct1 {
	return &Struct1{
		TypeMeta: 1,
		ListField: []OtherStruct1{
			{Key1Field: "key1", DataField: "data1"},
			{Key1Field: "key2", DataField: "data2"},
		},
		MinField: 5,
	}
}

// createInvalidStruct1 creates an invalid Struct1 instance
func createInvalidStruct1() *Struct1 {
	return &Struct1{
		TypeMeta: 1,
		ListField: []OtherStruct1{
			{Key1Field: "key1", DataField: "data1"},
			{Key1Field: "key2", DataField: "data2"},
		},
		MinField: 0, // Invalid: less than minimum of 1
	}
}

// createValidStruct2 creates a valid Struct2 instance
func createValidStruct2() *Struct2 {
	return &Struct2{
		TypeMeta: 1,
		ListField: []OtherStruct1{
			{Key1Field: "key1", DataField: "data1"},
			{Key1Field: "key2", DataField: "data2"},
		},
		MinField: 5,
	}
}

// createInvalidStruct2 creates an invalid Struct2 instance
func createInvalidStruct2() *Struct2 {
	return &Struct2{
		TypeMeta: 1,
		ListField: []OtherStruct1{
			{Key1Field: "key1", DataField: "data1"},
			{Key1Field: "key2", DataField: "data2"},
		},
		MinField: 0, // Invalid: less than minimum of 1
	}
}

// createModifiedStruct1 creates a Struct1 instance with multiple field changes
func createModifiedStruct1(base *Struct1) *Struct1 {
	// Make a copy and modify multiple fields
	modified := &Struct1{
		TypeMeta:  base.TypeMeta + 1,
		ListField: make([]OtherStruct1, len(base.ListField)),
		MinField:  base.MinField + 2,
	}

	// Copy and modify list fields
	for i, item := range base.ListField {
		modified.ListField[i] = OtherStruct1{
			Key1Field: item.Key1Field,               // Keep key (required for map matching)
			DataField: item.DataField + "-modified", // Modify data
		}
	}

	// Add a new list item
	modified.ListField = append(modified.ListField, OtherStruct1{
		Key1Field: "newKey",
		DataField: "newData",
	})

	return modified
}

// createModifiedStruct2 creates a Struct2 instance with multiple field changes
func createModifiedStruct2(base *Struct2) *Struct2 {
	// Make a copy and modify multiple fields
	modified := &Struct2{
		TypeMeta:  base.TypeMeta + 1,
		ListField: make([]OtherStruct1, len(base.ListField)),
		MinField:  base.MinField + 2,
	}

	// Copy and modify list fields
	for i, item := range base.ListField {
		modified.ListField[i] = OtherStruct1{
			Key1Field: item.Key1Field,               // Keep key (required for map matching)
			DataField: item.DataField + "-modified", // Modify data
		}
	}

	// Add a new list item
	modified.ListField = append(modified.ListField, OtherStruct1{
		Key1Field: "newKey",
		DataField: "newData",
	})

	return modified
}

// ---- Benchmarks for Update operations ----

// BenchmarkValidateStruct1Update_NoChange benchmarks validation when updating a Struct1 with no changes
func BenchmarkValidateStruct1Update_NoChange(b *testing.B) {
	ctx := context.Background()
	old := createValidStruct1()
	obj := createValidStruct1() // Same as old
	op := operation.Operation{Type: operation.Update}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate_Struct1(ctx, op, nil, obj, old)
	}
}

// BenchmarkValidateStruct2Update_NoChange benchmarks validation when updating a Struct2 with no changes
func BenchmarkValidateStruct2Update_NoChange(b *testing.B) {
	ctx := context.Background()
	old := createValidStruct2()
	obj := createValidStruct2() // Same as old
	op := operation.Operation{Type: operation.Update}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate_Struct2(ctx, op, nil, obj, old)
	}
}

// BenchmarkValidateStruct1Update_InvalidToValid benchmarks validation of fixing an invalid Struct1
func BenchmarkValidateStruct1Update_InvalidToValid(b *testing.B) {
	ctx := context.Background()
	old := createInvalidStruct1()
	obj := createValidStruct1() // Fixed the invalid field
	op := operation.Operation{Type: operation.Update}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate_Struct1(ctx, op, nil, obj, old)
	}
}

// BenchmarkValidateStruct2Update_InvalidToValid benchmarks validation of fixing an invalid Struct2
func BenchmarkValidateStruct2Update_InvalidToValid(b *testing.B) {
	ctx := context.Background()
	old := createInvalidStruct2()
	obj := createValidStruct2() // Fixed the invalid field
	op := operation.Operation{Type: operation.Update}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate_Struct2(ctx, op, nil, obj, old)
	}
}

// BenchmarkValidateStruct1Update_ValidToInvalid benchmarks validation when making a valid Struct1 invalid
func BenchmarkValidateStruct1Update_ValidToInvalid(b *testing.B) {
	ctx := context.Background()
	old := createValidStruct1()
	obj := createInvalidStruct1() // Made it invalid
	op := operation.Operation{Type: operation.Update}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate_Struct1(ctx, op, nil, obj, old)
	}
}

// BenchmarkValidateStruct2Update_ValidToInvalid benchmarks validation when making a valid Struct2 invalid
func BenchmarkValidateStruct2Update_ValidToInvalid(b *testing.B) {
	ctx := context.Background()
	old := createValidStruct2()
	obj := createInvalidStruct2() // Made it invalid
	op := operation.Operation{Type: operation.Update}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate_Struct2(ctx, op, nil, obj, old)
	}
}

// BenchmarkValidateStruct1Update_InvalidToInvalid benchmarks validation when updating an invalid Struct1 to another invalid state
func BenchmarkValidateStruct1Update_InvalidToInvalid(b *testing.B) {
	ctx := context.Background()
	old := createInvalidStruct1()
	// Copy the invalid struct but change something else
	obj := createInvalidStruct1()
	if len(obj.ListField) > 0 {
		obj.ListField[0].DataField = "changed data"
	}
	op := operation.Operation{Type: operation.Update}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate_Struct1(ctx, op, nil, obj, old)
	}
}

// BenchmarkValidateStruct2Update_InvalidToInvalid benchmarks validation when updating an invalid Struct2 to another invalid state
func BenchmarkValidateStruct2Update_InvalidToInvalid(b *testing.B) {
	ctx := context.Background()
	old := createInvalidStruct2()
	// Copy the invalid struct but change something else
	obj := createInvalidStruct2()
	if len(obj.ListField) > 0 {
		obj.ListField[0].DataField = "changed data"
	}
	op := operation.Operation{Type: operation.Update}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate_Struct2(ctx, op, nil, obj, old)
	}
}

// BenchmarkValidateStruct1Update_InvalidButUnchanged benchmarks validation of an invalid but unchanged Struct1
func BenchmarkValidateStruct1Update_InvalidButUnchanged(b *testing.B) {
	ctx := context.Background()
	old := createInvalidStruct1()
	// Create identical copy without any changes
	obj := &Struct1{
		TypeMeta: old.TypeMeta,
		MinField: old.MinField,
	}

	// Deep copy the list field
	obj.ListField = make([]OtherStruct1, len(old.ListField))
	for i, item := range old.ListField {
		obj.ListField[i] = item
	}

	op := operation.Operation{Type: operation.Update}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate_Struct1(ctx, op, nil, obj, old)
	}
}

// BenchmarkValidateStruct2Update_InvalidButUnchanged benchmarks validation of an invalid but unchanged Struct2
func BenchmarkValidateStruct2Update_InvalidButUnchanged(b *testing.B) {
	ctx := context.Background()
	old := createInvalidStruct2()
	// Create identical copy without any changes
	obj := &Struct2{
		TypeMeta: old.TypeMeta,
		MinField: old.MinField,
	}

	// Deep copy the list field
	obj.ListField = make([]OtherStruct1, len(old.ListField))
	for i, item := range old.ListField {
		obj.ListField[i] = item
	}

	op := operation.Operation{Type: operation.Update}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate_Struct2(ctx, op, nil, obj, old)
	}
}

// BenchmarkValidateStruct1Update_ComplexChanges benchmarks validation with multiple field changes
func BenchmarkValidateStruct1Update_ComplexChanges(b *testing.B) {
	ctx := context.Background()
	old := createValidStruct1()
	obj := createModifiedStruct1(old)
	op := operation.Operation{Type: operation.Update}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate_Struct1(ctx, op, nil, obj, old)
	}
}

// BenchmarkValidateStruct2Update_ComplexChanges benchmarks validation with multiple field changes
func BenchmarkValidateStruct2Update_ComplexChanges(b *testing.B) {
	ctx := context.Background()
	old := createValidStruct2()
	obj := createModifiedStruct2(old)
	op := operation.Operation{Type: operation.Update}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate_Struct2(ctx, op, nil, obj, old)
	}
}
