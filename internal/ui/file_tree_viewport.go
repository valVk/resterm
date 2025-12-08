package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/unkn0wn-root/resterm/internal/theme"
)

// fileTreeViewport wraps a viewport for scrollable tree navigation
type fileTreeViewport struct {
	viewport     viewport.Model
	items        []treeItem
	cursor       int
	selectedNode *treeNode
	theme        theme.Theme
}

func newFileTreeViewport(items []treeItem, th theme.Theme) fileTreeViewport {
	vp := viewport.New(0, 0)
	return fileTreeViewport{
		viewport: vp,
		items:    items,
		cursor:   0,
		theme:    th,
	}
}

func (ftv *fileTreeViewport) SetSize(width, height int) {
	ftv.viewport.Width = width
	ftv.viewport.Height = height
	ftv.updateContent()
}

func (ftv *fileTreeViewport) getMethodColor(method string) lipgloss.Color {
	// Swagger/Darcula-like colors for HTTP methods
	switch strings.ToUpper(method) {
	case "GET":
		return lipgloss.Color("#61AFFE") // Blue
	case "POST":
		return lipgloss.Color("#49CC90") // Green
	case "PUT":
		return lipgloss.Color("#FCA130") // Orange
	case "PATCH":
		return lipgloss.Color("#50E3C2") // Emerald/Turquoise
	case "DELETE":
		return lipgloss.Color("#F93E3E") // Red
	default:
		return lipgloss.Color("#FFFFFF") // White
	}
}

func (ftv *fileTreeViewport) updateContent() {
	if len(ftv.items) == 0 {
		ftv.viewport.SetContent("")
		return
	}

	var lines []string
	for i, item := range ftv.items {
		node := item.node
		indent := strings.Repeat(" ", node.depth)
		var line string

		switch node.nodeType {
		case treeNodeDir:
			prefix := "▶ "
			if node.expanded {
				prefix = "▼ "
			}
			line = indent + prefix + node.name

		case treeNodeFile:
			prefix := "▶ "
			if node.expanded {
				prefix = "▼ "
			}
			line = indent + prefix + node.name

		case treeNodeRequest:
			// Request with colored HTTP method
			if node.request != nil {
				method := strings.ToUpper(strings.TrimSpace(node.request.Method))
				if method == "" {
					method = "REQ"
				}
				methodColor := ftv.getMethodColor(method)
				methodStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(methodColor)

				// Method in color + bold, name in regular
				line = indent + "  " + methodStyle.Render(method) + " " + node.name
			} else {
				line = indent + "  " + node.name
			}
		}

		if i == ftv.cursor {
			// Highlight selected line
			line = lipgloss.NewStyle().Reverse(true).Render(line)
		}

		lines = append(lines, line)

		// Add spacing after directories and files (not after requests)
		if node.nodeType == treeNodeDir || node.nodeType == treeNodeFile {
			lines = append(lines, "")
		}
	}

	content := strings.Join(lines, "\n")
	ftv.viewport.SetContent(content)

	// Typewriter mode: keep cursor centered (or near center)
	// Calculate actual line position accounting for spacing
	actualLine := ftv.cursor
	for i := 0; i < ftv.cursor && i < len(ftv.items); i++ {
		node := ftv.items[i].node
		if node.nodeType == treeNodeDir || node.nodeType == treeNodeFile {
			actualLine++ // Add 1 for spacing line
		}
	}

	// Keep cursor in center (typewriter mode)
	centerOffset := ftv.viewport.Height / 2
	targetOffset := actualLine - centerOffset
	if targetOffset < 0 {
		targetOffset = 0
	}
	ftv.viewport.SetYOffset(targetOffset)
}

func (ftv *fileTreeViewport) SetItems(items []treeItem) {
	ftv.items = items
	if ftv.cursor >= len(items) {
		ftv.cursor = len(items) - 1
	}
	if ftv.cursor < 0 {
		ftv.cursor = 0
	}
	ftv.updateContent()
}

func (ftv *fileTreeViewport) Select(index int) {
	if index < 0 || index >= len(ftv.items) {
		return
	}
	ftv.cursor = index
	ftv.updateContent()
}

func (ftv *fileTreeViewport) SelectedItem() *treeNode {
	if ftv.cursor < 0 || ftv.cursor >= len(ftv.items) {
		return nil
	}
	return ftv.items[ftv.cursor].node
}

func (ftv *fileTreeViewport) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if ftv.cursor > 0 {
				ftv.cursor--
				ftv.updateContent()
			}
			return nil
		case "down", "j":
			if ftv.cursor < len(ftv.items)-1 {
				ftv.cursor++
				ftv.updateContent()
			}
			return nil
		case "g":
			ftv.cursor = 0
			ftv.updateContent()
			return nil
		case "G":
			ftv.cursor = len(ftv.items) - 1
			ftv.updateContent()
			return nil
		}
	}

	var cmd tea.Cmd
	ftv.viewport, cmd = ftv.viewport.Update(msg)
	return cmd
}

func (ftv *fileTreeViewport) View() string {
	return ftv.viewport.View()
}
