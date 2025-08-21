package view

import (
	"rel8/model"

	"github.com/rivo/tview"
)

type View struct {
	stateManager *model.ContextualStateManager
	model        model.State
	App          *tview.Application
	flex         *tview.Flex
	header       *Header
	tree         *Tree
	grid         *Grid
	details      *Detail
	editor       *Editor
	commandBar   *CommandBar
}

func NewView(stateManager *model.ContextualStateManager) *View {
	app := tview.NewApplication()

	// Create components
	header := NewHeader()
	tree := NewTree()
	details := NewEmptyDetail()
	editor := NewEmptyEditor()

	grid := NewEmptyGrid()

	// Create command bar (initially hidden)
	commandBar := NewCommandBar()

	// Set the server in components
	header.SetServer(stateManager.GetServer())
	tree.SetServer(stateManager.GetServer())

	// Create layout with tree as default instead of grid
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(WrapHeader(header), 7, 0, false). // Fixed header height with padding
		AddItem(WrapTree(tree), 0, 1, true)       // Tree view takes remaining space

	view := &View{
		stateManager: stateManager,
		model:        model.Initial,
		App:          app,
		flex:         flex,
		header:       header,
		tree:         tree,
		grid:         grid,
		details:      details,
		editor:       editor,
		commandBar:   commandBar,
	}

	return view
}

// findNextNode finds the next visible node in tree traversal order
func (v *View) findNextNode(current *tview.TreeNode) *tview.TreeNode {
	if current == nil {
		return nil
	}

	// If current node is expanded and has children, go to first child
	if current.IsExpanded() && len(current.GetChildren()) > 0 {
		return current.GetChildren()[0]
	}

	// Otherwise, find next sibling or go up to parent's next sibling
	return v.findNextSibling(current)
}

// findPrevNode finds the previous visible node in tree traversal order
func (v *View) findPrevNode(current *tview.TreeNode) *tview.TreeNode {
	if current == nil {
		return nil
	}

	// Find previous sibling
	prevSibling := v.findPrevSibling(current)
	if prevSibling != nil {
		// If previous sibling is expanded, go to its last descendant
		return v.findLastDescendant(prevSibling)
	}

	// No previous sibling, go to parent
	parent := v.findParent(current)
	return parent
}

// findNextSibling finds the next sibling or ancestor's next sibling
func (v *View) findNextSibling(node *tview.TreeNode) *tview.TreeNode {
	parent := v.findParent(node)
	if parent == nil {
		return nil
	}

	children := parent.GetChildren()
	for i, child := range children {
		if child == node && i+1 < len(children) {
			return children[i+1]
		}
	}

	// No next sibling, try parent's next sibling
	return v.findNextSibling(parent)
}

// findPrevSibling finds the previous sibling
func (v *View) findPrevSibling(node *tview.TreeNode) *tview.TreeNode {
	parent := v.findParent(node)
	if parent == nil {
		return nil
	}

	children := parent.GetChildren()
	for i, child := range children {
		if child == node && i > 0 {
			return children[i-1]
		}
	}

	return nil
}

// findParent finds the parent of a node by searching the tree
func (v *View) findParent(target *tview.TreeNode) *tview.TreeNode {
	root := v.tree.treeView.GetRoot()
	return v.findParentRecursive(root, target)
}

// findParentRecursive recursively searches for the parent of target
func (v *View) findParentRecursive(parent *tview.TreeNode, target *tview.TreeNode) *tview.TreeNode {
	if parent == nil {
		return nil
	}

	for _, child := range parent.GetChildren() {
		if child == target {
			return parent
		}
		if found := v.findParentRecursive(child, target); found != nil {
			return found
		}
	}

	return nil
}

// findLastDescendant finds the last visible descendant of a node
func (v *View) findLastDescendant(node *tview.TreeNode) *tview.TreeNode {
	if node == nil || !node.IsExpanded() || len(node.GetChildren()) == 0 {
		return node
	}

	// Get the last child and recursively find its last descendant
	children := node.GetChildren()
	lastChild := children[len(children)-1]
	return v.findLastDescendant(lastChild)
}
