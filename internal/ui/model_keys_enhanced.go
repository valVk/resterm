package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/unkn0wn-root/resterm/internal/ui/navigator"
)

// handleEditorVisibilityKeys handles the 'r' and 'e' keys for showing the editor.
// This is called from model_update_navigator.go for 'r'/'e' key handling.
// Returns the commands needed to show the editor and focus it.
func handleEditorVisibilityKeys(m *Model, key string, n *navigator.Node[any]) tea.Cmd {
	data := getEnhancedData(m)
	if data == nil {
		return nil
	}

	// Only handle 'r' and 'e' keys
	if key != "r" && key != "e" {
		return nil
	}

	// Only handle requests and workflows
	if n == nil || (n.Kind != navigator.KindRequest && n.Kind != navigator.KindWorkflow) {
		return nil
	}

	// Show editor if it was hidden
	wasHidden := !data.editorVisible
	if wasHidden {
		data.editorVisible = true
	}

	var allCmds []tea.Cmd

	// Handle based on item kind
	if n.Kind == navigator.KindRequest {
		req, _, cmds, ok := m.prepareNavigatorRequest()
		if ok {
			m.jumpToNavigatorRequest(req, true)
			if wasHidden {
				allCmds = append(allCmds, m.applyLayout())
			}
			if len(cmds) > 0 {
				allCmds = append(allCmds, cmds...)
			}
			m.setFocus(focusEditor)
			if len(allCmds) > 0 {
				return tea.Batch(allCmds...)
			}
		}
	} else if n.Kind == navigator.KindWorkflow {
		wf, _, cmds, ok := m.prepareNavigatorWorkflow()
		if ok {
			m.revealWorkflowInEditor(wf)
			if wasHidden {
				allCmds = append(allCmds, m.applyLayout())
			}
			if len(cmds) > 0 {
				allCmds = append(allCmds, cmds...)
			}
			m.setFocus(focusEditor)
			if len(allCmds) > 0 {
				return tea.Batch(allCmds...)
			}
		}
	}

	return nil
}
