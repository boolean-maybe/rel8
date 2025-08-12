package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"log/slog"
	"rel8/model"
)

type View struct {
	stateManager *model.ContextualStateManager
	model        *model.State
	App          *tview.Application
	flex         *tview.Flex
	header       *Header
	grid         *Grid
	details      *Detail
	editor       *Editor
	commandBar   *CommandBar
}

func NewView(stateManager *model.ContextualStateManager) *View {
	app := tview.NewApplication()

	// Create components
	header := NewHeader()
	details := NewEmptyDetail()
	editor := NewEmptyEditor()

	grid := NewEmptyGrid()

	// Create command bar (initially hidden)
	commandBar := NewCommandBar()

	// Set the server in the header and update the server info
	header.SetServer(stateManager.GetServer())

	// Create layout with command bar
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(WrapHeader(header), 7, 0, false). // Fixed header height with padding
		AddItem(WrapGrid(grid), 0, 1, true)       // Grid with padding takes remaining space

	view := &View{
		stateManager: stateManager,
		model:        model.Initial,
		App:          app,
		flex:         flex,
		header:       header,
		grid:         grid,
		details:      details,
		editor:       editor,
		commandBar:   commandBar,
	}

	return view
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
		// repopulate grid without recreating it
		v.grid.Populate(transition.To.TableHeaders, transition.To.TableData)

		// Restore the selected row if one was saved
		v.grid.RestoreSelection(transition.To.SelectedDataIndex, len(transition.To.TableData))

		v.flex.Clear()
		v.flex.AddItem(WrapHeader(v.header), 7, 0, false)
		v.flex.AddItem(WrapGrid(v.grid), 0, 1, true)
		v.App.SetFocus(v.grid)
	}

	if transition.To.Mode == model.Detail {
		v.flex.Clear()
		v.details = NewDetail(transition.To.DetailText)
		v.flex.AddItem(WrapHeader(v.header), 7, 0, false)
		v.flex.AddItem(WrapDetail(v.details), 0, 1, true)
		v.App.SetFocus(v.details)
	}

	if transition.To.Mode == model.Command {
		// Show command bar between header and table
		v.commandBar.Show()
		v.flex.Clear()
		v.flex.AddItem(WrapHeader(v.header), 7, 0, false)
		v.flex.AddItem(WrapCommandBar(v.commandBar), 3, 0, false)
		v.flex.AddItem(WrapGrid(v.grid), 0, 1, true)
		v.App.SetFocus(v.commandBar)
	}

	if transition.To.Mode == model.SQL {
		// Show command bar for SQL input between header and table
		v.commandBar.Show()
		v.flex.Clear()
		v.flex.AddItem(WrapHeader(v.header), 7, 0, false)
		v.flex.AddItem(WrapCommandBar(v.commandBar), 3, 0, false)
		v.flex.AddItem(WrapGrid(v.grid), 0, 1, true)
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
		slog.Info("Input captured", "key", event.Key(), "rune", event.Rune())
		e := &model.Event{Event: event}

		// Get current state from state manager
		currentState := v.stateManager.GetCurrentState()

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
		// if in browse mode also send current row
		if currentState.Mode == model.Browse {
			slog.Info("in browse mode sending row")
			row, _ := v.grid.GetSelection()
			slog.Info("sending row:", "row", row)
			e.Row = row
		}

		//todo this is a single place that requires state manager. Replace with a function
		return v.stateManager.HandleEvent(e)
	})

	// Run the application
	if err := v.App.SetRoot(v.flex, true).SetFocus(v.grid).Run(); err != nil {
		slog.Error("Error running tview app", "error", err)
	}
}
