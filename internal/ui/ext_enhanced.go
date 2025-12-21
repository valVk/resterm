package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// enhancedData holds all enhanced UI state for the extensions.
// This keeps enhanced data separate from the core Model struct.
type enhancedData struct {
	requestSpinner       spinner.Model
	lastEditorCursorLine int
}

// InstallEnhanced sets up all enhanced UI features.
// Call this from main after creating the model to enable enhanced features.
func InstallEnhanced(m *Model) {
	data := &enhancedData{
		requestSpinner: createRequestSpinner(),
	}

	ext := &Extensions{
		Data: data,
		Hooks: &ExtensionHooks{
			OnUpdate:        onUpdate,
			OnRequestStart:  onRequestStart,
			OnRequestEnd:    onRequestEnd,
			StatusBarExtras: statusBarExtras,
			HandleCustomKey: handleCustomKey,
		},
	}

	m.SetExtensions(ext)

	// Hide editor by default
	m.editorCollapsed = true
}

// createRequestSpinner initializes the spinner used during request execution.
func createRequestSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return s
}

// getEnhancedData safely retrieves the enhanced extension data from the model.
func getEnhancedData(m *Model) *enhancedData {
	ext := m.GetExtensions()
	if ext == nil || ext.Data == nil {
		return nil
	}
	data, ok := ext.Data.(*enhancedData)
	if !ok {
		return nil
	}
	return data
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

// handleCustomKey handles enhanced key bindings.
// Returns true if the key was handled.
func handleCustomKey(m *Model, key string) (bool, tea.Cmd) {
	switch key {
	case "r":
		// 'r' = reveal editor (expand it using the existing collapse system)
		if m.collapseState(paneRegionEditor) {
			// Editor is collapsed, so expand it
			return true, m.togglePaneCollapse(paneRegionEditor)
		}
	case "right":
		// Right arrow = reveal editor (expand it) when:
		// 1. Editor is collapsed
		// 2. Focus is on navigator (not on editor itself)
		// This prevents conflict with navigator's right arrow usage
		if m.collapseState(paneRegionEditor) &&
		   (m.focus == focusRequests || m.focus == focusFile || m.focus == focusWorkflows) {
			// Editor is collapsed and we're in navigator, so expand it
			return true, m.togglePaneCollapse(paneRegionEditor)
		}
	case "q":
		// 'q' = quit/hide editor (collapse it using the existing collapse system)
		// Only when editor is focused
		if m.focus == focusEditor && !m.collapseState(paneRegionEditor) {
			// Editor is visible and focused, so collapse it and move focus to requests
			var cmds []tea.Cmd
			if cmd := m.togglePaneCollapse(paneRegionEditor); cmd != nil {
				cmds = append(cmds, cmd)
			}
			if cmd := m.setFocus(focusRequests); cmd != nil {
				cmds = append(cmds, cmd)
			}
			if len(cmds) > 0 {
				return true, tea.Batch(cmds...)
			}
			return true, nil
		}
	}

	return false, nil
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
