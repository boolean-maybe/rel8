package view

import (
	"log/slog"
	"rel8/model"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type View struct {
	stateManager *model.ContextualStateManager
	model        *model.State
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

// New Notify process events - inspect model and redraw
func (v *View) OnStateTransition(transition model.StateTransition) {
	//todo take address?
	v.model = &transition.To

	// Update server info in header when state changes
	// This ensures the header shows the current database when it changes
	v.header.UpdateServerInfo()

	if transition.To.Mode == model.QuitMode {
		v.App.Stop()
	}

	if transition.To.Mode == model.Browse {
		v.flex.Clear()
		v.flex.AddItem(WrapHeader(v.header), 7, 0, false)
		
		// If we have table data, show the grid; otherwise show the tree
		if len(transition.To.TableData) > 0 {
			// repopulate grid without recreating it
			v.grid.Populate(transition.To.TableHeaders, transition.To.TableData)

			// stretch last column in view-like datasets so highlight paints to edge (non-invasive)
			if len(transition.To.TableHeaders) > 0 {
				h := transition.To.TableHeaders
				if (len(h) >= 5 && (h[0] == "NAME" && h[1] == "TYPE")) || (len(h) >= 6 && h[len(h)-1] == "SECURITY_TYPE") {
					v.grid.StretchLastColumn()
				}
			}

			// Restore the selected row if one was saved
			v.grid.RestoreSelection(transition.To.SelectedDataIndex, len(transition.To.TableData))

			v.flex.AddItem(WrapGrid(v.grid), 0, 1, true)
			v.App.SetFocus(v.grid)
		} else {
			// Show tree view as default
			v.flex.AddItem(WrapTree(v.tree), 0, 1, true)
			v.App.SetFocus(v.tree.TreeView)
		}
	}

	if transition.To.Mode == model.Detail {
		v.flex.Clear()
		v.details = NewDetail(transition.To.DetailText)
		v.flex.AddItem(WrapHeader(v.header), 7, 0, false)
		v.flex.AddItem(WrapDetail(v.details), 0, 1, true)
		v.App.SetFocus(v.details)
	}

	if transition.To.Mode == model.Command {
		// Show command bar between header and main content
		v.commandBar.Show()
		v.flex.Clear()
		v.flex.AddItem(WrapHeader(v.header), 7, 0, false)
		v.flex.AddItem(WrapCommandBar(v.commandBar), 3, 0, false)
		
		// Show grid if we have data, otherwise show tree
		if len(transition.To.TableData) > 0 {
			v.flex.AddItem(WrapGrid(v.grid), 0, 1, true)
		} else {
			v.flex.AddItem(WrapTree(v.tree), 0, 1, true)
		}
		v.App.SetFocus(v.commandBar)
	}

	if transition.To.Mode == model.SQL {
		// Show command bar for SQL input between header and main content
		v.commandBar.Show()
		v.flex.Clear()
		v.flex.AddItem(WrapHeader(v.header), 7, 0, false)
		v.flex.AddItem(WrapCommandBar(v.commandBar), 3, 0, false)
		
		// Show grid if we have data, otherwise show tree
		if len(transition.To.TableData) > 0 {
			v.flex.AddItem(WrapGrid(v.grid), 0, 1, true)
		} else {
			v.flex.AddItem(WrapTree(v.tree), 0, 1, true)
		}
		v.App.SetFocus(v.commandBar)
	}

	if transition.To.Mode == model.Editor {
		v.flex.Clear()
		v.flex.AddItem(WrapHeader(v.header), 7, 0, false)
		v.flex.AddItem(WrapEditor(v.editor), 0, 1, true)
		v.App.SetFocus(v.editor)
	}
}

// Run - run event cycle
func (v *View) Run() {
	// Add key bindings
	v.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Input captured
		
		// Get current state from state manager
		currentState := v.stateManager.GetCurrentState()
		
		// In browse mode with tree view, handle tree-specific keys first
		if currentState.Mode == model.Browse && len(currentState.TableData) == 0 {
			switch event.Key() {
			case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight:
				// Manual navigation for TreeView
				before := v.tree.treeView.GetCurrentNode()
				
				if event.Key() == tcell.KeyDown {
					nextNode := v.findNextNode(before)
					if nextNode != nil {
						v.tree.treeView.SetCurrentNode(nextNode)
					}
				} else if event.Key() == tcell.KeyUp {
					prevNode := v.findPrevNode(before)
					if prevNode != nil {
						v.tree.treeView.SetCurrentNode(prevNode)
					}
				}
				
				// No need to force redraw - tview handles selection highlighting automatically
				
				return nil // Consume the event since we handled it manually
			case tcell.KeyEnter:
				// Handle Enter key for tree expansion
				currentNode := v.tree.treeView.GetCurrentNode()
				if currentNode != nil {
					v.tree.HandleNodeActivation(currentNode)
				}
				return nil // Consume the event
			case tcell.KeyRune:
				if event.Rune() == ' ' {
					// Handle Space key for tree expansion
					currentNode := v.tree.treeView.GetCurrentNode()
					if currentNode != nil {
						v.tree.HandleNodeActivation(currentNode)
					}
					return nil // Consume the event
				}
				// Let other runes (like ':') pass through to global handler
			}
			// Other keys (like ':', '!', 's', Escape, Ctrl+C) fall through to global handler
		}
		
		e := &model.Event{Event: event}

		// if in command mode also send command bar text
		if currentState.Mode == model.Command {
			e.Text = v.commandBar.GetCommand()
		}
		// if in SQL mode also send command bar text (SQL query)
		if currentState.Mode == model.SQL {
			e.Text = v.commandBar.GetCommand()
		}
		// if in editor mode also send editor text
		if currentState.Mode == model.Editor {
			e.Text = v.editor.GetText()
		}
		// if in browse mode also send current selection (row for grid, node for tree)
		if currentState.Mode == model.Browse {
			if len(currentState.TableData) > 0 {
				// Grid mode - send current row
				slog.Info("in browse mode sending grid row")
				row, _ := v.grid.GetSelection()
				slog.Info("sending row:", "row", row)
				e.Row = row
			} else {
				// Tree mode - send current tree node info
				slog.Info("in browse mode sending tree node")
				selectedNode := v.tree.GetSelectedNode()
				if selectedNode != nil {
					// Convert tree node to model-compatible format
					var parentName, database string
					if selectedNode.Parent != nil {
						parentName = selectedNode.Parent.Name
					}
					if selectedNode.Type == "item" && selectedNode.Parent != nil && selectedNode.Parent.Parent != nil {
						database = selectedNode.Parent.Parent.Name
					}
					
					e.TreeNode = &model.TreeNodeInfo{
						Type:     selectedNode.Type,
						Name:     selectedNode.Name,
						Parent:   parentName,
						Database: database,
					}
					slog.Info("sending tree node:", "type", selectedNode.Type, "name", selectedNode.Name)
				}
			}
		}

		//todo this is a single place that requires state manager. Replace with a function
		return v.stateManager.HandleEvent(e)
	})

	// attach selection band painting via the grid, gated to Browse mode
	v.grid.AttachSelectionBand(v.App, func() bool {
		if v.model == nil {
			return false
		}
		// Only show grid selection band when we have table data
		return (v.model.Mode == model.Browse || v.model.Mode == model.Command || v.model.Mode == model.SQL) && len(v.model.TableData) > 0
	})
	
	// Note: Tree selection band removed - using tview's built-in selection highlighting
	// The manual navigation provides good visual selection feedback

	// Run the application
	if err := v.App.SetRoot(v.flex, true).SetFocus(v.grid).Run(); err != nil {
		slog.Error("Error running tview app", "error", err)
	}
}
