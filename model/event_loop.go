package model

import (
	"context"
	"log/slog"
	"rel8/db"
	"time"

	"github.com/gdamore/tcell/v2"
)

func (csm *ContextualStateManager) HandleEvent(ev *Event) *tcell.EventKey {
	// Create context with 5-second timeout for database operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if ev.EventType == Init {
		// initially show database tables
		headers, data := csm.server.FetchTables(ctx)
		serverInfo := HeaderInfo(csm.server.GetServerInfo(ctx))

		browseState := &BrowseState{
			BrowseClass: DatabaseTable,
			TableInfo: &TableInfo{
				TableHeaders:      headers,
				TableData:         data,
				SelectedDataIndex: 0,
			},
		}
		commonState := &CommonState{
			HeaderInfo: &serverInfo,
		}
		state := NewStateBuilder().SetCommon(commonState).SetBrowse(browseState).Build()
		csm.PushState(ctx, state)

		return nil
	}

	currentState := csm.GetCurrentState()
	// assume no transition unless done
	noChange := StateTransition{From: currentState, To: currentState, IsPop: false}

	// Handle Escape from any mode to go back to previous state
	if ev.Event.Key() == tcell.KeyEscape {
		csm.PopState(ctx)
		return nil
	}

	// If command bar is visible, handle only Enter else return event
	if csm.GetCurrentState().HasCommand() {
		switch ev.Event.Key() {
		case tcell.KeyEnter:
			slog.Debug("enter in command mode")
			command := ev.Text
			// Process the command
			switch command {
			case "table":
				headers, data := csm.server.FetchTables(ctx)
				serverInfo := HeaderInfo(csm.server.GetServerInfo(ctx))

				browseState := &BrowseState{
					BrowseClass: DatabaseTable,
					TableInfo: &TableInfo{
						TableHeaders:      headers,
						TableData:         data,
						SelectedDataIndex: 0,
					},
				}
				commonState := &CommonState{
					HeaderInfo: &serverInfo,
				}
				state := NewStateBuilder().SetCommon(commonState).SetBrowse(browseState).Build()
				csm.PushState(ctx, state)

			case "db", "database":
				headers, data := csm.server.FetchDatabases(ctx)
				serverInfo := HeaderInfo(csm.server.GetServerInfo(ctx))

				browseState := &BrowseState{
					BrowseClass: Database,
					TableInfo: &TableInfo{
						TableHeaders:      headers,
						TableData:         data,
						SelectedDataIndex: 0,
					},
				}
				commonState := &CommonState{
					HeaderInfo: &serverInfo,
				}
				state := NewStateBuilder().SetCommon(commonState).SetBrowse(browseState).Build()
				csm.PushState(ctx, state)

			case "view", "views":
				headers, data := csm.server.FetchViews(ctx)
				serverInfo := HeaderInfo(csm.server.GetServerInfo(ctx))

				browseState := &BrowseState{
					BrowseClass: View,
					TableInfo: &TableInfo{
						TableHeaders:      headers,
						TableData:         data,
						SelectedDataIndex: 0,
					},
				}
				commonState := &CommonState{
					HeaderInfo: &serverInfo,
				}
				state := NewStateBuilder().SetCommon(commonState).SetBrowse(browseState).Build()
				csm.PushState(ctx, state)

			case "procedure", "procedures", "proc", "procs":
				headers, data := csm.server.FetchProcedures(ctx)
				serverInfo := HeaderInfo(csm.server.GetServerInfo(ctx))

				browseState := &BrowseState{
					BrowseClass: Procedure,
					TableInfo: &TableInfo{
						TableHeaders:      headers,
						TableData:         data,
						SelectedDataIndex: 0,
					},
				}
				commonState := &CommonState{
					HeaderInfo: &serverInfo,
				}
				state := NewStateBuilder().SetCommon(commonState).SetBrowse(browseState).Build()
				csm.PushState(ctx, state)

			case "function", "functions", "func", "funcs":
				headers, data := csm.server.FetchFunctions(ctx)
				serverInfo := HeaderInfo(csm.server.GetServerInfo(ctx))

				browseState := &BrowseState{
					BrowseClass: Function,
					TableInfo: &TableInfo{
						TableHeaders:      headers,
						TableData:         data,
						SelectedDataIndex: 0,
					},
				}
				commonState := &CommonState{
					HeaderInfo: &serverInfo,
				}
				state := NewStateBuilder().SetCommon(commonState).SetBrowse(browseState).Build()
				csm.PushState(ctx, state)

			case "trigger", "triggers":
				headers, data := csm.server.FetchTriggers(ctx)
				serverInfo := HeaderInfo(csm.server.GetServerInfo(ctx))

				browseState := &BrowseState{
					BrowseClass: Trigger,
					TableInfo: &TableInfo{
						TableHeaders:      headers,
						TableData:         data,
						SelectedDataIndex: 0,
					},
				}
				commonState := &CommonState{
					HeaderInfo: &serverInfo,
				}
				state := NewStateBuilder().SetCommon(commonState).SetBrowse(browseState).Build()
				csm.PushState(ctx, state)
			}

			return nil
		default:
			// Let command bar handle other keys
			slog.Debug("return event for command bar")
			return ev.Event
		}
	}

	// If SQL mode is active, handle SQL commands on Enter else return event
	if csm.GetCurrentState().HasSql() {
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
	if csm.GetCurrentState().HasFullSql() {
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

	// Normal key bindings when command bar is not visible
	switch ev.Event.Key() {
	case tcell.KeyUp:
		currentState.GetBrowseState().TableInfo.SelectedDataIndex--
		return ev.Event
	case tcell.KeyDown:
		currentState.GetBrowseState().TableInfo.SelectedDataIndex++
		return ev.Event

	case tcell.KeyEnter:
		currentState := csm.GetCurrentState()
		if currentState.HasBrowse() && currentState.GetBrowseState().BrowseClass == DatabaseTable {
			// extract table name
			selectedIndex := currentState.GetBrowseState().TableInfo.SelectedDataIndex
			tableData := currentState.GetBrowseState().TableInfo.TableData
			tableName := tableData[selectedIndex].(db.PostgresTable).Name

			headers, data := csm.server.FetchTableRows(ctx, tableName)
			serverInfo := HeaderInfo(csm.server.GetServerInfo(ctx))

			browseState := &BrowseState{
				BrowseClass: TableRow,
				TableInfo: &TableInfo{
					TableHeaders:      headers,
					TableData:         data,
					SelectedDataIndex: 0,
				},
			}
			commonState := &CommonState{
				HeaderInfo: &serverInfo,
			}
			state := NewStateBuilder().SetCommon(commonState).SetBrowse(browseState).Build()
			csm.PushState(ctx, state)
		}

		return nil

	case tcell.KeyRune:
		switch ev.Event.Rune() {
		case ':':
			currentState := csm.GetCurrentState()
			newState := NewStateBuilderFromState(currentState).SetEmptyCommand().Build()
			csm.PushState(ctx, newState)
			return nil
		case '!':
			currentState := csm.GetCurrentState()
			newState := NewStateBuilderFromState(currentState).SetEmptySql().Build()
			csm.PushState(ctx, newState)
			return nil
		case 's':
			//csm.PushState(ctx, &EditorState{
			//	Mode: Editor,
			//})
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

func (csm *ContextualStateManager) createStateWithSqlRows(ctx context.Context, SQL string) State {
	// Fetch SQL rows using the extracted table name
	headers, data := csm.server.FetchSqlRows(ctx, SQL)
	serverInfo := HeaderInfo(csm.server.GetServerInfo(ctx))

	browseState := &BrowseState{
		BrowseClass: TableRow,
		TableInfo: &TableInfo{
			TableHeaders:      headers,
			TableData:         data,
			SelectedDataIndex: 0,
		},
	}
	commonState := &CommonState{
		HeaderInfo: &serverInfo,
	}
	return NewStateBuilder().SetCommon(commonState).SetBrowse(browseState).Build()
}
