package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mrtkrcm/ZeroUI/internal/logging"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
)

// ViewState represents the view states for the app
type ViewState int

const (
	ListView     ViewState = iota // List-based app selection
	FormView                      // Dynamic forms for configuration
	HelpView                      // Rich markdown help system
	ProgressView                  // Progress and loading operations
)

// App represents the TUI application with modern components
type App struct {
	engine     *toggle.Engine
	initialApp string
	program    *tea.Program
	ctx        context.Context
	logger     *logging.CharmLogger
}

// NewApp creates a new TUI application
func NewApp(initialApp string) (*App, error) {
	// Initialize logging first
	logConfig := logging.DefaultConfig()
	logger, err := logging.NewCharmLogger(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize the toggle engine
	engine, err := toggle.NewEngine()
	if err != nil {
		logger.LogError(err, "engine_initialization")
		return nil, fmt.Errorf("failed to create toggle engine: %w", err)
	}

	logger.Info("ZeroUI initialized",
		"initial_app", initialApp,
		"log_file", logger.GetFileLocation())

	return &App{
		engine:     engine,
		initialApp: initialApp,
		logger:     logger,
	}, nil
}

// Run starts the TUI application
func (app *App) Run() error {
	return app.RunWithContext(context.Background())
}

// RunWithContext starts the TUI application with a specific context
func (app *App) RunWithContext(ctx context.Context) error {
	app.ctx = ctx

	// Create the model
	model, err := NewModel(app.engine, app.initialApp, app.logger)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}
	model.ctx = ctx

	// Set up recovery handler
	defer func() {
		if r := recover(); r != nil {
			app.logger.LogPanic(r, "app_crash")
			app.logger.Error("Application crashed", "error", r)
		}
	}()

	// Create the Bubble Tea program with optimizations
	options := []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	}

	// Add context if provided
	if ctx != nil {
		options = append(options, tea.WithContext(ctx))
	}

	app.program = tea.NewProgram(model, options...)

	// Log startup
	app.logger.Info("Starting TUI application",
		"initial_app", app.initialApp,
		"alt_screen", true,
		"mouse_support", true)

	// Run the program
	finalModel, err := app.program.Run()
	
	// Ensure proper terminal cleanup regardless of how we exit
	defer func() {
		// Give terminal time to restore
		time.Sleep(50 * time.Millisecond)
	}()
	
	if err != nil {
		app.logger.LogError(err, "program_run")
		return fmt.Errorf("failed to run program: %w", err)
	}

	// Check if the model has an error
	if m, ok := finalModel.(*Model); ok && m.err != nil {
		app.logger.LogError(m.err, "model_error")
		return m.err
	}

	app.logger.Info("Application exited normally")
	return nil
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
