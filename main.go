package main

import (
	"rel8/config"
	"rel8/db"
	"rel8/model"
	"rel8/view"
)

func main() {
	connectString, useMock, demoScript := config.Configure()
	server := db.Connect(connectString, useMock)

	stateManager := model.NewContextualStateManager(server, *model.Initial, 20)
	view := view.NewView(stateManager)
	view.OnStateTransition(model.StateTransition{*model.Initial, *model.Initial})

	// Add a callback to notify view (synchronous to avoid race conditions)
	stateManager.AddSyncCallback(func(transition model.StateTransition) {
		view.OnStateTransition(transition)
	})

	// Add a callback to log state changes (async is fine for logging)
	stateManager.AddCallback(func(transition model.StateTransition) {
		//todo make slog
		//fmt.Printf("State transition: %s -> %s\n", transition.From, transition.To)
	})

	// Handle demo mode or run normally
	demoMode(view, demoScript)

	// Not demo mode, run normally
	view.Run()
}
