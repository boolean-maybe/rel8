package model

import (
	"context"
	"errors"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"log/slog"
	"rel8/db"
	"sync"
	"time"
)

type StateTransition struct {
	From State
	To   State
}

type StateChangeCallback func(transition StateTransition)

type ContextualStateManager struct {
	mu            sync.RWMutex
	stateStack    []State
	callbacks     []StateChangeCallback
	syncCallbacks []StateChangeCallback
	maxHistory    int
	server        db.DatabaseServer
}

func NewContextualStateManager(server db.DatabaseServer, initialState State, maxHistory int) *ContextualStateManager {
	return &ContextualStateManager{
		stateStack:    []State{initialState},
		callbacks:     make([]StateChangeCallback, 0),
		syncCallbacks: make([]StateChangeCallback, 0),
		maxHistory:    maxHistory,
		server:        server,
	}
}

func (csm *ContextualStateManager) AddCallback(callback StateChangeCallback) {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	csm.callbacks = append(csm.callbacks, callback)
}

func (csm *ContextualStateManager) AddSyncCallback(callback StateChangeCallback) {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	csm.syncCallbacks = append(csm.syncCallbacks, callback)
}

func (csm *ContextualStateManager) PushState(ctx context.Context, newState State) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	oldState := csm.stateStack[len(csm.stateStack)-1]
	csm.stateStack = append(csm.stateStack, newState)

	// Limit history size
	if len(csm.stateStack) > csm.maxHistory {
		csm.stateStack = csm.stateStack[1:]
	}

	// Notify callbacks
	transition := StateTransition{From: oldState, To: newState}

	// Call synchronous callbacks first (like UI updates)
	for _, callback := range csm.syncCallbacks {
		callback(transition)
	}

	// Then call async callbacks (like logging)
	for _, callback := range csm.callbacks {
		go callback(transition) // Non-blocking callbacks
	}

	return nil
}

func (csm *ContextualStateManager) PopState(ctx context.Context) (State, error) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	select {
	case <-ctx.Done():
		return *Initial, ctx.Err()
	default:
	}

	if len(csm.stateStack) <= 1 {
		return *Initial, fmt.Errorf("cannot pop the last state")
	}

	currentState := csm.stateStack[len(csm.stateStack)-1]
	csm.stateStack = csm.stateStack[:len(csm.stateStack)-1]
	previousState := csm.stateStack[len(csm.stateStack)-1]

	// Notify callbacks
	transition := StateTransition{From: currentState, To: previousState}

	// Call synchronous callbacks first (like UI updates)
	for _, callback := range csm.syncCallbacks {
		callback(transition)
	}

	// Then call async callbacks (like logging)
	for _, callback := range csm.callbacks {
		go callback(transition)
	}

	return previousState, nil
}

func (csm *ContextualStateManager) GetCurrentState() State {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	return csm.stateStack[len(csm.stateStack)-1]
}

func (csm *ContextualStateManager) GetHistory() []State {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	// Return a copy to prevent external modification
	history := make([]State, len(csm.stateStack))
	copy(history, csm.stateStack)
	return history
}

