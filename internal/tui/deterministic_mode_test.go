package tui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrtkrcm/ZeroUI/internal/appconfig"
	"github.com/mrtkrcm/ZeroUI/internal/logger"
	"github.com/mrtkrcm/ZeroUI/internal/service"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
	"github.com/mrtkrcm/ZeroUI/internal/validation"
)

// TestDeterministicSpinner ensures spinner frame is stable in ZEROUI_TEST_MODE
func TestDeterministicSpinner(t *testing.T) {
	t.Setenv("ZEROUI_TEST_MODE", "true")

	log := logger.Global()
	configLoader, err := appconfig.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	validator := validation.NewValidator()
	engine := toggle.NewEngineWithDeps(configLoader, log, validator)
	configService := service.NewConfigService(engine, configLoader, log)

	model, err := NewTestModel(configService, "ghostty")
	require.NoError(t, err)

	// Enter loading state
	model.state = FormView
	model.isLoading = true
	model.loadingText = "Loading configuration for test..."

	v1 := model.renderLoadingState()
	time.Sleep(50 * time.Millisecond)
	v2 := model.renderLoadingState()

	assert.Equal(t, v1, v2, "Spinner frame should be deterministic in test mode")
}

// TestStatusToast ensures transient status is appended and then expires
func TestStatusToast(t *testing.T) {
	log := logger.Global()
	configLoader, err := appconfig.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	validator := validation.NewValidator()
	engine := toggle.NewEngineWithDeps(configLoader, log, validator)
	configService := service.NewConfigService(engine, configLoader, log)

	model, err := NewTestModel(configService, "")
	require.NoError(t, err)

	model.width = 100
	model.height = 30

	// Inject a success status
	model.statusText = "Saved ✓ (ghostty.conf)"
	model.statusLevel = 1 // util.InfoTypeSuccess
	model.statusUntil = time.Now().Add(50 * time.Millisecond)

	viewWithStatus := model.View()
	assert.Contains(t, viewWithStatus, "Saved ✓", "Status line should be visible initially")

	// After expiry, status should disappear
	time.Sleep(60 * time.Millisecond)
	viewAfter := model.View()
	assert.NotContains(t, viewAfter, "Saved ✓", "Status line should expire and disappear")
}

// TestRefreshDebounce ensures rapid refresh messages are debounced
func TestRefreshDebounce(t *testing.T) {
	log := logger.Global()
	configLoader, err := appconfig.NewReferenceEnhancedLoader()
	require.NoError(t, err)
	validator := validation.NewValidator()
	engine := toggle.NewEngineWithDeps(configLoader, log, validator)
	configService := service.NewConfigService(engine, configLoader, log)

	model, err := NewTestModel(configService, "")
	require.NoError(t, err)

	// First refresh should set timestamp
	_, _ = model.Update(RefreshAppsMsg{})
	first := model.lastAppsRefresh
	assert.False(t, first.IsZero())

	// Immediate second refresh should be ignored (timestamp unchanged)
	_, _ = model.Update(RefreshAppsMsg{})
	second := model.lastAppsRefresh
	assert.Equal(t, first, second, "Second refresh too soon should be ignored")

	// After debounce window, refresh should update timestamp
	time.Sleep(310 * time.Millisecond)
	_, _ = model.Update(RefreshAppsMsg{})
	third := model.lastAppsRefresh
	assert.Greater(t, third.UnixNano(), second.UnixNano())
}
