package ui

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"

	"github.com/unkn0wn-root/resterm/internal/filesvc"
	"github.com/unkn0wn-root/resterm/internal/restfile"
)

type treeNodeType int

const (
	treeNodeDir treeNodeType = iota
	treeNodeFile
	treeNodeRequest
)

type treeNode struct {
	nodeType     treeNodeType
	name         string
	path         string
	depth        int
	expanded     bool
	parent       *treeNode
	children     []*treeNode
	request      *restfile.Request
	requestIndex int
}

type fileTree struct {
	root          *treeNode
	flatList      []*treeNode
	expandedPaths map[string]bool
	focusStack    []string
}

func newFileTree(root string) *fileTree {
	return &fileTree{
		root: &treeNode{
			nodeType: treeNodeDir,
			name:     filepath.Base(root),
			path:     root,
			depth:    0,
			expanded: true,
			children: []*treeNode{},
		},
		expandedPaths: make(map[string]bool),
		focusStack:    []string{},
	}
}

func (t *fileTree) buildFromFiles(entries []filesvc.FileEntry, root string) {
	dirMap := make(map[string]*treeNode)
	dirMap[root] = t.root

	for _, entry := range entries {
		relPath, err := filepath.Rel(root, entry.Path)
		if err != nil {
			relPath = entry.Name
		}

		parts := strings.Split(filepath.ToSlash(relPath), "/")
		currentPath := root
		parentNode := t.root

		for i, part := range parts {
			currentPath = filepath.Join(currentPath, part)
			isLast := i == len(parts)-1

			if isLast {
				fileNode := &treeNode{
					nodeType: treeNodeFile,
					name:     part,
					path:     entry.Path,
					depth:    i + 1,
					expanded: false,
					parent:   parentNode,
					children: []*treeNode{},
				}
				parentNode.children = append(parentNode.children, fileNode)
			} else {
				if _, exists := dirMap[currentPath]; !exists {
					dirNode := &treeNode{
						nodeType: treeNodeDir,
						name:     part,
						path:     currentPath,
						depth:    i + 1,
						expanded: t.expandedPaths[currentPath],
						parent:   parentNode,
						children: []*treeNode{},
					}
					dirMap[currentPath] = dirNode
					parentNode.children = append(parentNode.children, dirNode)
					parentNode = dirNode
				} else {
					parentNode = dirMap[currentPath]
				}
			}
		}
	}

	t.sortChildren(t.root)
}

func (t *fileTree) sortChildren(node *treeNode) {
	sort.Slice(node.children, func(i, j int) bool {
		if node.children[i].nodeType != node.children[j].nodeType {
			return node.children[i].nodeType < node.children[j].nodeType
		}
		return node.children[i].name < node.children[j].name
	})

	for _, child := range node.children {
		t.sortChildren(child)
	}
}

func (t *fileTree) addRequestsToFile(filePath string, requests []requestListItem) {
	node := t.findNodeByPath(t.root, filePath)
	if node == nil || node.nodeType != treeNodeFile {
		return
	}

	node.children = make([]*treeNode, 0, len(requests))
	for i, req := range requests {
		reqNode := &treeNode{
			nodeType:     treeNodeRequest,
			name:         req.Title(),
			path:         filePath,
			depth:        node.depth + 1,
			expanded:     false,
			parent:       node,
			children:     []*treeNode{},
			request:      req.request,
			requestIndex: i,
		}
		node.children = append(node.children, reqNode)
	}
}

func (t *fileTree) findNodeByPath(node *treeNode, path string) *treeNode {
	if node.path == path {
		return node
	}
	for _, child := range node.children {
		if found := t.findNodeByPath(child, path); found != nil {
			return found
		}
	}
	return nil
}

func (t *fileTree) flatten() []*treeNode {
	var result []*treeNode
	t.flattenNode(t.root, &result, false)
	t.flatList = result
	return result
}

func (t *fileTree) flattenNode(node *treeNode, result *[]*treeNode, skipRoot bool) {
	if !skipRoot {
		*result = append(*result, node)
	}

	if node.expanded || skipRoot {
		for _, child := range node.children {
			t.flattenNode(child, result, false)
		}
	}
}

func (t *fileTree) toggleExpand(node *treeNode) {
	if node.nodeType == treeNodeRequest {
		return
	}
	node.expanded = !node.expanded
	if node.nodeType == treeNodeDir {
		t.expandedPaths[node.path] = node.expanded
	}
}

func (t *fileTree) collapseNode(node *treeNode) {
	if node.nodeType == treeNodeRequest {
		return
	}
	node.expanded = false
	if node.nodeType == treeNodeDir {
		t.expandedPaths[node.path] = false
	}
}

func (t *fileTree) expandNode(node *treeNode) {
	if node.nodeType == treeNodeRequest {
		return
	}
	node.expanded = true
	if node.nodeType == treeNodeDir {
		t.expandedPaths[node.path] = true
	}
}

type treeItem struct {
	node *treeNode
}

func (ti treeItem) Title() string {
	indent := strings.Repeat(" ", ti.node.depth)
	prefix := ""
	name := ti.node.name

	switch ti.node.nodeType {
	case treeNodeDir:
		if ti.node.expanded {
			prefix = "▼ "
		} else {
			prefix = "▶ "
		}
	case treeNodeFile:
		if ti.node.expanded {
			prefix = "▼ "
		} else {
			prefix = "▶ "
		}
	case treeNodeRequest:
		prefix = "  "
		// For requests, show name in title with method prefix
		if ti.node.request != nil {
			method := strings.ToUpper(strings.TrimSpace(ti.node.request.Method))
			if method == "" {
				method = "REQ"
			}
			name = method + " " + name
		}
	}

	return indent + prefix + name
}

func (ti treeItem) Description() string {
	switch ti.node.nodeType {
	case treeNodeDir:
		return ""
	case treeNodeFile:
		return ""
	case treeNodeRequest:
		// Show URL in description with matching indentation
		if ti.node.request != nil {
			return strings.Repeat(" ", ti.node.depth) + "  " + ti.node.request.URL
		}
	}
	return ""
}

func (ti treeItem) FilterValue() string {
	return ti.node.name
}

func makeTreeItems(nodes []*treeNode) []list.Item {
	items := make([]list.Item, len(nodes))
	for i, node := range nodes {
		items[i] = treeItem{node: node}
	}
	return items
}

func makeTreeItemsArray(nodes []*treeNode) []treeItem {
	items := make([]treeItem, len(nodes))
	for i, node := range nodes {
		items[i] = treeItem{node: node}
	}
	return items
}

func selectedTreeNode(it list.Item) *treeNode {
	if it == nil {
		return nil
	}
	if ti, ok := it.(treeItem); ok {
		return ti.node
	}
	return nil
}
