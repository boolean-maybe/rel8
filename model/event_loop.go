package model

import (
	"context"
	"errors"
	"github.com/gdamore/tcell/v2"
	"log/slog"
	"rel8/db"
	"time"
)

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

			case "view", "views":
				headers, data := csm.server.FetchViews(ctx)
				csm.PushState(ctx, State{
					Mode:         Browse,
					TableMode:    View,
					TableHeaders: headers,
					TableData:    data,
				})

			case "procedure", "procedures", "proc", "procs":
				headers, data := csm.server.FetchProcedures(ctx)
				csm.PushState(ctx, State{
					Mode:         Browse,
					TableMode:    Procedure,
					TableHeaders: headers,
					TableData:    data,
				})

			case "function", "functions", "func", "funcs":
				headers, data := csm.server.FetchFunctions(ctx)
				csm.PushState(ctx, State{
					Mode:         Browse,
					TableMode:    Function,
					TableHeaders: headers,
					TableData:    data,
				})

			case "trigger", "triggers":
				headers, data := csm.server.FetchTriggers(ctx)
				csm.PushState(ctx, State{
					Mode:         Browse,
					TableMode:    Trigger,
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
