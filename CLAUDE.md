# CLAUDE.md - Repository Analysis

## Overview
**rel8** is a terminal-based database browser and management tool written in Go. It provides an interactive TUI (Terminal User Interface) for browsing databases, tables, and executing SQL queries across multiple database types (MySQL, PostgreSQL, SQLite).

**Key Features:**
- Multi-database support (MySQL, PostgreSQL, SQLite)
- Interactive terminal interface using tview
- **Hierarchical tree view** for database structure navigation (server → databases → categories → items)
- SQL syntax highlighting and execution
- MySQL/PostgreSQL server information display (version, user, host, port, database, max_connections, buffer pool size)
- **Manual keyboard navigation** with arrow keys and tree expansion/collapse
- Demo mode with scripted interactions
- Mock data mode for development/testing
- State-driven architecture with contextual navigation

## Repository Structure

### Core Architecture
```
rel8/
   main.go                 # Application entry point and demo orchestration
   model/                  # State management and business logic
      model.go           # Core state definitions and modes
      state_manager.go   # Event handling and state transitions
   view/                   # UI components and presentation layer
      view.go            # Main view coordinator with tree navigation logic
      tree.go            # Hierarchical tree view component for database structure
      grid.go            # Table display component
      detail.go          # Detail view component
      editor.go          # SQL editor component
      header.go          # Header display with context info
      keys.go            # Key bindings display system
      command.go         # Command input component
      color.go           # Centralized color theming
   db/                     # Database abstraction layer
      db.go              # Core database interface with tree-specific methods
      mysql.go           # MySQL-specific implementation
      mysql8.go          # MySQL 8+ compatibility
      postgres.go        # PostgreSQL-specific implementation
      postgres_mock.go   # PostgreSQL mock data for development/testing
      sql.go             # SQL syntax highlighting
   config/                 # Configuration and initialization 
       init.go            # CLI parsing and setup
       art.go             # ASCII art for header
```

### File Statistics
- **Total Go files:** 36
- **Total lines of code:** ~6,500
- **Test coverage:** 100% pass rate (53+ tests in view package, 160+ tests total across all packages)

## Technology Stack

### Core Dependencies
- **Go 1.24.2** - Primary language
- **github.com/rivo/tview v0.0.0-20250625164341** - Terminal UI framework
- **github.com/gdamore/tcell/v2 v2.8.1** - Terminal cell manipulation
- **github.com/spf13/viper v1.20.1** - Configuration management

### Database Drivers
- **github.com/go-sql-driver/mysql v1.9.3** - MySQL connectivity
- **github.com/jackc/pgx/v5 v5.7.5** - PostgreSQL connectivity
- **github.com/mattn/go-sqlite3 v1.14.30** - SQLite connectivity

### Testing & Development
- **github.com/DATA-DOG/go-sqlmock v1.5.2** - Database mocking
- **github.com/stretchr/testify v1.10.0** - Testing framework

## Architecture Overview

### State Management System
The application uses a sophisticated state management system (`model/state_manager.go:14`) with the following modes:

```go
type Mode int
const (
    Browse Mode = iota  // Database/table browsing
    Command            // Command input mode (:quit, :table, etc.)
    SQL               // SQL query input mode
    Detail            // Detail view for table descriptions
    Editor            // SQL editor mode
    QuitMode Mode = -1 // Application exit
)
```

### Key Components

#### 1. ContextualStateManager (`model/state_manager.go:14`)
- Manages application state transitions
- Handles keyboard events and routing
- Maintains state history stack (20 levels deep)
- Provides callback system for UI updates

#### 2. Tree View System (`view/tree.go`)
- **Hierarchical database structure** navigation (server → databases → categories → items)
- **Manual keyboard navigation** with custom tree traversal algorithms
- **Database-specific node types**: server, database, category (Tables/Views/Procedures/Functions/Triggers), items
- **Lazy loading** of category items when expanded
- **Multi-database support** with PostgreSQL and MySQL implementations

#### 3. View System (`view/view.go`)
- Coordinates UI components based on current state
- **Advanced navigation methods**: `findNextNode`, `findPrevNode`, `findNextSibling`, `findPrevSibling`, `findParent`, `findLastDescendant`
- Handles mode-specific rendering (tree view as default, grid view for table data)
- Manages component focus and input routing

#### 4. Database Abstraction (`db/db.go`)
- Multi-database support with unified interface
- **Tree-specific database methods**: `FetchTablesForDatabase`, `FetchViewsForDatabase`, `FetchProceduresForDatabase`, `FetchFunctionsForDatabase`, `FetchTriggersForDatabase`
- Automatic driver detection from connection strings
- MySQL/PostgreSQL server information extraction (version, user, configuration)
- Mock data support for development with comprehensive test data
- SQL syntax highlighting with 50+ keywords

