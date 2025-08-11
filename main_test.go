package main

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"rel8/db"
	"rel8/model"
	"rel8/view"
)

// TestMainComponents verifies that the main components can be created without panicking
func TestMainComponents(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	// Test state manager creation
	assert.NotPanics(t, func() {
		mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
		server := mysql

		initialState := model.State{Mode: model.Browse}
		stateManager := model.NewContextualStateManager(server, initialState, 10)
		assert.NotNil(t, stateManager)
	})

	// Test view creation
	assert.NotPanics(t, func() {
		mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
		server := mysql

		initialState := model.State{Mode: model.Browse}
		stateManager := model.NewContextualStateManager(server, initialState, 10)
		v := view.NewView(stateManager)
		assert.NotNil(t, v)
	})
}

// TestIntegrationStateManagerView tests the integration between state manager and view
func TestIntegrationStateManagerView(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	// Create state manager and view
	mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
	server := mysql

	initialState := model.State{Mode: model.Browse}
	stateManager := model.NewContextualStateManager(server, initialState, 10)
	v := view.NewView(stateManager)

	// Test that view is created successfully (we can't directly test private fields)
	assert.NotNil(t, v)

	// Test initial state
	currentState := stateManager.GetCurrentState()
	assert.Equal(t, model.Browse, currentState.Mode)

	// Test state transitions
	ctx := context.Background()

	commandState := model.State{Mode: model.Command}
	err = stateManager.PushState(ctx, commandState)
	assert.NoError(t, err)

	currentState = stateManager.GetCurrentState()
	assert.Equal(t, model.Command, currentState.Mode)

	// Test state popping
	previousState, err := stateManager.PopState(ctx)
	assert.NoError(t, err)
	assert.Equal(t, model.Browse, previousState.Mode)

	currentState = stateManager.GetCurrentState()
	assert.Equal(t, model.Browse, currentState.Mode)

	// Mock data for testing database operations
	rows := sqlmock.NewRows([]string{"TABLE_NAME", "TABLE_TYPE", "ENGINE", "TABLE_ROWS", "SIZE_MB"}).
		AddRow("users", "BASE TABLE", "InnoDB", 100, 1.5)
	mock.ExpectQuery("SELECT TABLE_NAME").WillReturnRows(rows)

	// Test that database operations work with mocked data
	ctx = context.Background()
	headers, data := server.FetchTables(ctx)
	assert.NotEmpty(t, headers)
	assert.NotEmpty(t, data)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestCallbackIntegration tests that callbacks work between state manager and view
func TestCallbackIntegration(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	mysql := &db.Mysql8{db.Mysql{DbInstance: mockDB}}
	server := mysql

	initialState := model.State{Mode: model.Browse}
	stateManager := model.NewContextualStateManager(server, initialState, 10)
	v := view.NewView(stateManager)

	// Track callback calls
	callbackCount := 0
	stateManager.AddSyncCallback(func(transition model.StateTransition) {
		callbackCount++
		// Simulate what happens in main.go - view gets notified
		v.OnStateTransition(transition)
	})

	ctx := context.Background()

	// Test state change triggers callback
	commandState := model.State{Mode: model.Command}
	err = stateManager.PushState(ctx, commandState)
	assert.NoError(t, err)

	// Verify callback was called
	assert.Equal(t, 1, callbackCount)

	// Verify view was updated - we can't directly access private fields,
	// but the fact that the callback was called indicates the integration works
}
