package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// StateMachine manages UI state transitions
type StateMachine struct {
	current     ViewState
	previous    ViewState
	transitions map[stateTransition]transitionHandler
	history     []ViewState
	maxHistory  int
}

// stateTransition represents a state change
type stateTransition struct {
	from ViewState
	to   ViewState
}

// transitionHandler validates and handles state transitions
type transitionHandler func() error

// NewStateMachine creates a new state machine
func NewStateMachine(initial ViewState) *StateMachine {
	sm := &StateMachine{
		current:     initial,
		previous:    initial,
		transitions: make(map[stateTransition]transitionHandler),
		history:     []ViewState{initial},
		maxHistory:  10,
	}

	// Define valid transitions
	sm.defineTransitions()

	return sm
}

// defineTransitions sets up valid state transitions
func (sm *StateMachine) defineTransitions() {
	// From ListView
	sm.addTransition(ListView, FormView, nil)
	sm.addTransition(ListView, HelpView, nil)
	sm.addTransition(ListView, ProgressView, nil)

	// From FormView
	sm.addTransition(FormView, ListView, nil)
	sm.addTransition(FormView, HelpView, nil)
	sm.addTransition(FormView, ProgressView, nil)

	// From HelpView
	sm.addTransition(HelpView, ListView, nil)
	sm.addTransition(HelpView, FormView, nil)

	// From ProgressView
	sm.addTransition(ProgressView, ListView, nil)
	sm.addTransition(ProgressView, FormView, nil)
	sm.addTransition(ProgressView, HelpView, nil)

	// Self-transitions (refresh)
	sm.addTransition(ListView, ListView, nil)
	sm.addTransition(FormView, FormView, nil)
	sm.addTransition(HelpView, HelpView, nil)
	sm.addTransition(ProgressView, ProgressView, nil)
}

// addTransition registers a valid transition
func (sm *StateMachine) addTransition(from, to ViewState, handler transitionHandler) {
	key := stateTransition{from: from, to: to}
	sm.transitions[key] = handler
}

// Transition attempts to change state
func (sm *StateMachine) Transition(to ViewState) error {
	// Check if transition is valid
	key := stateTransition{from: sm.current, to: to}
	handler, valid := sm.transitions[key]

	if !valid {
		return fmt.Errorf("invalid transition from %s to %s",
			sm.getStateName(sm.current),
			sm.getStateName(to))
	}

	// Execute transition handler if present
	if handler != nil {
		if err := handler(); err != nil {
			return fmt.Errorf("transition handler failed: %w", err)
		}
	}

	// Update state
	sm.previous = sm.current
	sm.current = to

	// Update history
	sm.addToHistory(to)

	return nil
}

// Current returns the current state
func (sm *StateMachine) Current() ViewState {
	return sm.current
}

// Previous returns the previous state
func (sm *StateMachine) Previous() ViewState {
	return sm.previous
}

// CanTransition checks if a transition is valid
func (sm *StateMachine) CanTransition(to ViewState) bool {
	key := stateTransition{from: sm.current, to: to}
	_, valid := sm.transitions[key]
	return valid
}

// Back transitions to the previous state
func (sm *StateMachine) Back() error {
	if len(sm.history) <= 1 {
		return fmt.Errorf("no previous state in history")
	}

	// Remove current from history
	sm.history = sm.history[:len(sm.history)-1]

	// Get previous state
	prev := sm.history[len(sm.history)-1]

	// Transition to previous
	return sm.Transition(prev)
}

// Reset resets to initial state
func (sm *StateMachine) Reset(initial ViewState) {
	sm.current = initial
	sm.previous = initial
	sm.history = []ViewState{initial}
}

// GetHistory returns state history
func (sm *StateMachine) GetHistory() []ViewState {
	return sm.history
}

// GetValidTransitions returns valid transitions from current state
func (sm *StateMachine) GetValidTransitions() []ViewState {
	var valid []ViewState

	for key := range sm.transitions {
		if key.from == sm.current {
			valid = append(valid, key.to)
		}
	}

	return valid
}

// HandleStateChange creates a command for state change
func (sm *StateMachine) HandleStateChange(to ViewState) tea.Cmd {
	if err := sm.Transition(to); err != nil {
		return func() tea.Msg {
			return StateChangeErrorMsg{
				From:  sm.current,
				To:    to,
				Error: err,
			}
		}
	}

	return func() tea.Msg {
		return StateChangedMsg{
			From: sm.previous,
			To:   sm.current,
		}
	}
}

// Private methods

func (sm *StateMachine) addToHistory(state ViewState) {
	sm.history = append(sm.history, state)

	// Limit history size
	if len(sm.history) > sm.maxHistory {
		sm.history = sm.history[1:]
	}
}

func (sm *StateMachine) getStateName(state ViewState) string {
	names := map[ViewState]string{
		ListView:     "ListView",
		FormView:     "FormView",
		HelpView:     "HelpView",
		ProgressView: "ProgressView",
	}

	if name, ok := names[state]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", state)
}

// Messages

// StateChangedMsg is sent when state changes successfully
type StateChangedMsg struct {
	From ViewState
	To   ViewState
}

// StateChangeErrorMsg is sent when state change fails
type StateChangeErrorMsg struct {
	From  ViewState
	To    ViewState
	Error error
}
