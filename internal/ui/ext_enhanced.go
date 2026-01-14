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

	// If we need to align viewport after expanding editor, keep trying until we succeed
	// Note: Cursor was already moved in handleCustomKey before focus was set
	if data.needsRevealOnNextUpdate {
		// Check if editor is now visible and has proper dimensions
		if !m.collapseState(paneRegionEditor) && m.editor.Height() > 5 {
			// Just align the viewport - cursor was already moved in handleCustomKey
			alignViewportToCursor(m)
			// Clear flag only after successful align
			data.needsRevealOnNextUpdate = false
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

// moveCursorToLineNoSync moves the editor cursor to the target line
// WITHOUT syncing the navigator. This is used when the navigator is already
// at the correct position and we just need to move the editor cursor to match.
func moveCursorToLineNoSync(m *Model, target int) {
	if target < 1 {
		target = 1
	}
	total := m.editor.LineCount()
	if total == 0 {
		return
	}
	if target > total {
		target = total
	}
	current := m.editor.Line() + 1 // editor.Line() is 0-based
	if current == target {
		return
	}

	wasFocused := m.editor.Focused()
	if !wasFocused {
		_ = m.editor.Focus()
	}
	defer func() {
		if !wasFocused {
			m.editor.Blur()
		}
	}()

	for current < target {
		m.editor, _ = m.editor.Update(tea.KeyMsg{Type: tea.KeyDown})
		current++
	}
	for current > target {
		m.editor, _ = m.editor.Update(tea.KeyMsg{Type: tea.KeyUp})
		current--
	}
	m.editor, _ = m.editor.Update(tea.KeyMsg{Type: tea.KeyHome})
	// NOTE: We deliberately don't call syncNavigatorWithEditorCursor() here
	// because the navigator is already at the correct position
}

// alignViewportToCursor centers the viewport on the current cursor position.
// This should be called after moveCursorToLineNoSync to properly center the cursor.
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

	data := getEnhancedData(m)
	if data == nil {
		return
	}

	// If editor is visible with proper height, sync immediately
	if !m.collapseState(paneRegionEditor) && m.editor.Height() > 10 {
		// Move cursor to the request's @name line and center it in viewport
		// Use NoSync version because navigator is already at correct position
		moveCursorToLineNoSync(m, req.LineRange.Start)
		alignViewportToCursor(m)
	} else {
		// Editor is collapsed, mark that we need to sync when it's revealed
		data.needsRevealOnNextUpdate = true
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
		}

		// CRITICAL: Move cursor BEFORE setting focus to prevent main code from
		// syncing navigator to old editor cursor position (model_update.go:506)
		if m.currentRequest != nil {
			moveCursorToLineNoSync(m, m.currentRequest.LineRange.Start)
			// Don't align viewport yet if editor was collapsed - height might be wrong
			// We'll align in onUpdate once editor has proper dimensions
			if !wasCollapsed {
				alignViewportToCursor(m)
			} else {
				// Mark that we need to align viewport on next update
				if data := getEnhancedData(m); data != nil {
					data.needsRevealOnNextUpdate = true
				}
			}
		}

		// Set focus to editor (whether it was collapsed or already visible)
		if cmd := m.setFocus(focusEditor); cmd != nil {
			cmds = append(cmds, cmd)
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
			}

			// CRITICAL: Move cursor BEFORE setting focus to prevent main code from
			// syncing navigator to old editor cursor position (model_update.go:506)
			if m.currentRequest != nil {
				moveCursorToLineNoSync(m, m.currentRequest.LineRange.Start)
				// Don't align viewport yet if editor was collapsed - height might be wrong
				// We'll align in onUpdate once editor has proper dimensions
				if !wasCollapsed {
					alignViewportToCursor(m)
				} else {
					// Mark that we need to align viewport on next update
					if data := getEnhancedData(m); data != nil {
						data.needsRevealOnNextUpdate = true
					}
				}
			}

			// Set focus to editor (whether it was collapsed or already visible)
			if cmd := m.setFocus(focusEditor); cmd != nil {
				cmds = append(cmds, cmd)
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
