package ui

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) refreshFileTree() {
	if m.fileTree == nil {
		return
	}
	nodes := m.fileTree.flatten()
	items := makeTreeItems(nodes)
	m.fileList.SetItems(items)
}

func (m *Model) expandOrOpenTreeNode() tea.Cmd {
	node := selectedTreeNode(m.fileList.SelectedItem())
	if node == nil {
		return nil
	}

	switch node.nodeType {
	case treeNodeDir:
		m.fileTree.toggleExpand(node)
		m.refreshFileTree()
		return nil

	case treeNodeFile:
		if len(node.children) > 0 {
			m.fileTree.toggleExpand(node)
			m.refreshFileTree()
			return nil
		}
		cmd := m.openFile(node.path)
		if cmd == nil && len(m.requestItems) > 0 {
			m.fileTree.addRequestsToFile(node.path, m.requestItems)
			m.fileTree.expandNode(node)
			m.refreshFileTree()
		}
		return cmd

	case treeNodeRequest:
		m.loadRequestFromTree(node)
		return m.sendRequestFromList(true)
	}

	return nil
}

func (m *Model) collapseOrCloseTreeNode() tea.Cmd {
	node := selectedTreeNode(m.fileList.SelectedItem())
	if node == nil {
		return nil
	}

	if node.expanded && len(node.children) > 0 {
		m.fileTree.collapseNode(node)
		m.refreshFileTree()
		return nil
	}

	if node.parent != nil && node.parent != m.fileTree.root {
		currentIndex := m.fileList.Index()
		m.fileTree.collapseNode(node.parent)
		m.refreshFileTree()

		nodes := m.fileTree.flatList
		for i, n := range nodes {
			if n == node.parent {
				if i < currentIndex {
					m.fileList.Select(i)
				}
				break
			}
		}
	}

	return nil
}

func (m *Model) loadRequestFromTree(node *treeNode) {
	if node.nodeType != treeNodeRequest || node.request == nil {
		return
	}

	m.currentRequest = node.request
	m.activeRequestTitle = node.name
	m.activeRequestKey = requestKey(node.request)

	if node.parent != nil && node.parent.nodeType == treeNodeFile {
		if m.currentFile != node.parent.path {
			m.openFile(node.parent.path)
		}
	}

	if node.requestIndex >= 0 && node.requestIndex < len(m.requestItems) {
		m.requestList.Select(node.requestIndex)
		m.syncEditorWithRequestSelection(-1)
	}
}
