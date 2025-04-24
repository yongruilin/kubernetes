package ratcheting

import (
	"context"
	"fmt"
	"testing"

	operation "k8s.io/apimachinery/pkg/api/operation"
)

// createElement1 creates an Element1 with specified nesting depth
func createElement1(depth int) *Element1 {
	if depth <= 0 {
		return &Element1{
			TypeMeta: 1,
			Value:    nil,
		}
	}

	return &Element1{
		TypeMeta: 1,
		Value:    createElement1(depth - 1),
	}
}

// createElement2 creates an Element2 with specified nesting depth
func createElement2(depth int) *Element2 {
	if depth <= 0 {
		return &Element2{
			TypeMeta: 1,
			Value:    nil,
		}
	}

	return &Element2{
		TypeMeta: 1,
		Value:    createElement2(depth - 1),
	}
}

// createModifiedElement1 creates a modified version of an Element1
// The modification occurs at the specified modifyAtDepth
func createModifiedElement1(base *Element1, currentDepth, modifyAtDepth int) *Element1 {
	if base == nil {
		return nil
	}

	result := &Element1{
		TypeMeta: base.TypeMeta,
	}

	// If we're at the depth to modify
	if currentDepth == modifyAtDepth {
		// Modify the TypeMeta
		result.TypeMeta = base.TypeMeta + 1
	}

	// Recursively handle the Value field
	if base.Value != nil {
		result.Value = createModifiedElement1(base.Value, currentDepth+1, modifyAtDepth)
	}

	return result
}

// createModifiedElement2 creates a modified version of an Element2
// The modification occurs at the specified modifyAtDepth
func createModifiedElement2(base *Element2, currentDepth, modifyAtDepth int) *Element2 {
	if base == nil {
		return nil
	}

	result := &Element2{
		TypeMeta: base.TypeMeta,
	}

	// If we're at the depth to modify
	if currentDepth == modifyAtDepth {
		// Modify the TypeMeta
		result.TypeMeta = base.TypeMeta + 1
	}

	// Recursively handle the Value field
	if base.Value != nil {
		result.Value = createModifiedElement2(base.Value, currentDepth+1, modifyAtDepth)
	}

	return result
}

// ---------- Direct Comparison Benchmarks ----------

// BenchmarkElementsComparison_NoChange directly compares Element1 vs Element2 with no changes
func BenchmarkElementsComparison_NoChange(b *testing.B) {
	ctx := context.Background()
	op := operation.Operation{Type: operation.Update}

	// Test with a deep structure to emphasize differences
	depth := 20

	// Setup Element1
	old1 := createElement1(depth)
	obj1 := createElement1(depth) // same as old

	// Setup Element2
	old2 := createElement2(depth)
	obj2 := createElement2(depth) // same as old

	b.Run("Element1", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Validate_Element1(ctx, op, nil, obj1, old1)
		}
	})

	b.Run("Element2", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Validate_Element2(ctx, op, nil, obj2, old2)
		}
	})
}

// ---------- Individual Test Cases Grouped By Scenario ----------

// BenchmarkValidateUpdate_NoChange benchmarks validation of unchanged elements
func BenchmarkValidateUpdate_NoChange(b *testing.B) {
	ctx := context.Background()
	op := operation.Operation{Type: operation.Update}

	for _, depth := range []int{1, 5, 10} {
		// Test Element1
		b.Run(fmt.Sprintf("Element1_Depth%d", depth), func(b *testing.B) {
			old := createElement1(depth)
			obj := createElement1(depth) // Same as old

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Validate_Element1(ctx, op, nil, obj, old)
			}
		})

		// Test Element2
		b.Run(fmt.Sprintf("Element2_Depth%d", depth), func(b *testing.B) {
			old := createElement2(depth)
			obj := createElement2(depth) // Same as old

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Validate_Element2(ctx, op, nil, obj, old)
			}
		})
	}
}

// BenchmarkValidateUpdate_ChangeAtRoot benchmarks validation when changing the root element
func BenchmarkValidateUpdate_ChangeAtRoot(b *testing.B) {
	ctx := context.Background()
	op := operation.Operation{Type: operation.Update}

	for _, depth := range []int{1, 5, 10} {
		// Test Element1
		b.Run(fmt.Sprintf("Element1_Depth%d", depth), func(b *testing.B) {
			old := createElement1(depth)
			obj := createModifiedElement1(old, 0, 0) // Modify at root

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Validate_Element1(ctx, op, nil, obj, old)
			}
		})

		// Test Element2
		b.Run(fmt.Sprintf("Element2_Depth%d", depth), func(b *testing.B) {
			old := createElement2(depth)
			obj := createModifiedElement2(old, 0, 0) // Modify at root

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Validate_Element2(ctx, op, nil, obj, old)
			}
		})
	}
}

// BenchmarkValidateUpdate_ChangeAtLeaf benchmarks validation when changing the leaf element
func BenchmarkValidateUpdate_ChangeAtLeaf(b *testing.B) {
	ctx := context.Background()
	op := operation.Operation{Type: operation.Update}

	for _, depth := range []int{600} {
		// Test Element1
		b.Run(fmt.Sprintf("Element1_Depth%d", depth), func(b *testing.B) {
			old := createElement1(depth)
			obj := createModifiedElement1(old, 0, depth-1) // Modify at leaf

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Validate_Element1(ctx, op, nil, obj, old)
			}
		})

		// Test Element2
		b.Run(fmt.Sprintf("Element2_Depth%d", depth), func(b *testing.B) {
			old := createElement2(depth)
			obj := createModifiedElement2(old, 0, depth-1) // Modify at leaf

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Validate_Element2(ctx, op, nil, obj, old)
			}
		})
	}
}

// BenchmarkValidateUpdate_ChangeAtMiddle benchmarks validation when changing the middle element
func BenchmarkValidateUpdate_ChangeAtMiddle(b *testing.B) {
	ctx := context.Background()
	op := operation.Operation{Type: operation.Update}

	for _, depth := range []int{3, 5, 10} {
		// Test Element1
		b.Run(fmt.Sprintf("Element1_Depth%d", depth), func(b *testing.B) {
			old := createElement1(depth)
			middleDepth := depth / 2
			obj := createModifiedElement1(old, 0, middleDepth) // Modify at middle

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Validate_Element1(ctx, op, nil, obj, old)
			}
		})

		// Test Element2
		b.Run(fmt.Sprintf("Element2_Depth%d", depth), func(b *testing.B) {
			old := createElement2(depth)
			middleDepth := depth / 2
			obj := createModifiedElement2(old, 0, middleDepth) // Modify at middle

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Validate_Element2(ctx, op, nil, obj, old)
			}
		})
	}
}
