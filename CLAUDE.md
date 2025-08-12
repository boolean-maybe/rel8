# CLAUDE.md - Repository Analysis

## Overview
**rel8** is a terminal-based database browser and management tool written in Go. It provides an interactive TUI (Terminal User Interface) for browsing databases, tables, and executing SQL queries across multiple database types (MySQL, PostgreSQL, SQLite).

**Key Features:**
- Multi-database support (MySQL, PostgreSQL, SQLite)
- Interactive terminal interface using tview
- SQL syntax highlighting and execution
- MySQL server information display (version, user, host, port, database, max_connections, buffer pool size)
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
      view.go            # Main view coordinator
      grid.go            # Table display component
      detail.go          # Detail view component
      editor.go          # SQL editor component
      header.go          # Header display with context info
      keys.go            # Key bindings display system
      command.go         # Command input component
      color.go           # Centralized color theming
   db/                     # Database abstraction layer
      db.go              # Core database interface
      mysql.go           # MySQL-specific implementation
      mysql8.go          # MySQL 8+ compatibility
      sql.go             # SQL syntax highlighting
   config/                 # Configuration and initialization 
       init.go            # CLI parsing and setup
       art.go             # ASCII art for header
```

### File Statistics
- **Total Go files:** 33
- **Total lines of code:** 5,669
- **Test coverage:** 100% pass rate (113 tests across all packages)

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

#### 2. View System (`view/view.go`)
- Coordinates UI components based on current state
- Handles mode-specific rendering
- Manages component focus and input routing

#### 3. Database Abstraction (`db/db.go`)
- Multi-database support with unified interface
- Automatic driver detection from connection strings
- MySQL server information extraction (version, user, configuration)
- Mock data support for development
- SQL syntax highlighting with 50+ keywords

#### 4. Color Theming (`view/color.go`)
- Centralized color management system
- Configurable color schemes
- Support for both tcell.Color and tview color tags
- Selection band color exposed as `Colors.SelectionBandBg`

## Row Selection, Full-Width Highlighting, and Column Expansion

### Selection behavior
- Table configured for rows-only selection via `SetSelectable(true, false)` (`view/grid.go`).
- Initial selection starts at first data row; selection is restored on state transitions.

### Per-cell selection styling
- `RefreshSelectionHighlight()` styles the selected row's cells (foreground/background) and resets others.
- For the selected row only, missing trailing cells up to `headerCount` are synthesized so the last logical column exists and receives highlight.

### Full-width selection band
- A full-width band is painted after draw to cover any right-side gap beyond the last cell.
- `Grid.DrawSelectionBand(screen)` sets only the background for the selected screen row, preserving runes and foreground colors.
- `Grid.AttachSelectionBand(app, shouldDraw)` chains into the app's after-draw and paints the band when `shouldDraw()` is true. We enable this in Browse, Command, and SQL modes.
- Row alignment accounts for scroll via `getRowOffsetUnsafe()`.

### Column expansion
- Default expansion is uniform (`SetExpansion(1)`) for predictable layout.
- After populate, `StretchLastColumn()` increases expansion on the last existing column to better absorb leftover width, complementing the band.

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

### Database Operations
1. **Browse Databases** - List available databases
2. **Browse Tables** - Show tables in selected database
3. **Table Rows** - Display table data with pagination
4. **Table Description** - Show CREATE TABLE statements
5. **SQL Execution** - Execute custom SQL queries
6. **Server Information** - Display MySQL server details in the header

## Development Patterns

### Testing Strategy
- **Unit Tests:** Comprehensive coverage for all packages
- **Mock Testing:** Database operations use sqlmock for isolation
- **Integration Tests:** State transitions and component interactions
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
rel8         : 113 tests - PASS
rel8/config  : 4 tests   - PASS  
rel8/db      : 12 tests  - PASS
rel8/model   : 15 tests  - PASS
rel8/view    : 29 tests  - PASS
```

### Key Test Areas
- **State Management:** Event handling and transitions
- **Database Operations:** CRUD operations with error scenarios
- **UI Components:** Alignment, formatting, and interaction
- **SQL Highlighting:** Keyword recognition and syntax coloring
- **Color System:** Theme application and consistency

## Common Development Tasks

### Adding New Database Support
1. Implement driver detection logic in `db/db.go:DetermineDriver`
2. Add connection string parsing
3. Handle database-specific SQL variations
4. Add comprehensive tests

### Adding New UI Components
1. Create component in `view/` package
2. Follow existing patterns (border, padding, color usage)
3. Use centralized colors from `view/color.go`
4. Add comprehensive alignment tests
5. Integrate into main view coordinator

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
- **State Transitions:** Complex interaction between modes
- **UI Alignment:** Precise formatting requirements for terminal display
- **Database Compatibility:** Multi-database support considerations
- **Color Theming:** Centralized system affects entire application
- **Testing:** Comprehensive coverage expectations

### Common Pitfalls to Avoid
- **Direct Color Usage:** Always use `Colors` global instance
- **State Mutation:** Follow proper state transition patterns
- **Test Isolation:** Ensure tests don't affect each other
- **UI Spacing:** Maintain consistent alignment and padding
- **Error Handling:** Provide graceful degradation paths

---
*Generated on 2025-08-12 by Claude Code for comprehensive repository understanding and AI-assisted development.*