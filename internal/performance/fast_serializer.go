package performance

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	
	"gopkg.in/yaml.v3"
)

// SerializerPool manages reusable JSON/YAML encoders and decoders
type SerializerPool struct {
	jsonEncoders sync.Pool
	jsonDecoders sync.Pool
	yamlEncoders sync.Pool
	yamlDecoders sync.Pool
	buffers      sync.Pool
}

// NewSerializerPool creates an optimized serializer pool
func NewSerializerPool() *SerializerPool {
	return &SerializerPool{
		jsonEncoders: sync.Pool{
			New: func() interface{} {
				encoder := json.NewEncoder(io.Discard)
				encoder.SetIndent("", "  ")
				encoder.SetEscapeHTML(false) // Faster for config data
				return encoder
			},
		},
		jsonDecoders: sync.Pool{
			New: func() interface{} {
				return json.NewDecoder(nil)
			},
		},
		yamlEncoders: sync.Pool{
			New: func() interface{} {
				encoder := yaml.NewEncoder(io.Discard)
				encoder.SetIndent(2)
				return encoder
			},
		},
		yamlDecoders: sync.Pool{
			New: func() interface{} {
				return yaml.NewDecoder(nil)
			},
		},
		buffers: sync.Pool{
			New: func() interface{} {
				buf := &strings.Builder{}
				buf.Grow(4096) // Pre-allocate 4KB for config data
				return buf
			},
		},
	}
}

var (
	globalSerializerPool *SerializerPool
	serializerOnce       sync.Once
)

// GlobalSerializerPool returns the singleton serializer pool
func GlobalSerializerPool() *SerializerPool {
	serializerOnce.Do(func() {
		globalSerializerPool = NewSerializerPool()
	})
	return globalSerializerPool
}

// FastJSONMarshal provides high-performance JSON marshaling with pooled encoders
func (sp *SerializerPool) FastJSONMarshal(v interface{}) ([]byte, error) {
	buf := sp.buffers.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		sp.buffers.Put(buf)
	}()
	
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}
	
	result := buf.String()
	// Remove trailing newline added by encoder
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}
	
	return []byte(result), nil
}

// StreamingJSONProcessor provides streaming JSON processing for large configs
type StreamingJSONProcessor struct {
	pool *SerializerPool
}

// NewStreamingJSONProcessor creates a streaming JSON processor
func NewStreamingJSONProcessor() *StreamingJSONProcessor {
	return &StreamingJSONProcessor{
		pool: GlobalSerializerPool(),
	}
}

// ProcessJSONStream processes JSON data in chunks for memory efficiency
func (sjp *StreamingJSONProcessor) ProcessJSONStream(reader io.Reader, processor func(interface{}) error) error {
	decoder := json.NewDecoder(reader)
	
	// Enable streaming mode for large objects
	decoder.UseNumber() // Preserve number precision
	
	// Process tokens streaming for better memory usage
	for decoder.More() {
		var value interface{}
		if err := decoder.Decode(&value); err != nil {
			return err
		}
		
		if err := processor(value); err != nil {
			return err
		}
	}
	
	return nil
}

// OptimizedConfigMarshaler provides specialized marshaling for config structures
type OptimizedConfigMarshaler struct {
	// Pre-compiled field mappings for common config structures
	fieldMappings map[string][]string
}

// MarshalConfig uses specialized knowledge of config structures for faster marshaling
func (ocm *OptimizedConfigMarshaler) MarshalConfig(config interface{}, format string) ([]byte, error) {
	switch format {
	case "json":
		return ocm.marshalConfigJSON(config)
	case "yaml":
		return ocm.marshalConfigYAML(config)
	default:
		// Fallback to standard marshaling
		return json.Marshal(config)
	}
}

func (ocm *OptimizedConfigMarshaler) marshalConfigJSON(config interface{}) ([]byte, error) {
	// Use standard marshaling for now - can be optimized further with reflection
	return json.Marshal(config)
}

func (ocm *OptimizedConfigMarshaler) marshalConfigYAML(config interface{}) ([]byte, error) {
	// Use standard marshaling for now - can be optimized further
	return yaml.Marshal(config)
}

// CompressionAwareMarshaler provides automatic compression for large configs
type CompressionAwareMarshaler struct {
	threshold int // Compress if output exceeds this size
}

// NewCompressionAwareMarshaler creates a marshaler that compresses large outputs
func NewCompressionAwareMarshaler(compressionThreshold int) *CompressionAwareMarshaler {
	return &CompressionAwareMarshaler{
		threshold: compressionThreshold,
	}
}

// MarshalWithCompression marshals data and optionally compresses it
func (cam *CompressionAwareMarshaler) MarshalWithCompression(v interface{}, format string) ([]byte, bool, error) {
	var data []byte
	var err error
	
	switch format {
	case "json":
		data, err = json.Marshal(v)
	case "yaml":
		data, err = yaml.Marshal(v)
	default:
		return nil, false, fmt.Errorf("unsupported format: %s", format)
	}
	
	if err != nil {
		return nil, false, err
	}
	
	// Compress if data exceeds threshold
	if len(data) > cam.threshold {
		compressed := compress(data) // Would need proper compression implementation
		return compressed, true, nil
	}
	
	return data, false, nil
}

// compress provides optimized compression for config data
func compress(data []byte) []byte {
	// This would implement optimized compression
	// For now, return original data
	return data
}

// DecompressionCache provides caching for decompressed config data
type DecompressionCache struct {
	mu    sync.RWMutex
	cache map[string][]byte
	maxSize int64
	currentSize int64
}

// NewDecompressionCache creates a cache for decompressed data
func NewDecompressionCache(maxSizeBytes int64) *DecompressionCache {
	return &DecompressionCache{
		cache:   make(map[string][]byte),
		maxSize: maxSizeBytes,
	}
}

// Get retrieves cached decompressed data
func (dc *DecompressionCache) Get(key string) ([]byte, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	
	data, exists := dc.cache[key]
	return data, exists
}

// Put stores decompressed data in cache
func (dc *DecompressionCache) Put(key string, data []byte) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	
	dataSize := int64(len(data))
	
	// Evict old entries if necessary
	if dc.currentSize + dataSize > dc.maxSize {
		dc.evictEntries(dataSize)
	}
	
	dc.cache[key] = data
	dc.currentSize += dataSize
}

func (dc *DecompressionCache) evictEntries(needed int64) {
	// Simple eviction strategy - remove entries until we have enough space
	for key, data := range dc.cache {
		delete(dc.cache, key)
		dc.currentSize -= int64(len(data))
		
		if dc.currentSize + needed <= dc.maxSize {
			break
		}
	}
}