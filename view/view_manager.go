package view

import (
	"fmt"
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
	pageCount    int
	state        model.State
}

func NewViewManager(eventHandler func(ev *model.Event) *tcell.EventKey, app *tview.Application, pages *tview.Pages) *ViewManager {
	return &ViewManager{
		eventHandler: eventHandler,
		App:          app,
		pages:        pages,
		pageCount:    0,
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

	//if state.HasBrowse() && state.GetBrowseState().BrowseClass == model.DatabaseTable {
	if state.HasBrowse() {
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

// New Notify process events - inspect model and redraw
func (v *ViewManager) OnStateTransition(transition model.StateTransition) {

	newState, isPop := transition.To, transition.IsPop
	v.state = newState

	if !isPop {
		// on transition to new state add a page and create visual component
		component := v.makeComponent(newState)
		// create its key capture
		//capture := v.makeInputCapture(newState)

		//if flex, ok := component.(*tview.Flex); ok {
		//	flex.SetInputCapture(capture)
		//}

		v.pageCount++
		v.pages.AddAndSwitchToPage(fmt.Sprintf("page-%d", v.pageCount), component, true)
	} else {
		// on popping remove current page and show previous
		//name, _ := v.pages.GetFrontPage()
		v.pages.RemovePage(fmt.Sprintf("page-%d", v.pageCount))
		v.pageCount--
		v.pages.SwitchToPage(fmt.Sprintf("page-%d", v.pageCount))
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

		if event.Key() == tcell.KeyEscape {
			e := &model.Event{Event: event, EventType: model.Other}
			v.eventHandler(e)
		}

		// let view component handle other events
		return event
	})

	// Run the application
	if err := v.App.SetRoot(v.pages, true).SetFocus(v.pages).Run(); err != nil {
		slog.Error("Error running tview app", "error", err)
	}
}
