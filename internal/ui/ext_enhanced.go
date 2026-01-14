package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/unkn0wn-root/resterm/internal/ui/scroll"
)

// enhancedData holds all enhanced UI state for the extensions.
// This keeps enhanced data separate from the core Model struct.
type enhancedData struct {
	lastEditorCursorLine int
	needsRevealOnNextUpdate bool
}

// InstallEnhanced sets up all enhanced UI features.
// Call this from main after creating the model to enable enhanced features.
func InstallEnhanced(m *Model) {
	data := &enhancedData{}

	ext := &Extensions{
		Data: data,
		Hooks: &ExtensionHooks{
			OnUpdate:                   onUpdate,
			OnRequestStart:             onRequestStart,
			OnRequestEnd:               onRequestEnd,
			StatusBarExtras:            statusBarExtras,
			HandleCustomKey:            handleCustomKey,
			OnNavigatorSelectionChange: onNavigatorSelectionChange,
		},
	}

	m.SetExtensions(ext)

	// Hide editor by default
	m.editorCollapsed = true

	// Install typewriter mode scroll overrides
	scroll.AlignOverride = typewriterScrollAlign
	scroll.RevealOverride = typewriterScrollReveal
}

// typewriterScrollAlign implements true typewriter mode scrolling.
// The cursor stays centered in the viewport (at h/2) whenever possible.
func typewriterScrollAlign(sel, off, h, total int) (offset int, override bool) {
	if h <= 0 || total <= 0 {
		return 0, false
	}
	if sel < 0 {
		sel = 0
	}
	if sel >= total {
		sel = total - 1
	}
	if h > total {
		h = total
	}
	maxOff := total - h
	if maxOff < 0 {
		maxOff = 0
	}

	// Typewriter mode: keep cursor centered
	center := h / 2
	targetOff := sel - center

	// Clamp to valid range
	if targetOff < 0 {
		targetOff = 0
	}
	if targetOff > maxOff {
		targetOff = maxOff
	}

	return targetOff, true
}

// typewriterScrollReveal implements typewriter mode for revealing spans.
// When revealing a request, center it in the viewport instead of using the default buffer.
func typewriterScrollReveal(start, end, off, h, total int) (offset int, override bool) {
	if h <= 0 || total <= 0 {
		return 0, false
	}
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	if end >= total {
		end = total - 1
	}
	if h > total {
		h = total
	}
	maxOff := total - h
	if maxOff < 0 {
		maxOff = 0
	}

	// Typewriter mode: center the span in the viewport
	// Calculate the middle of the span
	spanMiddle := (start + end) / 2

	// Position viewport so span is centered
	center := h / 2
	targetOff := spanMiddle - center

	// Clamp to valid range
	if targetOff < 0 {
		targetOff = 0
	}
	if targetOff > maxOff {
		targetOff = maxOff
	}

	return targetOff, true
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

// onUpdate handles extension-specific updates.
// Main now handles request animations, so this is available for future extensions.
func onUpdate(m *Model, msg tea.Msg) tea.Cmd {
	data := getEnhancedData(m)
	if data == nil {
		return nil
	}

	// If we need to reveal after expanding editor, do it now if editor has proper height
	if data.needsRevealOnNextUpdate && !m.collapseState(paneRegionEditor) && m.editor.Height() > 10 {
		data.needsRevealOnNextUpdate = false
		if m.currentRequest != nil {
			// Move cursor to the request's @name line and center it in viewport
			m.moveCursorToLine(m.currentRequest.LineRange.Start)
			alignViewportToCursor(m)
		}
	}

	return nil
}

// onRequestStart is called when a request begins executing.
// Main now handles request animations, so this is available for future extensions.
func onRequestStart(m *Model) tea.Cmd {
	return nil
}

// onRequestEnd is called when a request completes.
func onRequestEnd(m *Model) tea.Cmd {
	// Cleanup if needed
	return nil
}

// statusBarExtras adds enhanced items to the status bar.
// Main now handles request status display, so this is available for future extensions.
func statusBarExtras(m *Model) []string {
	return nil
}

// alignViewportToCursor centers the viewport on the current cursor position.
// This should be called after moveCursorToLine to properly center the cursor.
func alignViewportToCursor(m *Model) {
	cursorLine := m.editor.Line() // 0-based
	h := m.editor.Height()
	if h <= 0 {
		return
	}
	total := m.editor.LineCount()
	offset := scroll.Align(cursorLine, m.editor.ViewStart(), h, total)
	m.editor.SetViewStart(offset)
}

// onNavigatorSelectionChange is called when the navigator selection changes.
// This implements Navigation → Editor sync: when navigating requests,
// the editor scrolls to show the selected request (only if editor is visible).
func onNavigatorSelectionChange(m *Model) {
	// Get the current active request
	req := m.currentRequest
	if req == nil {
		return
	}

	// Only sync if editor is visible with proper height
	// If collapsed, we'll sync when revealing with 'r' or right arrow
	if !m.collapseState(paneRegionEditor) && m.editor.Height() > 10 {
		// Move cursor to the request's @name line and center it in viewport
		m.moveCursorToLine(req.LineRange.Start)
		alignViewportToCursor(m)
	}
}

// handleCustomKey handles enhanced key bindings.
// Returns true if the key was handled.
func handleCustomKey(m *Model, key string) (bool, tea.Cmd) {
	switch key {
	case "r":
		// 'r' = reveal editor and focus it
		// If collapsed: expand it first, then focus
		// If already visible: just focus it
		var cmds []tea.Cmd
		wasCollapsed := m.collapseState(paneRegionEditor)
		if wasCollapsed {
			// Editor is collapsed, so expand it
			if cmd := m.togglePaneCollapse(paneRegionEditor); cmd != nil {
				cmds = append(cmds, cmd)
			}
			// Mark that we need to reveal on next update (after editor has proper height)
			if data := getEnhancedData(m); data != nil {
				data.needsRevealOnNextUpdate = true
			}
		}
		// Set focus to editor (whether it was collapsed or already visible)
		if cmd := m.setFocus(focusEditor); cmd != nil {
			cmds = append(cmds, cmd)
		}
		// If editor was already visible, sync cursor immediately
		if !wasCollapsed && m.currentRequest != nil {
			m.moveCursorToLine(m.currentRequest.LineRange.Start)
			alignViewportToCursor(m)
		}
		if len(cmds) > 0 {
			return true, tea.Batch(cmds...)
		}
		return true, nil
	case "right":
		// Right arrow = reveal editor and focus it when in navigator
		// This prevents conflict with navigator's right arrow usage
		if m.focus == focusRequests || m.focus == focusFile || m.focus == focusWorkflows {
			var cmds []tea.Cmd
			wasCollapsed := m.collapseState(paneRegionEditor)
			if wasCollapsed {
				// Editor is collapsed, so expand it
				if cmd := m.togglePaneCollapse(paneRegionEditor); cmd != nil {
					cmds = append(cmds, cmd)
				}
				// Mark that we need to reveal on next update (after editor has proper height)
				if data := getEnhancedData(m); data != nil {
					data.needsRevealOnNextUpdate = true
				}
			}
			// Set focus to editor (whether it was collapsed or already visible)
			if cmd := m.setFocus(focusEditor); cmd != nil {
				cmds = append(cmds, cmd)
			}
			// If editor was already visible, sync cursor immediately
			if !wasCollapsed && m.currentRequest != nil {
				m.moveCursorToLine(m.currentRequest.LineRange.Start)
				alignViewportToCursor(m)
			}
			if len(cmds) > 0 {
				return true, tea.Batch(cmds...)
			}
			return true, nil
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