#### 5. Color Theming (`view/color.go`)
- Centralized color management system
- Configurable color schemes
- Support for both tcell.Color and tview color tags
- **Tree-specific colors**: `TreeRootColor`, `TreeDatabaseColor`, `TreeCategoryColor`, `TreeItemColor`
- Selection highlighting integrated with tview's native system

## Tree View Navigation and Grid Selection

### Tree View Navigation (`view/tree.go`)
- **Hierarchical structure**: Server → Databases → Categories (Tables/Views/Procedures/Functions/Triggers) → Items
- **Manual keyboard navigation**: Custom tree traversal algorithms handle arrow key navigation
- **Node expansion**: Enter/Space keys expand/collapse tree nodes with lazy loading
- **Selection highlighting**: Uses tview's native selection system with custom colors
- **Database tracking**: Automatically updates current database context when database nodes are selected

### Grid Selection and Highlighting (`view/grid.go`)
- Table configured for rows-only selection via `SetSelectable(true, false)`
- Initial selection starts at first data row; selection is restored on state transitions
- `RefreshSelectionHighlight()` styles the selected row's cells (foreground/background) and resets others
- For the selected row only, missing trailing cells up to `headerCount` are synthesized so the last logical column exists and receives highlight

### Full-width selection band (Grid only)
- A full-width band is painted after draw to cover any right-side gap beyond the last cell
- `Grid.DrawSelectionBand(screen)` sets only the background for the selected screen row, preserving runes and foreground colors
- `Grid.AttachSelectionBand(app, shouldDraw)` chains into the app's after-draw and paints the band when `shouldDraw()` is true. We enable this in Browse, Command, and SQL modes
- Row alignment accounts for scroll via `getRowOffsetUnsafe()`

### Column expansion (Grid only)
- Default expansion is uniform (`SetExpansion(1)`) for predictable layout
- After populate, `StretchLastColumn()` increases expansion on the last existing column to better absorb leftover width, complementing the band

## Key Workflows

### Application Startup
1. **Configuration** (`config/init.go:15`) - Parse CLI flags and environment
2. **Database Connection** (`db/db.go:25`) - Establish database connection or mock
3. **State Initialization** (`main.go:14`) - Create state manager with initial state
4. **View Setup** (`main.go:15`) - Initialize UI components
5. **Event Loop** (`view/view.go:45`) - Start tview application

### State Transitions
- **`:` key**  Command mode for database operations
- **`!` key**  SQL query mode
- **`s` key**  Editor mode (when not in Command/SQL mode)
- **`F5` key**  Execute SQL in Editor mode
- **`Escape` key**  Return to previous state
- **`Ctrl+C` key**  Quit application

### Tree Navigation (Browse Mode)
- **`Arrow Up/Down`** Navigate between visible tree nodes in hierarchy order
- **`Enter/Space`** Expand/collapse tree nodes (databases show categories, categories load and show items)
- **Tree traversal** follows proper depth-first order with sibling navigation
- **Lazy loading** fetches table/view/procedure data only when category is expanded

### Database Operations
1. **Tree Navigation** - Browse hierarchical database structure (server → databases → categories → items)
2. **Database Expansion** - Expand databases to show categories: Tables, Views, Procedures, Functions, Triggers
3. **Category Expansion** - Expand categories to show individual items with lazy loading
4. **Table Data Display** - Click on table items to show table data with pagination in grid view
5. **Table Description** - Show CREATE TABLE statements in detail view
6. **SQL Execution** - Execute custom SQL queries in editor mode
7. **Server Information** - Display MySQL/PostgreSQL server details in the header

## Development Patterns

### Testing Strategy
- **Unit Tests:** Comprehensive coverage for all packages
- **Mock Testing:** Database operations use sqlmock for isolation with comprehensive mock data for MySQL and PostgreSQL
- **Integration Tests:** State transitions and component interactions
- **Tree Navigation Tests:** Comprehensive testing of tree traversal algorithms and navigation methods
- **Alignment Tests:** UI component spacing and formatting verification

### Error Handling
- Graceful degradation to mock data when database operations fail
- Comprehensive error logging with contextual information
- User-friendly error messages in UI components

### Code Organization
- **Package Separation:** Clear boundaries between model, view, and data layers
- **Interface Usage:** Database operations abstracted behind interfaces
- **Configuration:** Centralized in dedicated config package
- **Color Management:** All UI colors centralized for easy theming

## Testing & Quality Assurance

