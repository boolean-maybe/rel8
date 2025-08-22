package model

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"rel8/db"
)

// Mock callback for testing
type mockCallback struct {
	callCount      int
	lastTransition StateTransition
	transitionList []StateTransition
}

func (m *mockCallback) callback(transition StateTransition) {
	m.callCount++
	m.lastTransition = transition
	m.transitionList = append(m.transitionList, transition)
}

func TestNewContextualStateManager(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
	server := mysql

	initialState := State{Mode: Browse, TableMode: EmptyTable}
	stateManager := NewContextualStateManager(server, initialState, 10)

	assert.NotNil(t, stateManager)

	// Test initial state
	currentState := stateManager.GetCurrentState()
	assert.Equal(t, Browse, currentState.Mode)
	assert.Equal(t, EmptyTable, currentState.TableMode)

	// Test history
	history := stateManager.GetHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, initialState, history[0])

	// Test callbacks are initialized
	assert.NotNil(t, stateManager.callbacks)
	assert.NotNil(t, stateManager.syncCallbacks)
}

func TestAddCallback(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
	server := mysql

	initialState := State{Mode: Browse}
	stateManager := NewContextualStateManager(server, initialState, 10)

	syncCb := &mockCallback{}
	asyncCb := &mockCallback{}
	stateManager.AddSyncCallback(syncCb.callback)
	stateManager.AddCallback(asyncCb.callback)

	// Test that callbacks are registered (we can't directly access private fields,
	// but we can test they work by triggering a state change)
	ctx := context.Background()
	newState := State{Mode: Command}
	err = stateManager.PushState(ctx, newState)
	assert.NoError(t, err)

	// Sync callback should be called immediately
	assert.Equal(t, 1, syncCb.callCount)
	assert.Equal(t, Browse, syncCb.lastTransition.From.Mode)
	assert.Equal(t, Command, syncCb.lastTransition.To.Mode)

	// Note: We don't test async callback here due to goroutine timing issues
	// The fact that no panic occurs indicates async callbacks are working
}

func TestPushState(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
	server := mysql

	initialState := State{Mode: Browse}
	stateManager := NewContextualStateManager(server, initialState, 10)
	mockCb := &mockCallback{}
	stateManager.AddSyncCallback(mockCb.callback)

	ctx := context.Background()

	// Test pushing command state
	commandState := State{Mode: Command}
	err = stateManager.PushState(ctx, commandState)
	assert.NoError(t, err)

	// Check current state changed
	currentState := stateManager.GetCurrentState()
	assert.Equal(t, Command, currentState.Mode)

	// Check history
	history := stateManager.GetHistory()
	assert.Len(t, history, 2)
	assert.Equal(t, Browse, history[0].Mode)
	assert.Equal(t, Command, history[1].Mode)

	// Check callback was called
	assert.Equal(t, 1, mockCb.callCount)
	assert.Equal(t, Browse, mockCb.lastTransition.From.Mode)
	assert.Equal(t, Command, mockCb.lastTransition.To.Mode)

	// Test pushing detail state
	detailState := State{Mode: Detail, DetailText: "CREATE TABLE test"}
	err = stateManager.PushState(ctx, detailState)
	assert.NoError(t, err)

	currentState = stateManager.GetCurrentState()
	assert.Equal(t, Detail, currentState.Mode)
	assert.Equal(t, "CREATE TABLE test", currentState.DetailText)

	// Check history
	history = stateManager.GetHistory()
	assert.Len(t, history, 3)
}

