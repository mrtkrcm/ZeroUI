package performance

import (
	"sync"
	"testing"
)

func TestParserPool(t *testing.T) {
	t.Run("JSONParser", func(t *testing.T) {
		p1 := GetJSONParser()
		if p1 == nil {
			t.Fatal("GetJSONParser returned nil")
		}
		PutJSONParser(p1)

		p2 := GetJSONParser()
		if p2 == nil {
			t.Fatal("GetJSONParser returned nil after put")
		}
		PutJSONParser(p2)
	})

	t.Run("YAMLParser", func(t *testing.T) {
		p1 := GetYAMLParser()
		if p1 == nil {
			t.Fatal("GetYAMLParser returned nil")
		}
		PutYAMLParser(p1)

		p2 := GetYAMLParser()
		if p2 == nil {
			t.Fatal("GetYAMLParser returned nil after put")
		}
		PutYAMLParser(p2)
	})

	t.Run("TOMLParser", func(t *testing.T) {
		p1 := GetTOMLParser()
		if p1 == nil {
			t.Fatal("GetTOMLParser returned nil")
		}
		PutTOMLParser(p1)

		p2 := GetTOMLParser()
		if p2 == nil {
			t.Fatal("GetTOMLParser returned nil after put")
		}
		PutTOMLParser(p2)
	})
}

func TestGetParserForFormat(t *testing.T) {
	tests := []struct {
		format string
		isNil  bool
	}{
		{"json", false},
		{"yaml", false},
		{"yml", false},
		{"toml", false},
		{"unknown", true},
		{"", true},
	}

	for _, tt := range tests {
		p := GetParserForFormat(tt.format)
		if (p == nil) != tt.isNil {
			t.Errorf("GetParserForFormat(%q) returned nil=%v, want nil=%v", tt.format, p == nil, tt.isNil)
		}
		if p != nil {
			PutParserForFormat(tt.format, p)
		}
	}
}

func TestParserPoolConcurrency(t *testing.T) {
	const goroutines = 50
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 3) // Testing 3 parser types

	// Test JSON parsers
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				p := GetJSONParser()
				if p == nil {
					t.Error("GetJSONParser returned nil during concurrent access")
				}
				PutJSONParser(p)
			}
		}()
	}

	// Test YAML parsers
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				p := GetYAMLParser()
				if p == nil {
					t.Error("GetYAMLParser returned nil during concurrent access")
				}
				PutYAMLParser(p)
			}
		}()
	}

	// Test TOML parsers
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				p := GetTOMLParser()
				if p == nil {
					t.Error("GetTOMLParser returned nil during concurrent access")
				}
				PutTOMLParser(p)
			}
		}()
	}

	wg.Wait()
}

func BenchmarkParserPool(b *testing.B) {
	b.Run("JSONWithPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := GetJSONParser()
			PutJSONParser(p)
		}
	})

	b.Run("YAMLWithPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := GetYAMLParser()
			PutYAMLParser(p)
		}
	})

	b.Run("TOMLWithPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := GetTOMLParser()
			PutTOMLParser(p)
		}
	})
}