### Test Coverage
```
rel8         : All tests - PASS
rel8/config  : 4 tests   - PASS  
rel8/db      : 12+ tests - PASS (includes PostgreSQL mock tests)
rel8/model   : 15+ tests - PASS
rel8/view    : 53+ tests - PASS (includes comprehensive tree navigation tests)
```

### Key Test Areas
- **State Management:** Event handling and transitions
- **Database Operations:** CRUD operations with error scenarios for MySQL and PostgreSQL
- **Tree Navigation:** Complete testing of tree traversal algorithms, node finding, and navigation methods
- **UI Components:** Alignment, formatting, and interaction
- **SQL Highlighting:** Keyword recognition and syntax coloring
- **Color System:** Theme application and consistency including tree-specific colors

## Common Development Tasks

### Adding New Database Support
1. Implement driver detection logic in `db/db.go:DetermineDriver`
2. Add connection string parsing
3. Create database-specific implementation file (e.g., `postgres.go`)
4. Implement tree-specific methods: `FetchTablesForDatabase`, `FetchViewsForDatabase`, `FetchProceduresForDatabase`, `FetchFunctionsForDatabase`, `FetchTriggersForDatabase`
5. Create corresponding mock implementation for testing
6. Handle database-specific SQL variations
7. Add comprehensive tests including tree navigation scenarios

### Adding New UI Components
1. Create component in `view/` package
2. Follow existing patterns (border, padding, color usage)
3. Use centralized colors from `view/color.go` (add new colors if needed: `TreeXXXColor` pattern)
4. Consider navigation requirements and input handling
5. Add comprehensive alignment tests and navigation tests if applicable
6. Integrate into main view coordinator with proper state transitions

### Adding New Key Bindings
1. Update key handling in `model/state_manager.go`
2. Add key documentation in `view/keys.go:getDefaultKeyPairs`
3. Test state transitions
4. Update help documentation

### Modifying Color Scheme
1. Update `view/color.go:DefaultColors`
2. Ensure consistency across all components
3. Test color visibility and contrast
4. Run alignment tests to verify formatting

## Demo & Testing Features

### Demo Mode
- **Scripted Interactions:** Automated keypress sequences
- **File-Based Scripts:** Load demo scripts from files
- **Development Testing:** Rapid UI testing without manual interaction
- **Documentation:** Comprehensive examples in `demo.md`

### Mock Data Mode
- **Database Independence:** Run without real database connections
- **Development Speed:** Fast iteration during development
- **Testing Reliability:** Consistent test data
- **CI/CD Friendly:** No external dependencies

## Performance Considerations

### Memory Management
- **State History:** Limited to 20 levels to prevent memory leaks
- **Table Data:** Efficient handling of large result sets
- **UI Components:** Proper cleanup and resource management

### Database Efficiency
- **Connection Pooling:** Handled by database drivers
- **Query Optimization:** Efficient metadata queries
- **Error Recovery:** Graceful handling of connection issues

## AI Development Guidelines

### Best Practices for AI Assistance
1. **Maintain Patterns:** Follow existing code organization and naming conventions
2. **Test Coverage:** Always add tests for new functionality
3. **Color Consistency:** Use centralized color system for all UI changes
4. **State Management:** Understand state transition implications
5. **Documentation:** Update relevant documentation for significant changes

### Areas Requiring Special Attention
- **Tree Navigation:** Complex tree traversal algorithms and manual keyboard navigation
- **State Transitions:** Complex interaction between modes (tree view vs. grid view)
- **UI Alignment:** Precise formatting requirements for terminal display
- **Database Compatibility:** Multi-database support considerations (MySQL vs. PostgreSQL vs. SQLite)
- **Color Theming:** Centralized system affects entire application including tree-specific colors
- **Input Handling:** Proper event flow between global handlers and tree navigation
- **Testing:** Comprehensive coverage expectations including tree navigation scenarios

### Common Pitfalls to Avoid
- **Direct Color Usage:** Always use `Colors` global instance, especially for tree colors
- **Navigation Conflicts:** Avoid conflicting input handlers between global and tree navigation
- **State Mutation:** Follow proper state transition patterns (tree view vs. grid view modes)
- **Tree Traversal:** Use existing navigation helper methods rather than reimplementing tree walking
- **Test Isolation:** Ensure tests don't affect each other, especially tree navigation tests
- **Mock Data:** Use appropriate mock implementations (MySQL vs. PostgreSQL) in tests
- **UI Spacing:** Maintain consistent alignment and padding
- **Error Handling:** Provide graceful degradation paths to mock data

---
*Generated on 2025-08-12 by Claude Code for comprehensive repository understanding and AI-assisted development.*  
*Updated on 2025-08-16 with tree view implementation, PostgreSQL support, and advanced navigation features.*