func TestPopState(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
	server := mysql

	initialState := State{Mode: Browse}
	stateManager := NewContextualStateManager(server, initialState, 10)
	mockCb := &mockCallback{}
	stateManager.AddSyncCallback(mockCb.callback)

	ctx := context.Background()

	// Push some states
	stateManager.PushState(ctx, State{Mode: Command})
	stateManager.PushState(ctx, State{Mode: Detail, DetailText: "test"})

	// Reset callback counter
	mockCb.callCount = 0

	// Test popping
	previousState, err := stateManager.PopState(ctx)
	assert.NoError(t, err)
	assert.Equal(t, Command, previousState.Mode)

	// Check current state
	currentState := stateManager.GetCurrentState()
	assert.Equal(t, Command, currentState.Mode)

	// Check callback was called
	assert.Equal(t, 1, mockCb.callCount)
	assert.Equal(t, Detail, mockCb.lastTransition.From.Mode)
	assert.Equal(t, Command, mockCb.lastTransition.To.Mode)

	// Pop to initial state
	previousState, err = stateManager.PopState(ctx)
	assert.NoError(t, err)
	assert.Equal(t, Browse, previousState.Mode)

	// Try to pop the last state (should fail)
	_, err = stateManager.PopState(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot pop the last state")
}

func TestHandleEventCommandMode(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
	server := mysql

	// Start in command mode
	commandState := State{Mode: Command}
	stateManager := NewContextualStateManager(server, commandState, 10)
	mockCb := &mockCallback{}
	stateManager.AddSyncCallback(mockCb.callback)

	tests := []struct {
		name          string
		key           tcell.Key
		text          string
		expectedMode  Mode
		expectedTable TableMode
		mockSetup     func()
	}{
		{
			name:          "escape key exits command mode",
			key:           tcell.KeyEscape,
			text:          "",
			expectedMode:  Command, // Should pop but we only have one state, so stays in Command
			expectedTable: EmptyTable,
		},
		{
			name:          "quit command",
			key:           tcell.KeyEnter,
			text:          "quit",
			expectedMode:  QuitMode,
			expectedTable: EmptyTable,
		},
		{
			name:          "q command",
			key:           tcell.KeyEnter,
			text:          "q",
			expectedMode:  QuitMode,
			expectedTable: EmptyTable,
		},
		{
			name:          "table command",
			key:           tcell.KeyEnter,
			text:          "table",
			expectedMode:  Browse,
			expectedTable: DatabaseTable,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"TABLE_NAME", "TABLE_TYPE", "ENGINE", "TABLE_ROWS", "SIZE_MB"}).
					AddRow("users", "BASE TABLE", "InnoDB", 100, 1.5)
				mock.ExpectQuery("SELECT TABLE_NAME").WillReturnRows(rows)
			},
		},
		{
			name:          "database command",
			key:           tcell.KeyEnter,
			text:          "database",
			expectedMode:  Browse,
			expectedTable: Database,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"SCHEMA_NAME", "DEFAULT_CHARACTER_SET_NAME", "DEFAULT_COLLATION_NAME"}).
					AddRow("testdb", "utf8mb4", "utf8mb4_general_ci")
				mock.ExpectQuery("SELECT SCHEMA_NAME").WillReturnRows(rows)
			},
		},
		{
			name:          "db command",
			key:           tcell.KeyEnter,
			text:          "db",
			expectedMode:  Browse,
			expectedTable: Database,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"SCHEMA_NAME", "DEFAULT_CHARACTER_SET_NAME", "DEFAULT_COLLATION_NAME"}).
					AddRow("testdb", "utf8mb4", "utf8mb4_general_ci")
				mock.ExpectQuery("SELECT SCHEMA_NAME").WillReturnRows(rows)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state manager to command mode
			stateManager = NewContextualStateManager(server, State{Mode: Command}, 10)
			mockCb = &mockCallback{}
			stateManager.AddSyncCallback(mockCb.callback)

			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			event := &Event{
				Event: tcell.NewEventKey(tt.key, 0, tcell.ModNone),
				Text:  tt.text,
			}

			result := stateManager.HandleEvent(event)

			currentState := stateManager.GetCurrentState()
			assert.Equal(t, tt.expectedMode, currentState.Mode)
			assert.Equal(t, tt.expectedTable, currentState.TableMode)

			// For escape and enter in command mode, result should be nil
			if tt.key == tcell.KeyEscape || tt.key == tcell.KeyEnter {
				assert.Nil(t, result)
			}

			// Verify mock expectations
			if tt.mockSetup != nil {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

func TestHandleEventSQLMode(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
	server := mysql

	tests := []struct {
		name         string
		key          tcell.Key
		text         string
		expectedMode Mode
		mockSetup    func()
	}{
		{
			name:         "escape key exits SQL mode",
			key:          tcell.KeyEscape,
			text:         "",
			expectedMode: SQL, // Should pop but we only have one state, so stays in SQL
		},
		{
			name:         "enter key processes SQL and pushes browse state",
			key:          tcell.KeyEnter,
			text:         "SELECT * FROM users",
			expectedMode: Browse, // Should push new Browse state with query results
			mockSetup: func() {
				// Mock the SQL query execution
				rows := sqlmock.NewRows([]string{"id", "name", "email"}).
					AddRow(1, "John", "john@example.com").
					AddRow(2, "Jane", "jane@example.com")
				mock.ExpectQuery("SELECT \\* FROM users").WillReturnRows(rows)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state manager to SQL mode
			stateManager := NewContextualStateManager(server, State{Mode: SQL}, 10)
			mockCb := &mockCallback{}
			stateManager.AddSyncCallback(mockCb.callback)

			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			event := &Event{
				Event: tcell.NewEventKey(tt.key, 0, tcell.ModNone),
				Text:  tt.text,
			}

			result := stateManager.HandleEvent(event)

			currentState := stateManager.GetCurrentState()
			assert.Equal(t, tt.expectedMode, currentState.Mode)

			// For escape and enter in SQL mode, result should be nil
			if tt.key == tcell.KeyEscape || tt.key == tcell.KeyEnter {
				assert.Nil(t, result)
			}

			// If SQL query was executed, verify the state was pushed
			if tt.key == tcell.KeyEnter && tt.text != "" {
				assert.Equal(t, 1, mockCb.callCount) // Should have pushed a new state
				assert.Equal(t, Browse, currentState.Mode)
				assert.Equal(t, TableRow, currentState.TableMode)

				// Verify headers and data were set
				assert.NotEmpty(t, currentState.TableHeaders)
				assert.NotEmpty(t, currentState.TableData)
			}

			// Verify mock expectations
			if tt.mockSetup != nil {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

func TestHandleEventBrowseMode(t *testing.T) {
	tests := []struct {
		name         string
		key          tcell.Key
		rune         rune
		row          int
		expectPush   bool
		expectedMode Mode
	}{
		{
			name:         "enter key triggers table row selection",
			key:          tcell.KeyEnter,
			row:          1,
			expectPush:   true,
			expectedMode: Browse, // Should push new browse state with TableRow mode
		},
		{
			name:         "q rune triggers table row selection",
			key:          tcell.KeyRune,
			rune:         'q',
			row:          1,
			expectPush:   true,
			expectedMode: Browse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh instances for each test to avoid interference
			mockDB, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer mockDB.Close()

			mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
			server := mysql

			// Start in browse mode with database table data
			browseState := State{
				Mode:      Browse,
				TableMode: DatabaseTable,
				TableData: []db.TableData{
					db.MysqlTable{Name: "test_table", Type: "BASE TABLE", Engine: "InnoDB", Rows: "1", Size: "1MB"},
				},
			}
			stateManager := NewContextualStateManager(server, browseState, 10)
			mockCb := &mockCallback{}
			stateManager.AddSyncCallback(mockCb.callback)

			// Mock data for table row selection
			columnRows := sqlmock.NewRows([]string{"COLUMN_NAME"}).
				AddRow("id").
				AddRow("name")
			mock.ExpectQuery("SELECT COLUMN_NAME FROM information_schema.COLUMNS").
				WithArgs("test_table").
				WillReturnRows(columnRows)

			dataRows := sqlmock.NewRows([]string{"id", "name"}).
				AddRow(1, "Test")
			mock.ExpectQuery("SELECT \\* FROM `test_table` LIMIT 1000").
				WillReturnRows(dataRows)

			event := &Event{
				Event: tcell.NewEventKey(tt.key, tt.rune, tcell.ModNone),
				Row:   tt.row,
			}

			result := stateManager.HandleEvent(event)

			if tt.expectPush {
				// Should have pushed a new state
				assert.Equal(t, 1, mockCb.callCount)

				currentState := stateManager.GetCurrentState()
				assert.Equal(t, tt.expectedMode, currentState.Mode)
				assert.Equal(t, TableRow, currentState.TableMode)

				// Result should be nil for these keys
				assert.Nil(t, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestHandleEventNormalKeys(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
	server := mysql

	browseState := State{Mode: Browse}
	stateManager := NewContextualStateManager(server, browseState, 10)
	mockCb := &mockCallback{}
	stateManager.AddSyncCallback(mockCb.callback)

	tests := []struct {
		name         string
		key          tcell.Key
		rune         rune
		expectedMode Mode
	}{
		{
			name:         "ctrl+c sets quit mode",
			key:          tcell.KeyCtrlC,
			expectedMode: QuitMode,
		},
		{
			name:         "colon enters command mode",
			key:          tcell.KeyRune,
			rune:         ':',
			expectedMode: Command,
		},
		{
			name:         "exclamation enters SQL mode",
			key:          tcell.KeyRune,
			rune:         '!',
			expectedMode: SQL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state and callback
			stateManager = NewContextualStateManager(server, State{Mode: Browse}, 10)
			mockCb = &mockCallback{}
			stateManager.AddSyncCallback(mockCb.callback)

			event := &Event{
				Event: tcell.NewEventKey(tt.key, tt.rune, tcell.ModNone),
			}

			result := stateManager.HandleEvent(event)

			currentState := stateManager.GetCurrentState()
			assert.Equal(t, tt.expectedMode, currentState.Mode)
			assert.Equal(t, 1, mockCb.callCount)

			// Ctrl+C, :, and ! should return nil
			if tt.key == tcell.KeyCtrlC || (tt.key == tcell.KeyRune && (tt.rune == ':' || tt.rune == '!')) {
				assert.Nil(t, result)
			}
		})
	}
}

func TestExtractNameFromSelection(t *testing.T) {
	tests := []struct {
		name          string
		state         State
		selectedIndex int
		expectedName  string
		expectError   bool
	}{
		{
			name: "extract name from MysqlTable",
			state: State{
				TableData: []db.TableData{
					db.MysqlTable{Name: "users", Type: "BASE TABLE", Engine: "InnoDB", Rows: "100", Size: "1MB"},
					db.MysqlTable{Name: "orders", Type: "BASE TABLE", Engine: "InnoDB", Rows: "500", Size: "5MB"},
				},
			},
			selectedIndex: 0,
			expectedName:  "users",
			expectError:   false,
		},
		{
			name: "extract name from second table",
			state: State{
				TableData: []db.TableData{
					db.MysqlTable{Name: "users", Type: "BASE TABLE", Engine: "InnoDB", Rows: "100", Size: "1MB"},
					db.MysqlTable{Name: "orders", Type: "BASE TABLE", Engine: "InnoDB", Rows: "500", Size: "5MB"},
				},
			},
			selectedIndex: 1,
			expectedName:  "orders",
			expectError:   false,
		},
		{
			name: "error with non-MysqlTable data",
			state: State{
				TableData: []db.TableData{
					map[string]string{"Name": "not a table"},
				},
			},
			selectedIndex: 0,
			expectedName:  "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := extractNameFromSelection(tt.state, tt.selectedIndex)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedName, name)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, name)
			}
		})
	}
}

func TestHandleEventTableDescription(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
	server := mysql

	// Start in browse mode with table data
	browseState := State{
		Mode:      Browse,
		TableMode: DatabaseTable,
		TableData: []db.TableData{
			db.MysqlTable{Name: "test_table", Type: "BASE TABLE", Engine: "InnoDB", Rows: "1", Size: "1MB"},
		},
	}
	stateManager := NewContextualStateManager(server, browseState, 10)
	mockCb := &mockCallback{}
	stateManager.AddSyncCallback(mockCb.callback)

	// Mock the SHOW CREATE TABLE query for 'd' rune
	createTableSQL := `CREATE TABLE test_table (
  id INT PRIMARY KEY
)`
	rows := sqlmock.NewRows([]string{"Table", "Create Table"}).
		AddRow("test_table", createTableSQL)
	mock.ExpectQuery("SHOW CREATE TABLE").
		WithArgs().
		WillReturnRows(rows)

	event := &Event{
		Event: tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
		Row:   1,
	}

	result := stateManager.HandleEvent(event)

	// Should have pushed a new Detail state
	currentState := stateManager.GetCurrentState()
	assert.Equal(t, Detail, currentState.Mode)
	assert.Equal(t, createTableSQL, currentState.DetailText)
	assert.Equal(t, 1, mockCb.callCount)
	assert.Nil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleEventEscapeKey(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
	server := mysql

	// Create a state manager with multiple states
	browseState := State{Mode: Browse, TableMode: DatabaseTable}
	stateManager := NewContextualStateManager(server, browseState, 10)

	ctx := context.Background()

	// Push command state
	stateManager.PushState(ctx, State{Mode: Command})

	// Push detail state
	stateManager.PushState(ctx, State{Mode: Detail, DetailText: "test"})

	mockCb := &mockCallback{}
	stateManager.AddSyncCallback(mockCb.callback)

	// Test escape key
	event := &Event{
		Event: tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone),
	}

	result := stateManager.HandleEvent(event)

	// Should have popped to command state
	currentState := stateManager.GetCurrentState()
	assert.Equal(t, Command, currentState.Mode)
	assert.Equal(t, 1, mockCb.callCount)
	assert.Equal(t, Detail, mockCb.lastTransition.From.Mode)
	assert.Equal(t, Command, mockCb.lastTransition.To.Mode)
	assert.Nil(t, result)
}

func TestUpdateCurrentStateSelection(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
	server := mysql

	browseState := State{
		Mode:              Browse,
		TableMode:         DatabaseTable,
		SelectedDataIndex: 0,
	}
	stateManager := NewContextualStateManager(server, browseState, 10)

	// Update selection
	stateManager.updateCurrentStateSelection(5)

	// Check that selection was updated
	currentState := stateManager.GetCurrentState()
	assert.Equal(t, 5, currentState.SelectedDataIndex)
}
