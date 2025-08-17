package tui

import (
	"github.com/mrtkrcm/ZeroUI/internal/logging"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

// NewTestModel creates a model for testing with a default logger
func NewTestModel(engine *toggle.Engine, initialApp string) (*Model, error) {
	logger, err := logging.NewCharmLogger(logging.DefaultConfig())
	if err != nil {
		// For tests, we can use a minimal config or ignore the logger error
		logger = &logging.CharmLogger{} // minimal logger for tests
	}
	return NewModel(engine, initialApp, logger)
}