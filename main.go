package main

import (
	"rel8/config"
	"rel8/db"
	"rel8/model"
	"rel8/view"

	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	pages := tview.NewPages()

	//connectString, useMock, demoScript := config.Configure()
	connectString, useMock, _ := config.Configure()
	server := db.Connect(connectString, useMock)

	stateManager := model.NewContextualStateManager(server, 20)
	viewManager := view.NewViewManager(stateManager.HandleEvent, app, pages)

	// Add a callback to notify view (synchronous to avoid race conditions)
	stateManager.AddSyncCallback(func(transition model.StateTransition) {
		viewManager.OnStateTransition(transition)
	})

	// Add a callback to log state changes (async is fine for logging)
	stateManager.AddCallback(func(transition model.StateTransition) {
		//todo make slog
		//fmt.Printf("State transition: %s -> %s\n", transition.From, transition.To)
	})

	// send Init event
	stateManager.HandleEvent(model.InitEvent)

	// Handle demo mode or run normally
	//demoMode(viewManager, demoScript)

	// Not demo mode, run normally
	viewManager.Run()
}
