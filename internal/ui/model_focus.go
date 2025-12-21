package ui

import (
	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) cycleFocus(forward bool) tea.Cmd {
	// Check if extension wants to skip editor
	shouldSkip := false
	if ext := m.GetExtensions(); ext != nil && ext.Hooks != nil && ext.Hooks.ShouldSkipEditor != nil {
		shouldSkip = ext.Hooks.ShouldSkipEditor(m)
	}

	switch m.focus {
	case focusFile, focusRequests, focusWorkflows:
		if forward {
			// Skip editor if extension says so
			if shouldSkip {
				return m.setFocus(focusResponse)
			}
			return m.setFocus(focusEditor)
		} else {
			return m.setFocus(focusResponse)
		}
	case focusEditor:
		if forward {
			return m.setFocus(focusResponse)
		} else {
			return m.setFocus(focusRequests)
		}
	case focusResponse:
		if forward {
			return m.setFocus(focusRequests)
		} else {
			// Skip editor if extension says so
			if shouldSkip {
				return m.setFocus(focusRequests)
			}
			return m.setFocus(focusEditor)
		}
	}
	return nil
}

func (m *Model) setFocus(target paneFocus) tea.Cmd {
	var cmds []tea.Cmd
	if m.focus == target {
		return nil
	}
	prev := m.focus
	m.focus = target
	if target != focusResponse {
		m.responsePaneChord = false
	}
	if target == focusEditor {
		if m.editorInsertMode {
			// Adopt main's command batching pattern
			if cmd := m.editor.Cursor.SetMode(cursor.CursorBlink); cmd != nil {
				cmds = append(cmds, cmd)
			}
			if cmd := m.editor.Cursor.Focus(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		} else {
			if cmd := m.editor.Cursor.SetMode(cursor.CursorStatic); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		if cmd := m.editor.Focus(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	} else {
		if prev == focusEditor && m.editorInsertMode {
			if cmd := m.setInsertMode(false, false); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		m.editor.Blur()
	}
	if target == focusResponse {
		m.ensurePaneFocusValid()
		m.setLivePane(m.responsePaneFocus)
	}
	return batchCommands(cmds...)
}

func (m *Model) setInsertMode(enabled bool, announce bool) tea.Cmd {
	var cmds []tea.Cmd
	if enabled == m.editorInsertMode {
		return nil
	}
	m.editorInsertMode = enabled
	if enabled {
		m.editor.SetMotionsEnabled(false)
		m.editor.KeyMap = m.editorWriteKeyMap
		// Adopt main's command batching pattern
		if cmd := m.editor.Cursor.SetMode(cursor.CursorBlink); cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.editor.Cursor.Blink = true
		if cmd := m.editor.Cursor.Focus(); cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.editor.SetMetadataHintsEnabled(true)
		// Mode is shown in status bar, no need for temporary message
	} else {
		m.editor.ClearSelection()
		m.editor.SetMotionsEnabled(true)
		m.editor.KeyMap = m.editorViewKeyMap
		if cmd := m.editor.Cursor.SetMode(cursor.CursorStatic); cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.editor.Cursor.Blink = false
		m.editor.SetMetadataHintsEnabled(false)
		// Mode is shown in status bar, no need for temporary message
	}
	m.editor.undoCoalescing = false
	return batchCommands(cmds...)
}
