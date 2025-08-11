package performance

import (
	"compress/gzip"
	"crypto/tls"
	"net/http"
	"sync"
	"time"
)

var (
	// HTTPClient is a pre-configured, optimized HTTP client
	HTTPClient *http.Client
	clientOnce sync.Once
)

// OptimizedHTTPClient returns a singleton optimized HTTP client
func OptimizedHTTPClient() *http.Client {
	clientOnce.Do(func() {
		HTTPClient = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				// Connection pooling optimization
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				MaxConnsPerHost:     50,
				IdleConnTimeout:     90 * time.Second,

				// TCP optimization
				DisableKeepAlives:  false,
				DisableCompression: false,
				ForceAttemptHTTP2:  true,

				// TLS optimization
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false,
					MinVersion:         tls.VersionTLS12,
				},

				// Timeouts
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}
	})
	return HTTPClient
}

// HTTPResponsePool pools HTTP response bodies to reduce allocations
var HTTPResponsePool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 4096) // Pre-allocate 4KB buffer
	},
}

// GetResponseBuffer gets a response buffer from the pool
func GetResponseBuffer() []byte {
	return HTTPResponsePool.Get().([]byte)[:0]
}

// PutResponseBuffer returns a response buffer to the pool
func PutResponseBuffer(buf []byte) {
	if cap(buf) < 64*1024 { // Don't pool oversized buffers
		HTTPResponsePool.Put(buf)
	}
}

// GzipReaderPool pools gzip readers for decompression
var GzipReaderPool = sync.Pool{
	New: func() interface{} {
		return &gzip.Reader{}
	},
}

// GetGzipReader gets a gzip reader from the pool
func GetGzipReader() *gzip.Reader {
	return GzipReaderPool.Get().(*gzip.Reader)
}

// PutGzipReader returns a gzip reader to the pool
func PutGzipReader(reader *gzip.Reader) {
	reader.Close()
	GzipReaderPool.Put(reader)
}
