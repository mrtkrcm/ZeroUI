package performance

import (
	"bytes"
	"strings"
	"sync"
)

// BufferPool manages a pool of reusable byte buffers
var BufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// StringBuilderPool manages a pool of reusable string builders
var StringBuilderPool = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

// MapPool manages pools of reusable maps of different types
var MapPool = struct {
	StringInterface *sync.Pool
	StringString    *sync.Pool
	StringBool      *sync.Pool
}{
	StringInterface: &sync.Pool{
		New: func() interface{} {
			return make(map[string]interface{}, 16)
		},
	},
	StringString: &sync.Pool{
		New: func() interface{} {
			return make(map[string]string, 16)
		},
	},
	StringBool: &sync.Pool{
		New: func() interface{} {
			return make(map[string]bool, 16)
		},
	},
}

// SlicePool manages pools of reusable slices
var SlicePool = struct {
	String    *sync.Pool
	Interface *sync.Pool
	Byte      *sync.Pool
}{
	String: &sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 16)
		},
	},
	Interface: &sync.Pool{
		New: func() interface{} {
			return make([]interface{}, 0, 16)
		},
	},
	Byte: &sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 1024)
		},
	},
}

// GetBuffer gets a buffer from the pool
func GetBuffer() *bytes.Buffer {
	buf := BufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutBuffer returns a buffer to the pool
func PutBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	// Only return buffers that aren't too large
	if buf.Cap() > 64*1024 {
		return
	}
	buf.Reset()
	BufferPool.Put(buf)
}

// GetStringBuilder gets a string builder from the pool
func GetStringBuilder() *strings.Builder {
	sb := StringBuilderPool.Get().(*strings.Builder)
	sb.Reset()
	return sb
}

// PutStringBuilder returns a string builder to the pool
func PutStringBuilder(sb *strings.Builder) {
	if sb == nil {
		return
	}
	// Only return builders that aren't too large
	if sb.Cap() > 64*1024 {
		return
	}
	sb.Reset()
	StringBuilderPool.Put(sb)
}

// GetStringInterfaceMap gets a map[string]interface{} from the pool
func GetStringInterfaceMap() map[string]interface{} {
	m := MapPool.StringInterface.Get().(map[string]interface{})
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	return m
}

// PutStringInterfaceMap returns a map to the pool
func PutStringInterfaceMap(m map[string]interface{}) {
	if m == nil {
		return
	}
	// Only return maps that aren't too large
	if len(m) > 1024 {
		return
	}
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	MapPool.StringInterface.Put(m)
}

// GetStringStringMap gets a map[string]string from the pool
func GetStringStringMap() map[string]string {
	m := MapPool.StringString.Get().(map[string]string)
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	return m
}

// PutStringStringMap returns a map to the pool
func PutStringStringMap(m map[string]string) {
	if m == nil {
		return
	}
	// Only return maps that aren't too large
	if len(m) > 1024 {
		return
	}
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	MapPool.StringString.Put(m)
}

// GetStringBoolMap gets a map[string]bool from the pool
func GetStringBoolMap() map[string]bool {
	m := MapPool.StringBool.Get().(map[string]bool)
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	return m
}

// PutStringBoolMap returns a map[string]bool to the pool
func PutStringBoolMap(m map[string]bool) {
	if m == nil {
		return
	}
	// Only return maps that aren't too large
	if len(m) > 1024 {
		return
	}
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	MapPool.StringBool.Put(m)
}

// GetStringSlice gets a string slice from the pool
func GetStringSlice() []string {
	s := SlicePool.String.Get().([]string)
	return s[:0]
}

// PutStringSlice returns a string slice to the pool
func PutStringSlice(s []string) {
	if s == nil {
		return
	}
	// Only return slices that aren't too large
	if cap(s) > 1024 {
		return
	}
	SlicePool.String.Put(s[:0])
}

// GetByteSlice gets a byte slice from the pool
func GetByteSlice(size int) []byte {
	s := SlicePool.Byte.Get().([]byte)
	if cap(s) < size {
		return make([]byte, size)
	}
	return s[:size]
}

// PutByteSlice returns a byte slice to the pool
func PutByteSlice(s []byte) {
	if s == nil {
		return
	}
	// Only return slices that aren't too large
	if cap(s) > 64*1024 {
		return
	}
	SlicePool.Byte.Put(s[:0])
}

// WithBuffer executes a function with a pooled buffer
func WithBuffer(fn func(*bytes.Buffer)) {
	buf := GetBuffer()
	defer PutBuffer(buf)
	fn(buf)
}

// WithStringBuilder executes a function with a pooled string builder
func WithStringBuilder(fn func(*strings.Builder) string) string {
	sb := GetStringBuilder()
	defer PutStringBuilder(sb)
	return fn(sb)
}

// BuildString efficiently builds a string using a pooled builder
func BuildString(parts ...string) string {
	return WithStringBuilder(func(sb *strings.Builder) string {
		for _, part := range parts {
			sb.WriteString(part)
		}
		return sb.String()
	})
}

// BuildStringWithSeparator efficiently builds a string with separator
func BuildStringWithSeparator(separator string, parts ...string) string {
	if len(parts) == 0 {
		return ""
	}

	return WithStringBuilder(func(sb *strings.Builder) string {
		for i, part := range parts {
			if i > 0 {
				sb.WriteString(separator)
			}
			sb.WriteString(part)
		}
		return sb.String()
	})
}

// Aliases for backward compatibility
var (
	GetBuilder = GetStringBuilder
	PutBuilder = PutStringBuilder
)

// GetSpacer returns a string of spaces for padding
func GetSpacer(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}