func (csm *ContextualStateManager) HandleEvent(ev *Event) *tcell.EventKey {

	currentState := csm.GetCurrentState()
	// assume no transition unless done
	noChange := StateTransition{From: currentState, To: currentState}

	// Create context with 5-second timeout for database operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Handle Ctrl-C from any mode to quit immediately
	if ev.Event.Key() == tcell.KeyCtrlC {
		csm.PushState(ctx, *Quit)
		return nil
	}

	// Handle Escape from any mode to go back to previous state
	if ev.Event.Key() == tcell.KeyEscape {
		csm.PopState(ctx)
		return nil
	}

	// If command bar is visible, handle only Enter else return event
	if csm.GetCurrentState().Mode == Command {
		switch ev.Event.Key() {
		case tcell.KeyEnter:
			slog.Debug("enter in command mode")
			command := ev.Text
			// Process the command
			switch command {
			case "q", "quit":
				csm.PushState(ctx, *Quit)

			case "table":
				headers, data := csm.server.FetchTables(ctx)
				csm.PushState(ctx, State{
					Mode:         Browse,
					TableMode:    DatabaseTable,
					TableHeaders: headers,
					TableData:    data,
				})

			case "db", "database":
				headers, data := csm.server.FetchDatabases(ctx)
				csm.PushState(ctx, State{
					Mode:         Browse,
					TableMode:    Database,
					TableHeaders: headers,
					TableData:    data,
				})
			}

			return nil
		default:
			// Let command bar handle other keys
			slog.Debug("return event for command bar")
			return ev.Event
		}
	}

	// If SQL mode is active, handle SQL commands on Enter else return event
	if csm.GetCurrentState().Mode == SQL {
		switch ev.Event.Key() {
		case tcell.KeyEnter:
			slog.Debug("enter in SQL mode")
			SQL := ev.Text
			slog.Debug("executing SQL query", "query", SQL)
			newState := csm.createStateWithSqlRows(ctx, SQL)
			csm.PushState(ctx, newState)
			return nil
		default:
			// Let command bar handle other keys for SQL input
			slog.Debug("return event for SQL command bar")
			return ev.Event
		}
	}

	// If Editor mode is active, handle F5 to execute SQL
	if csm.GetCurrentState().Mode == Editor {
		switch ev.Event.Key() {
		case tcell.KeyF5:
			slog.Debug("F5 in editor mode")
			SQL := ev.Text
			slog.Debug("executing SQL query from editor", "query", SQL)
			newState := csm.createStateWithSqlRows(ctx, SQL)
			csm.PushState(ctx, newState)
			return nil
		default:
			// Let editor handle other keys
			slog.Debug("return event for editor")
			return ev.Event
		}
	}

	// action on row in browse
	if csm.GetCurrentState().Mode == Browse {
		if csm.GetCurrentState().TableMode == DatabaseTable {
			switch ev.Event.Key() {
			case tcell.KeyEnter:
				// First update current state to save selection
				// to preserve row selection when returning
				csm.updateCurrentStateSelection(ev.Row - 1)
				newState := csm.createStateWithTableRows(ctx, ev)
				csm.PushState(ctx, newState)
				return nil
			case tcell.KeyRune:
				switch ev.Event.Rune() {
				case 'q':
					// First update current state to save selection
					// to preserve row selection when returning
					csm.updateCurrentStateSelection(ev.Row - 1)
					newState := csm.createStateWithTableRows(ctx, ev)
					csm.PushState(ctx, newState)
					return nil
				case 'd':
					// First update current state to save selection
					// to preserve row selection when returning
					csm.updateCurrentStateSelection(ev.Row - 1)
					newState := csm.createStateWithTableDescr(ctx, ev)
					csm.PushState(ctx, newState)
					return nil
				}
			}
		}
	}

	// Normal key bindings when command bar is not visible
	switch ev.Event.Key() {
	case tcell.KeyRune:
		switch ev.Event.Rune() {
		case ':':
			csm.PushState(ctx, State{
				Mode: Command,
			})
			return nil
		case '!':
			csm.PushState(ctx, State{
				Mode: SQL,
			})
			return nil
		case 's':
			csm.PushState(ctx, State{
				Mode: Editor,
			})
			return nil
		default:
			// Let other rune keys pass through to the table
			// Call synchronous callbacks first
			for _, callback := range csm.syncCallbacks {
				callback(noChange)
			}
			// Then call async callbacks
			for _, callback := range csm.callbacks {
				go callback(noChange) // Non-blocking callbacks
			}
			return ev.Event
		}
	default:
		// Let navigation keys (arrows, page up/down, etc.) pass through to the table
		return ev.Event
	}
}

func (csm *ContextualStateManager) createStateWithTableRows(ctx context.Context, ev *Event) State {
	slog.Debug("row", "row", ev.Row)

	// this creates a shallow copy
	newState := csm.GetCurrentState()

	newState.SelectedDataIndex = ev.Row - 1
	tableName, _ := extractNameFromSelection(csm.GetCurrentState(), ev.Row-1)
	// Fetch table rows using the extracted table name
	headers, data := csm.server.FetchTableRows(ctx, tableName)
	newState.TableMode = TableRow
	newState.TableHeaders = headers
	newState.TableData = data

	return newState
}

func (csm *ContextualStateManager) createStateWithSqlRows(ctx context.Context, SQL string) State {
	newState := State{
		Mode: Browse,
	}

	// Fetch SQL rows using the extracted table name
	headers, data := csm.server.FetchSqlRows(ctx, SQL)
	newState.TableMode = TableRow
	newState.TableHeaders = headers
	newState.TableData = data

	return newState
}

func (csm *ContextualStateManager) createStateWithTableDescr(ctx context.Context, ev *Event) State {
	slog.Debug("row", "row", ev.Row)
	// this creates a shallow copy
	newState := csm.GetCurrentState()

	newState.SelectedDataIndex = ev.Row - 1
	tableName, _ := extractNameFromSelection(csm.GetCurrentState(), ev.Row-1)
	// Fetch table description using the extracted table name
	newState.Mode = Detail
	newState.DetailText = csm.server.FetchTableDescr(ctx, tableName)

	return newState
}

func (csm *ContextualStateManager) updateCurrentStateSelection(selectedIndex int) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if len(csm.stateStack) > 0 {
		currentState := &csm.stateStack[len(csm.stateStack)-1]
		currentState.SelectedDataIndex = selectedIndex
	}
}

// attempt to extract object name such as table name from selected row in table data
func extractNameFromSelection(state State, selected int) (string, error) {
	selectedTable := state.TableData[selected]

	if mysqlTable, ok := selectedTable.(db.MysqlTable); ok {
		return mysqlTable.Name, nil
	} else {
		return "", errors.New("failed to cast selected row to db.MysqlTable")
	}
}
