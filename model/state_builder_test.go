package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStateBuilder(t *testing.T) {
	builder := NewStateBuilder()

	assert.NotNil(t, builder)
	assert.False(t, builder.hasBrowse)
	assert.Nil(t, builder.browseState)
	assert.False(t, builder.hasCommand)
	assert.Nil(t, builder.commandState)
	assert.False(t, builder.hasSql)
	assert.Nil(t, builder.sqlState)
	assert.False(t, builder.hasTree)
	assert.Nil(t, builder.treeState)
	assert.False(t, builder.hasDetail)
	assert.Nil(t, builder.detailState)
	assert.False(t, builder.hasFullSql)
	assert.Nil(t, builder.fullSqlState)
}

func TestNewStateBuilderFromState(t *testing.T) {
	// Create an original state with all sub-states
	originalBuilder := NewStateBuilder().
		SetBrowse(&BrowseState{BrowseClass: DatabaseTable}).
		SetCommand(&CommandState{Text: "test command"}).
		SetSql(&SqlState{Sql: "SELECT * FROM test"}).
		SetTree(&TreeState{Text: "node1"}).
		SetDetail(&DetailState{Text: "detail text"}).
		SetFullSql(&FullSqlState{Sql: "editor content"})

	originalState := originalBuilder.Build()

	// Create a new builder from the original state
	copyBuilder := NewStateBuilderFromState(originalState)

	// Verify all states were copied
	assert.True(t, copyBuilder.hasBrowse)
	assert.NotNil(t, copyBuilder.browseState)
	assert.Equal(t, DatabaseTable, copyBuilder.browseState.BrowseClass)

	assert.True(t, copyBuilder.hasCommand)
	assert.NotNil(t, copyBuilder.commandState)
	assert.Equal(t, "test command", copyBuilder.commandState.Text)

	assert.True(t, copyBuilder.hasSql)
	assert.NotNil(t, copyBuilder.sqlState)
	assert.Equal(t, "SELECT * FROM test", copyBuilder.sqlState.Sql)

	assert.True(t, copyBuilder.hasTree)
	assert.NotNil(t, copyBuilder.treeState)
	assert.Equal(t, "node1", copyBuilder.treeState.Text)

	assert.True(t, copyBuilder.hasDetail)
	assert.NotNil(t, copyBuilder.detailState)
	assert.Equal(t, "detail text", copyBuilder.detailState.Text)

	assert.True(t, copyBuilder.hasFullSql)
	assert.NotNil(t, copyBuilder.fullSqlState)
	assert.Equal(t, "editor content", copyBuilder.fullSqlState.Sql)
}

func TestNewStateBuilderFromStatePartial(t *testing.T) {
	// Create a state with only some sub-states
	originalBuilder := NewStateBuilder().
		SetBrowse(&BrowseState{BrowseClass: Database}).
		SetCommand(&CommandState{Text: "test command"})

	originalState := originalBuilder.Build()

	// Create a new builder from the partial state
	copyBuilder := NewStateBuilderFromState(originalState)

	// Verify only the present states were copied
	assert.True(t, copyBuilder.hasBrowse)
	assert.NotNil(t, copyBuilder.browseState)
	assert.Equal(t, Database, copyBuilder.browseState.BrowseClass)

	assert.True(t, copyBuilder.hasCommand)
	assert.NotNil(t, copyBuilder.commandState)
	assert.Equal(t, "test command", copyBuilder.commandState.Text)

	// Verify absent states remain false/nil
	assert.False(t, copyBuilder.hasSql)
	assert.Nil(t, copyBuilder.sqlState)
	assert.False(t, copyBuilder.hasTree)
	assert.Nil(t, copyBuilder.treeState)
	assert.False(t, copyBuilder.hasDetail)
	assert.Nil(t, copyBuilder.detailState)
	assert.False(t, copyBuilder.hasFullSql)
	assert.Nil(t, copyBuilder.fullSqlState)
}

func TestNewStateBuilderFromStateEmpty(t *testing.T) {
	// Create an empty state
	originalBuilder := NewStateBuilder()
	originalState := originalBuilder.Build()

	// Create a new builder from the empty state
	copyBuilder := NewStateBuilderFromState(originalState)

	// Verify all states remain false/nil
	assert.False(t, copyBuilder.hasBrowse)
	assert.Nil(t, copyBuilder.browseState)
	assert.False(t, copyBuilder.hasCommand)
	assert.Nil(t, copyBuilder.commandState)
	assert.False(t, copyBuilder.hasSql)
	assert.Nil(t, copyBuilder.sqlState)
	assert.False(t, copyBuilder.hasTree)
	assert.Nil(t, copyBuilder.treeState)
	assert.False(t, copyBuilder.hasDetail)
	assert.Nil(t, copyBuilder.detailState)
	assert.False(t, copyBuilder.hasFullSql)
	assert.Nil(t, copyBuilder.fullSqlState)
}

func TestNewStateBuilderFromStateBuild(t *testing.T) {
	// Create an original state
	originalBuilder := NewStateBuilder().
		SetBrowse(&BrowseState{BrowseClass: DatabaseTable}).
		SetTree(&TreeState{Text: "node1"})

	originalState := originalBuilder.Build()

	// Create a copy builder and build a new state
	copyBuilder := NewStateBuilderFromState(originalState)
	newState := copyBuilder.Build()

	// Verify the new state has the same data
	assert.True(t, newState.HasBrowse())
	assert.Equal(t, DatabaseTable, newState.GetBrowseState().BrowseClass)

	assert.True(t, newState.HasTree())
	assert.Equal(t, "node1", newState.GetTreeState().Text)

	// Verify absent states
	assert.False(t, newState.HasCommand())
	assert.False(t, newState.HasSql())
	assert.False(t, newState.HasDetail())
	assert.False(t, newState.HasFullSql())
}
