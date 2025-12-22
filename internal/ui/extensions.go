package ui

import tea "github.com/charmbracelet/bubbletea"

// ExtensionHooks provides extension points for customizing the UI behavior.
// All hooks are optional (nil-safe) and called at strategic points in the application lifecycle.
// This allows users to extend functionality without modifying core files.
type ExtensionHooks struct {
	// OnModelInit is called after the model is constructed but before it's fully initialized.
	// Use this to set up any extension-specific state.
	OnModelInit func(m *Model)

	// OnUpdate is called at the very start of the Update loop, before any message handling.
	// Useful for updating extension state (e.g., animations, timers).
	OnUpdate func(m *Model, msg tea.Msg) tea.Cmd

	// OnRequestStart is called when a request execution begins.
	// Return commands that should run alongside the request (e.g., start animations).
	OnRequestStart func(m *Model) tea.Cmd

	// OnRequestEnd is called when a request completes (success or failure).
	// Useful for cleanup or post-processing.
	OnRequestEnd func(m *Model) tea.Cmd

	// StatusBarExtras returns additional segments to display in the status bar.
	// These are appended to the standard status information.
	StatusBarExtras func(m *Model) []string

	// HandleCustomKey is called for each key press before standard handling.
	// Return (true, cmd) if the key was handled, or (false, nil) to continue normal processing.
	HandleCustomKey func(m *Model, key string) (handled bool, cmd tea.Cmd)

	// OnNavigatorSelectionChange is called when the navigator selection changes.
	// This allows extensions to react to navigation changes (e.g., scroll editor to show selected request).
	OnNavigatorSelectionChange func(m *Model)
}

// Extensions holds custom extension data.
// Use this to store any state your extensions need.
type Extensions struct {
	// Data can hold any extension-specific state
	Data interface{}

	// Hooks defines the extension behavior
	Hooks *ExtensionHooks
}

// GetExtensions safely retrieves the Extensions from the model.
// Returns nil if no extensions are installed.
func (m *Model) GetExtensions() *Extensions {
	if m.extensions == nil {
		return nil
	}
	ext, ok := m.extensions.(*Extensions)
	if !ok {
		return nil
	}
	return ext
}

// SetExtensions installs extensions into the model.
func (m *Model) SetExtensions(ext *Extensions) {
	m.extensions = ext
}
