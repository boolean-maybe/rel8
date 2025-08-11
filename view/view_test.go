package view

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"rel8/db"
	"rel8/model"
)

func TestNewView(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
	server := mysql

	initialState := model.State{Mode: model.Browse}
	stateManager := model.NewContextualStateManager(server, initialState, 10)
	view := NewView(stateManager)

	assert.NotNil(t, view)
	assert.Equal(t, stateManager, view.stateManager)
	assert.Equal(t, model.Initial, view.model) // Still has model pointer for current state
	assert.NotNil(t, view.App)
	assert.NotNil(t, view.flex)
	assert.NotNil(t, view.header)
	assert.NotNil(t, view.grid)
	assert.NotNil(t, view.commandBar)
}

func TestViewOnStateTransition(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
	server := mysql

	initialState := model.State{Mode: model.Browse}
	stateManager := model.NewContextualStateManager(server, initialState, 10)
	view := NewView(stateManager)

	tests := []struct {
		name      string
		fromState model.State
		toState   model.State
	}{
		{
			name:      "browse mode transition",
			fromState: model.State{Mode: model.Command},
			toState: model.State{
				Mode:         model.Browse,
				TableMode:    model.DatabaseTable,
				TableHeaders: []string{"Name", "Type"},
				TableData: []db.TableData{
					map[string]string{"Name": "table1", "Type": "BASE TABLE"},
				},
			},
		},
		{
			name:      "command mode transition",
			fromState: model.State{Mode: model.Browse},
			toState: model.State{
				Mode:         model.Command,
				TableMode:    model.EmptyTable,
				TableHeaders: []string{},
				TableData:    []db.TableData{},
			},
		},
		{
			name:      "SQL mode transition",
			fromState: model.State{Mode: model.Browse},
			toState: model.State{
				Mode:         model.SQL,
				TableMode:    model.EmptyTable,
				TableHeaders: []string{},
				TableData:    []db.TableData{},
			},
		},
		{
			name:      "detail mode transition",
			fromState: model.State{Mode: model.Browse},
			toState: model.State{
				Mode:       model.Detail,
				DetailText: "CREATE TABLE test (id INT)",
			},
		},
		{
			name:      "quit mode transition",
			fromState: model.State{Mode: model.Browse},
			toState:   model.State{Mode: model.QuitMode},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transition := model.StateTransition{
				From: tt.fromState,
				To:   tt.toState,
			}

			// This should not panic and should update the view
			assert.NotPanics(t, func() {
				view.OnStateTransition(transition)
			})

			// Verify the view's model pointer was updated
			assert.Equal(t, &tt.toState, view.model)

			// Verify table was populated correctly if in browse mode
			if tt.toState.Mode == model.Browse {
				rowCount := view.grid.GetRowCount()
				colCount := view.grid.GetColumnCount()

				expectedRows := len(tt.toState.TableData) + 1 // +1 for header
				expectedCols := len(tt.toState.TableHeaders)

				assert.Equal(t, expectedRows, rowCount)
				if len(tt.toState.TableHeaders) > 0 {
					assert.Equal(t, expectedCols, colCount)
				}
			}

			// Verify details view was updated if in detail mode
			if tt.toState.Mode == model.Detail {
				assert.NotNil(t, view.details)
				detailsText := view.details.GetText(false)
				// The displayed text should be the SQL-highlighted version
				expectedHighlighted := db.HighlightSQL(tt.toState.DetailText)
				assert.Equal(t, expectedHighlighted, detailsText)
			}
		})
	}
}

func TestViewOnStateTransitionDetailMode(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
	server := mysql

	initialState := model.State{Mode: model.Browse}
	stateManager := model.NewContextualStateManager(server, initialState, 10)
	view := NewView(stateManager)

	// Test detail mode transition
	transition := model.StateTransition{
		From: model.State{Mode: model.Browse},
		To:   model.State{Mode: model.Detail, DetailText: "CREATE TABLE test (id INT)"},
	}

	// This should not panic and should update the view to detail mode
	assert.NotPanics(t, func() {
		view.OnStateTransition(transition)
	})

	// Verify the details TextView was updated with SQL highlighting
	assert.NotNil(t, view.details)
	detailsText := view.details.GetText(false)
	expectedHighlighted := "[lightblue]CREATE[-] [lightblue]TABLE[-] test (id [lightgreen]INT[-])"
	assert.Equal(t, expectedHighlighted, detailsText)

	// Verify the view's model pointer was updated
	assert.Equal(t, model.Detail, view.model.Mode)
	assert.Equal(t, "CREATE TABLE test (id INT)", view.model.DetailText)
}
