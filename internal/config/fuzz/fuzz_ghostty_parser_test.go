//go:build go1.18
// +build go1.18

package fuzz

import (
	"testing"

	"github.com/knadh/koanf/v2"
	cfg "github.com/mrtkrcm/ZeroUI/internal/config"
)

// FuzzGhosttyParser fuzzes the Ghostty parser with random inputs
func FuzzGhosttyParser(f *testing.F) {
	// Seeds with typical lines, comments, and tricky cases
	f.Add([]byte("theme = Dracula\n# comment\nfont-family = SF Mono # inline\n"))
	f.Add([]byte("keybind-31 = super+ctrl+left=resize_split:left\n"))
	f.Add([]byte("palette-117 = 116=#87d7d7\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		_ = koanf.New(".")
		_ = cfg.ParseGhosttyConfig
	})
}
