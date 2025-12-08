package ui

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) activatePrevSidebarTab() tea.Cmd {
	tabs := m.availableSidebarTabs()
	idx := indexOfSidebarTab(tabs, m.activeSidebarTab)
	if idx == -1 {
		m.activeSidebarTab = tabs[0]
	} else {
		idx = (idx - 1 + len(tabs)) % len(tabs)
		m.activeSidebarTab = tabs[idx]
	}
	return nil
}

func (m *Model) activateNextSidebarTab() tea.Cmd {
	tabs := m.availableSidebarTabs()
	idx := indexOfSidebarTab(tabs, m.activeSidebarTab)
	if idx == -1 {
		m.activeSidebarTab = tabs[0]
	} else {
		idx = (idx + 1) % len(tabs)
		m.activeSidebarTab = tabs[idx]
	}
	return nil
}

func indexOfSidebarTab(tabs []sidebarTab, target sidebarTab) int {
	for i, tab := range tabs {
		if tab == target {
			return i
		}
	}
	return -1
}

func (m *Model) availableSidebarTabs() []sidebarTab {
	tabs := []sidebarTab{sidebarTabFiles, sidebarTabRequests}
	if len(m.workflowItems) > 0 {
		tabs = append(tabs, sidebarTabWorkflows)
	}
	return tabs
}

func (m *Model) sidebarTabLabel(tab sidebarTab) string {
	switch tab {
	case sidebarTabFiles:
		return "Files"
	case sidebarTabRequests:
		return "Requests"
	case sidebarTabWorkflows:
		return "Workflows"
	default:
		return "?"
	}
}

func (m *Model) setSidebarTab(tab sidebarTab) {
	m.activeSidebarTab = tab
	switch tab {
	case sidebarTabFiles:
		m.setFocus(focusFile)
	case sidebarTabRequests:
		m.setFocus(focusRequests)
	case sidebarTabWorkflows:
		m.setFocus(focusWorkflows)
	}
}
