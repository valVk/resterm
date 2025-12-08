package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// fileTreeViewport wraps a viewport for scrollable tree navigation
type fileTreeViewport struct {
	viewport     viewport.Model
	items        []treeItem
	cursor       int
	selectedNode *treeNode
}

func newFileTreeViewport(items []treeItem) fileTreeViewport {
	vp := viewport.New(0, 0)
	return fileTreeViewport{
		viewport: vp,
		items:    items,
		cursor:   0,
	}
}

func (ftv *fileTreeViewport) SetSize(width, height int) {
	ftv.viewport.Width = width
	ftv.viewport.Height = height
	ftv.updateContent()
}

func (ftv *fileTreeViewport) updateContent() {
	if len(ftv.items) == 0 {
		ftv.viewport.SetContent("")
		return
	}

	var lines []string
	for i, item := range ftv.items {
		line := item.Title()
		if i == ftv.cursor {
			// Highlight selected line
			line = lipgloss.NewStyle().Reverse(true).Render(line)
		}
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	ftv.viewport.SetContent(content)

	// Ensure cursor is visible
	ftv.viewport.SetYOffset(ftv.cursor)
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
