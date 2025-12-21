package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EnhancedExtensionData holds all enhanced UI state for the extensions.
// This keeps enhanced data separate from the core Model struct.
type EnhancedExtensionData struct {
	editorVisible        bool
	requestSpinner       spinner.Model
	lastEditorCursorLine int
}

// InstallEnhancedExtensions sets up all enhanced UI features.
// Call this from main or CLI initialization code to enable enhanced features.
func InstallEnhancedExtensions(m *Model) {
	data := &EnhancedExtensionData{
		editorVisible:  false,
		requestSpinner: createRequestSpinner(),
	}

	ext := &Extensions{
		Data: data,
		Hooks: &ExtensionHooks{
			OnModelInit:        onModelInit,
			OnUpdate:           onUpdate,
			OnRequestStart:     onRequestStart,
			OnRequestEnd:       onRequestEnd,
			StatusBarExtras:    statusBarExtras,
			ShouldSkipEditor:   shouldSkipEditor,
			CustomLayoutAdjust: customLayoutAdjust,
			HandleCustomKey:    handleCustomKey,
		},
	}

	m.SetExtensions(ext)
}

// createRequestSpinner initializes the spinner used during request execution.
func createRequestSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return s
}

// getEnhancedData safely retrieves the enhanced extension data from the model.
func getEnhancedData(m *Model) *EnhancedExtensionData {
	ext := m.GetExtensions()
	if ext == nil || ext.Data == nil {
		return nil
	}
	data, ok := ext.Data.(*EnhancedExtensionData)
	if !ok {
		return nil
	}
	return data
}

// onModelInit is called after the model is created.
func onModelInit(m *Model) {
	// Any initialization logic can go here
	// Currently the data is already initialized in InstallEnhancedExtensions
}

// onUpdate handles spinner animation during request execution.
func onUpdate(m *Model, msg tea.Msg) tea.Cmd {
	data := getEnhancedData(m)
	if data == nil || !m.sending {
		return nil
	}

	// Update spinner animation
	var cmd tea.Cmd
	data.requestSpinner, cmd = data.requestSpinner.Update(msg)
	return cmd
}

// onRequestStart is called when a request begins executing.
// Returns the spinner tick command to start animation.
func onRequestStart(m *Model) tea.Cmd {
	data := getEnhancedData(m)
	if data == nil {
		return nil
	}

	// Start spinner animation
	return data.requestSpinner.Tick
}

// onRequestEnd is called when a request completes.
func onRequestEnd(m *Model) tea.Cmd {
	// Cleanup if needed
	return nil
}

// statusBarExtras adds enhanced items to the status bar.
func statusBarExtras(m *Model) []string {
	data := getEnhancedData(m)
	if data == nil {
		return nil
	}

	var extras []string

	// Show spinner when request is in progress
	if m.sending {
		spinnerText := data.requestSpinner.View() + " Sending request"
		extras = append(extras, spinnerText)
	}

	return extras
}

// shouldSkipEditor returns true if the editor should be skipped during focus cycling.
// This implements the editor visibility feature.
func shouldSkipEditor(m *Model) bool {
	data := getEnhancedData(m)
	if data == nil {
		return false
	}
	return !data.editorVisible
}

// customLayoutAdjust modifies editor dimensions when the editor is hidden.
func customLayoutAdjust(m *Model, defaultWidth, defaultHeight int) (width, height int, override bool) {
	data := getEnhancedData(m)
	if data == nil {
		return 0, 0, false
	}

	// If editor is hidden, set dimensions to 0
	if !data.editorVisible {
		return 0, 0, true
	}

	// Use default dimensions
	return 0, 0, false
}

// handleCustomKey handles enhanced key bindings.
// Returns true if the key was handled.
func handleCustomKey(m *Model, key string) (bool, tea.Cmd) {
	// Note: This receives the key and returns whether it was handled.
	// The actual logic for 'r' and 'e' keys is in model_update_navigator.go
	// where it has access to the navigator state.
	// This hook is here for potential future enhanced key handling.
	return false, nil
}

// SetEditorVisible exposes the editor visibility control.
// This can be called from anywhere that has access to the model.
func SetEditorVisible(m *Model, visible bool) {
	data := getEnhancedData(m)
	if data == nil {
		return
	}
	data.editorVisible = visible
}

// IsEditorVisible returns the current editor visibility state.
func IsEditorVisible(m *Model) bool {
	data := getEnhancedData(m)
	if data == nil {
		return true // Default to visible if no extensions
	}
	return data.editorVisible
}

// GetLastEditorCursorLine returns the last tracked editor cursor line.
func GetLastEditorCursorLine(m *Model) int {
	data := getEnhancedData(m)
	if data == nil {
		return 0
	}
	return data.lastEditorCursorLine
}

// SetLastEditorCursorLine updates the last tracked editor cursor line.
func SetLastEditorCursorLine(m *Model, line int) {
	data := getEnhancedData(m)
	if data == nil {
		return
	}
	data.lastEditorCursorLine = line
}
