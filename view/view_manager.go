package view

import (
	"log/slog"
	"rel8/model"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ViewManager struct {
	// state manager event handler to pass events to
	eventHandler func(ev *model.Event) *tcell.EventKey
	App          *tview.Application
	pages        *tview.Pages
	state        model.State
}

func NewViewManager(eventHandler func(ev *model.Event) *tcell.EventKey, app *tview.Application, pages *tview.Pages) *ViewManager {
	return &ViewManager{
		eventHandler: eventHandler,
		App:          app,
		pages:        pages,
	}
}

type UIComponent struct {
	item       tview.Primitive
	fixedSize  int
	proportion int
	focus      bool
}

// create UI component out of new state
func (v *ViewManager) makeComponent(state model.State) tview.Primitive {
	components := make([]UIComponent, 0)

	headerInfo := state.GetHeaderInfo()
	header := NewHeader(headerInfo)
	components = append(components, UIComponent{
		header,
		7,
		0,
		false,
	})

	if state.HasCommand() {
		commandBar := NewCommandBar()
		components = append(components, UIComponent{
			commandBar,
			3,
			0,
			true,
		})
	}

	if state.HasBrowse() && state.GetBrowseState().BrowseClass == model.DatabaseTable {
		// browsing database tables
		data := state.GetBrowseState().TableInfo
		grid := NewGrid(data.TableHeaders, data.TableData, v.eventHandler)

		components = append(components, UIComponent{
			grid,
			0,
			1,
			!state.HasCommand(),
		})

	}

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	for _, component := range components {
		flex.AddItem(component.item, component.fixedSize, component.proportion, component.focus)
	}

	return flex
}

// create input capture function for new state
func (v *ViewManager) makeInputCapture(state *model.State) func(event *tcell.EventKey) *tcell.EventKey {

	hasBrowse := (*state).HasBrowse()
	hasCommand := (*state).HasCommand()

	if hasCommand {
		return func(event *tcell.EventKey) *tcell.EventKey {
			return event
		}
	}

	if hasBrowse && (*state).GetBrowseState().BrowseClass == model.DatabaseTable {
		// keys to process in this mode - Enter, d, :
		return func(event *tcell.EventKey) *tcell.EventKey {
			// Get current state from state manager

			e := &model.Event{
				EventType: model.Other,
				Event:     event,
			}

			// if in command mode also send command bar text
			//if state.GetMode().Kind == model.Command {
			//	//e.Text = v.commandBar.GetCommand()
			//}

			return v.eventHandler(e)
		}
	}

	// example key capture function
	return func(event *tcell.EventKey) *tcell.EventKey {
		// Get current state from state manager

		e := &model.Event{Event: event}

		// if in command mode also send command bar text
		//if state.GetMode().Kind == model.Command {
		//	//e.Text = v.commandBar.GetCommand()
		//}

		//todo this is a single place that requires state manager. Replace with a function
		return v.eventHandler(e)
	}
}

// New Notify process events - inspect model and redraw
func (v *ViewManager) OnStateTransition(transition model.StateTransition) {

	newState, isPop := transition.To, transition.IsPop
	v.state = newState

	//if transition.To.GetMode().Kind == model.QuitKind {
	//	v.App.Stop()
	//}

	if !isPop {
		// on transition to new state add a page
		// create visual component
		component := v.makeComponent(newState)
		// create its key capture
		//capture := v.makeInputCapture(newState)

		//if flex, ok := component.(*tview.Flex); ok {
		//	flex.SetInputCapture(capture)
		//} else {
		//	println("Error creating box")
		//}

		v.pages.AddAndSwitchToPage("name", component, true)
	} else {
		// on popping remove current page and show previous
		name, _ := v.pages.GetFrontPage()
		v.pages.RemovePage(name)
	}
}

// Run - run event cycle
func (v *ViewManager) Run() {
	// App.SetInputCapture sets a function which captures all key events before they are
	// forwarded to the key event handler of the primitive which currently has focus.
	v.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			v.App.Stop()
		}

		// let view component handle other events
		return event

		// if events need to be forwarded to the state manager
		// e := &model.Event{Event: event}
		// return v.eventHandler(e)

		// Get current state from state manager
		// TODO: Need access to state manager or current state
		//currentState := v.state
		//
		//e := &model.Event{Event: event}
		//
		//// if in command mode also send command bar text
		//if (*currentState).GetMode().Kind == model.Command {
		//	//e.Text = v.commandBar.GetCommand()
		//}
		//
		////todo this is a single place that requires state manager. Replace with a function
		//return v.eventHandler(e)
	})

	// Run the application
	if err := v.App.SetRoot(v.pages, true).SetFocus(v.pages).Run(); err != nil {
		slog.Error("Error running tview app", "error", err)
	}
}
