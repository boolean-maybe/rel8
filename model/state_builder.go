package model

type StateBuilder struct {
	commonState  *CommonState
	hasBrowse    bool
	browseState  *BrowseState
	hasCommand   bool
	commandState *CommandState
	hasSql       bool
	sqlState     *SqlState
	hasTree      bool
	treeState    *TreeState
	hasDetail    bool
	detailState  *DetailState
	hasFullSql   bool
	fullSqlState *FullSqlState
}

func NewStateBuilder() *StateBuilder {
	return &StateBuilder{
		commonState:  nil,
		hasBrowse:    false,
		browseState:  nil,
		hasCommand:   false,
		commandState: nil,
		hasSql:       false,
		sqlState:     nil,
		hasTree:      false,
		treeState:    nil,
		hasDetail:    false,
		detailState:  nil,
		hasFullSql:   false,
		fullSqlState: nil,
	}
}

func NewStateBuilderFromState(state State) *StateBuilder {
	builder := &StateBuilder{}

	// Copy the CommonState (header info)
	headerInfo := state.GetHeaderInfo()
	if headerInfo != nil {
		builder.commonState = &CommonState{
			HeaderInfo: headerInfo,
		}
	}

	if state.HasBrowse() {
		browseState := state.GetBrowseState()
		builder.hasBrowse = true
		builder.browseState = &browseState
	}

	if state.HasCommand() {
		commandState := state.GetCommandState()
		builder.hasCommand = true
		builder.commandState = &commandState
	}

	if state.HasSql() {
		sqlState := state.GetSqlState()
		builder.hasSql = true
		builder.sqlState = &sqlState
	}

	if state.HasTree() {
		treeState := state.GetTreeState()
		builder.hasTree = true
		builder.treeState = &treeState
	}

	if state.HasDetail() {
		detailState := state.GetDetailState()
		builder.hasDetail = true
		builder.detailState = &detailState
	}

	if state.HasFullSql() {
		fullSqlState := state.GetFullSqlState()
		builder.hasFullSql = true
		builder.fullSqlState = &fullSqlState
	}

	return builder
}

func (b *StateBuilder) Build() State {
	return &ConcreteState{
		CommonState:  b.commonState,
		hasBrowse:    b.hasBrowse,
		browseState:  b.browseState,
		hasCommand:   b.hasCommand,
		commandState: b.commandState,
		hasSql:       b.hasSql,
		sqlState:     b.sqlState,
		hasTree:      b.hasTree,
		treeState:    b.treeState,
		hasDetail:    b.hasDetail,
		detailState:  b.detailState,
		hasFullSql:   b.hasFullSql,
		fullSqlState: b.fullSqlState,
	}
}

func (b *StateBuilder) SetCommon(state *CommonState) *StateBuilder {
	b.commonState = state
	return b
}

func (b *StateBuilder) SetBrowse(state *BrowseState) *StateBuilder {
	b.hasBrowse = true
	b.browseState = state
	return b
}

func (b *StateBuilder) SetCommand(state *CommandState) *StateBuilder {
	b.hasCommand = true
	b.commandState = state
	return b
}

func (b *StateBuilder) SetEmptyCommand() *StateBuilder {
	state := &CommandState{
		Text: "",
	}
	b.hasCommand = true
	b.commandState = state
	return b
}

func (b *StateBuilder) SetSql(state *SqlState) *StateBuilder {
	b.hasSql = true
	b.sqlState = state
	return b
}

func (b *StateBuilder) SetEmptySql() *StateBuilder {
	state := &SqlState{
		Sql: "",
	}
	b.hasSql = true
	b.sqlState = state
	return b
}

func (b *StateBuilder) SetTree(state *TreeState) *StateBuilder {
	b.hasTree = true
	b.treeState = state
	return b
}

func (b *StateBuilder) SetDetail(state *DetailState) *StateBuilder {
	b.hasDetail = true
	b.detailState = state
	return b
}

func (b *StateBuilder) SetFullSql(state *FullSqlState) *StateBuilder {
	b.hasFullSql = true
	b.fullSqlState = state
	return b
}

type ConcreteState struct {
	*CommonState
	hasBrowse    bool
	browseState  *BrowseState
	hasCommand   bool
	commandState *CommandState
	hasSql       bool
	sqlState     *SqlState
	hasTree      bool
	treeState    *TreeState
	hasDetail    bool
	detailState  *DetailState
	hasFullSql   bool
	fullSqlState *FullSqlState
}

func (s *ConcreteState) HasCommand() bool {
	return s.hasCommand
}

func (s *ConcreteState) GetCommandState() CommandState {
	if s.commandState != nil {
		return *s.commandState
	}
	return CommandState{}
}

func (s *ConcreteState) HasBrowse() bool {
	return s.hasBrowse
}

func (s *ConcreteState) GetBrowseState() BrowseState {
	if s.browseState != nil {
		return *s.browseState
	}
	return BrowseState{}
}

func (s *ConcreteState) HasSql() bool {
	return s.hasSql
}

func (s *ConcreteState) GetSqlState() SqlState {
	if s.sqlState != nil {
		return *s.sqlState
	}
	return SqlState{}
}

func (s *ConcreteState) HasTree() bool {
	return s.hasTree
}

func (s *ConcreteState) GetTreeState() TreeState {
	if s.treeState != nil {
		return *s.treeState
	}
	return TreeState{}
}

func (s *ConcreteState) HasDetail() bool {
	return s.hasDetail
}

func (s *ConcreteState) GetDetailState() DetailState {
	if s.detailState != nil {
		return *s.detailState
	}
	return DetailState{}
}

func (s *ConcreteState) HasFullSql() bool {
	return s.hasFullSql
}

func (s *ConcreteState) GetFullSqlState() FullSqlState {
	if s.fullSqlState != nil {
		return *s.fullSqlState
	}
	return FullSqlState{}
}
