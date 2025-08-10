package tui

import (
	"fmt"
	"testing"
)

// BenchmarkFieldViewLookup_Optimized tests O(1) value lookup performance
func BenchmarkFieldViewLookup_Optimized(b *testing.B) {
	// Create a field with many values to show the performance difference
	values := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		values[i] = fmt.Sprintf("value_%d", i)
	}
	
	// Create optimized field view with lookup map
	field := &FieldView{
		Key:         "test_field",
		Type:        "choice",
		Values:      values,
		valueLookup: make(map[string]int),
	}
	
	// Build lookup map
	for i, value := range values {
		field.valueLookup[value] = i
	}
	
	// Test lookup for value near the end (worst case for linear search)
	targetValue := "value_999"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx, found := field.GetValueIndex(targetValue)
		if !found || idx != 999 {
			b.Fatalf("Expected to find value at index 999, got %d, found: %v", idx, found)
		}
	}
}

// BenchmarkFieldViewLookup_Linear simulates the old O(n) approach
func BenchmarkFieldViewLookup_Linear(b *testing.B) {
	// Create a field with many values
	values := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		values[i] = fmt.Sprintf("value_%d", i)
	}
	
	// Test linear search for value near the end (worst case)
	targetValue := "value_999"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		found := false
		idx := 0
		
		// Linear search (old approach)
		for j, value := range values {
			if value == targetValue {
				idx = j
				found = true
				break
			}
		}
		
		if !found || idx != 999 {
			b.Fatalf("Expected to find value at index 999, got %d, found: %v", idx, found)
		}
	}
}

// BenchmarkTUIFieldCaching_Large tests performance with many fields
func BenchmarkTUIFieldCaching_Large(b *testing.B) {
	// Simulate a large config with many fields
	fields := make([]FieldView, 100)
	
	for i := 0; i < 100; i++ {
		// Create field with many values
		values := make([]string, 50)
		lookup := make(map[string]int)
		
		for j := 0; j < 50; j++ {
			value := fmt.Sprintf("field_%d_value_%d", i, j)
			values[j] = value
			lookup[value] = j
		}
		
		fields[i] = FieldView{
			Key:         fmt.Sprintf("field_%d", i),
			Type:        "choice", 
			Values:      values,
			valueLookup: lookup,
		}
	}
	
	b.ResetTimer()
	
	// Test lookup operations across all fields
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			targetValue := fmt.Sprintf("field_%d_value_25", j) // Middle value
			idx, found := fields[j].GetValueIndex(targetValue)
			if !found || idx != 25 {
				b.Fatalf("Field %d: expected index 25, got %d, found: %v", j, idx, found)
			}
		}
	}
}

// TestFieldViewPerformanceComparison validates the performance improvement
func TestFieldViewPerformanceComparison(t *testing.T) {
	// Run benchmarks and compare results
	optimizedResult := testing.Benchmark(BenchmarkFieldViewLookup_Optimized)
	linearResult := testing.Benchmark(BenchmarkFieldViewLookup_Linear)
	
	t.Logf("Optimized O(1) lookup:  %s", optimizedResult)
	t.Logf("Linear O(n) search:     %s", linearResult)
	
	if optimizedResult.NsPerOp() > 0 && linearResult.NsPerOp() > 0 {
		improvement := float64(linearResult.NsPerOp()) / float64(optimizedResult.NsPerOp())
		t.Logf("Performance improvement: %.2fx faster", improvement)
		
		// With 1000 values and lookup for value 999, we should see significant improvement
		if improvement < 10 {
			t.Logf("Warning: Expected >10x improvement for O(1) vs O(n), got %.2fx", improvement)
		}
	}
	
	// Test memory usage
	t.Logf("Memory usage comparison:")
	t.Logf("  Optimized: %d B/op, %d allocs/op", optimizedResult.AllocedBytesPerOp(), optimizedResult.AllocsPerOp())
	t.Logf("  Linear:    %d B/op, %d allocs/op", linearResult.AllocedBytesPerOp(), linearResult.AllocsPerOp())
}

// TestFieldViewEdgeCases tests edge cases for the optimized lookup
func TestFieldViewEdgeCases(t *testing.T) {
	// Test empty field
	emptyField := &FieldView{
		Key:    "empty",
		Type:   "choice", 
		Values: []string{},
	}
	
	idx, found := emptyField.GetValueIndex("nonexistent")
	if found {
		t.Errorf("Expected not found for empty field, got index %d", idx)
	}
	
	// Test field without lookup map (fallback to linear search)
	noLookupField := &FieldView{
		Key:    "no_lookup",
		Type:   "choice",
		Values: []string{"a", "b", "c"},
		// valueLookup intentionally nil to test fallback
	}
	
	idx, found = noLookupField.GetValueIndex("b")
	if !found || idx != 1 {
		t.Errorf("Expected to find 'b' at index 1, got %d, found: %v", idx, found)
	}
	
	// Test HasValue method
	if !noLookupField.HasValue("c") {
		t.Error("Expected HasValue('c') to return true")
	}
	
	if noLookupField.HasValue("nonexistent") {
		t.Error("Expected HasValue('nonexistent') to return false")
	}
}