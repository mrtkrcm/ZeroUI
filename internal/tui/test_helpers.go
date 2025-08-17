package tui

import (
	"github.com/mrtkrcm/ZeroUI/internal/logging"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
	"github.com/mrtkrcm/ZeroUI/internal/tui/components"
)

// NewTestModel creates a model for testing with a default logger
// and initializes commonly used subcomponents (like helpSystem) so
// tests that interact with those subsystems don't nil-deref or hang.
func NewTestModel(engine *toggle.Engine, initialApp string) (*Model, error) {
	logger, err := logging.NewCharmLogger(logging.DefaultConfig())
	if err != nil {
		// For tests, we can use a minimal config or ignore the logger error
		logger = &logging.CharmLogger{} // minimal logger for tests
	}

	model, err := NewModel(engine, initialApp, logger)
	if err != nil {
		return nil, err
	}

	// Initialize help system so tests that call ShowPage/View won't hit nil.
	// Use the components constructor which provides sensible defaults.
	model.helpSystem = components.NewGlamourHelp()

	// Ensure component sizes are set for deterministic test snapshots.
	// If tests haven't set explicit sizes, provide reasonable defaults.
	if model.width == 0 {
		model.width = 120
	}
	if model.height == 0 {
		model.height = 40
	}
	model.updateComponentSizes()

	return model, nil
}
