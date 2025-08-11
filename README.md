# rel8

A terminal-based database browser and management tool.

## Usage

### Basic Usage
```shell
./rel8
```

### Command Line Options

- `-m, -mock`: Use mock data instead of real database connection (useful for testing and development)
- `-v`: Verbosity level (use `-v`, `-vv`, or `-vvv` for increasing levels of debug output)
- `-demo`: Run demo mode with specified script or file (e.g., `-demo="s(1000),a,b,Enter"` or `-demo="demo.txt"`)
- `demo`: Legacy positional argument for demo mode (still supported for backward compatibility)

### Examples

#### Using with a real database connection
```shell
export DB_DATABASE_CONNECTION_STRING="admin:password@tcp(localhost:3306)/ospatch?parseTime=true"
./rel8
```

#### Using mock data for development/testing
```shell
./rel8 -m
```
or
```shell
./rel8 --mock
```

#### Running demo mode with the demo flag (recommended)
```shell
# Inline script
./rel8 -m -demo="s(1000),:,t,a,b,l,e,Enter,Down,d,s(500),:,q,Enter"

# From file
./rel8 -m -demo="demo.txt"
```

#### Running demo mode with custom commands (legacy syntax)
```shell
./rel8 -m demo "s(1000),:,t,a,b,l,e,Enter,Down,d,s(500),:,q,Enter"
```

#### Quick demo examples
```shell
# Simple table browsing demo
./rel8 -m -demo="s(2000),:,t,a,b,l,e,Enter,s(1000),Down,Down,s(1000),:,q,Enter"

# Fast demo with minimal delays  
./rel8 -m -demo="s(1000),:,t,a,b,l,e,Enter,s(500),Down,d,s(500),:,q,Enter"
```

##### Demo Command Format
Demo commands are comma-separated sequences of:
- **Single letters**: Any letter (e.g., `a`, `b`, `q`) sends that key
- **Special keys**: `Up`, `Down`, `Left`, `Right`, `Enter`, `Tab`, `Escape`, `Backspace`, `Delete`
- **Sleep commands**: `sleep(milliseconds)` or `s(milliseconds)` pauses execution (e.g., `sleep(1000)` or `s(1000)` for 1 second)

Examples:
- `"sleep(2000),:,t,a,b,l,e,Enter,sleep(1000),Down,d"` 
- `"s(2000),:,t,a,b,l,e,Enter,s(1000),Down,d"` (shorter syntax)

Both will:
1. Wait 2 seconds
2. Send `:table` + Enter to show tables
3. Wait 1 second  
4. Send Down arrow
5. Send `d` key

### Demo Script Files

You can store demo scripts in text files for easier management and reuse:

#### Creating Script Files
```bash
# Create a demo script file
echo "s(2000),:,t,a,b,l,e,Enter,s(1000),Down,s(1000),d,s(1000),:,q,Enter" > table-demo.txt

# Run the demo from file
./rel8 -m -demo="table-demo.txt"
```

#### File Format
- **Multiple lines supported**: Commands can be on separate lines for readability
- **Comments**: Everything after `#` on a line is ignored as a comment
- **Whitespace**: Leading/trailing whitespace is automatically trimmed
- **Empty lines**: Automatically ignored
- **Encoding**: Plain text (UTF-8)

#### Example Script Files

**Simple single-line format:**
```bash
# quick-demo.txt
s(1000),:,t,a,b,l,e,Enter,s(500),Down,d,s(500),:,q,Enter
```

**Multi-line format with comments:**
```bash
# commented-demo.txt
# Demo script for table browsing
# Wait for app to start
s(2000)
# Enter command mode and type "table"
:,s(100),t,s(100),a,s(100),b,s(100),l,s(100),e,s(100)
# Execute the command
Enter,s(1000)
# Navigate through the table
Down,s(500),Down,s(500) # Move down twice
# Show details
d,s(2000) # Press 'd' to view details
# Exit the application  
:,q,s(500),Enter,s(1000) # Type :q and exit
```

**Complex demo with inline comments:**
```bash
# detailed-demo.txt  
s(3000) # Wait for startup
:,s(200),t,s(200),a,s(200),b,s(200),l,s(200),e,s(200) # Type "table"
Enter,s(2000) # Execute command
Down,s(1000),Down,s(1000) # Navigate down
d,s(3000) # Show details
:,q,s(1000),Enter,s(1000) # Exit
```

### Scripting Rules and Best Practices

#### Command Structure
- Commands are **comma-separated** - each command must be separated by a comma
- **No spaces around commands** - use `s(1000),a,b` not `s(1000), a, b`
- Commands are processed **sequentially** in the order specified
- **Case-sensitive** - `Enter` works, `enter` does not

#### Sleep Commands
- **Purpose**: Control timing and pacing of demo execution
- **Formats**: `sleep(ms)` or `s(ms)` where `ms` is milliseconds
- **Examples**: `s(1000)` = 1 second, `s(500)` = 0.5 seconds
- **Best practices**:
  - Use `s(2000-3000)` at start to allow app initialization
  - Add `s(100-200)` between rapid keystrokes for visibility
  - Use `s(1000)` after commands that trigger screen changes
  - End with `s(1000)` before quit to show final state

#### Special Key Names
- **Navigation**: `Up`, `Down`, `Left`, `Right` (arrow keys)
- **Actions**: `Enter`, `Tab`, `Escape`, `Backspace`, `Delete`
- **Exact case required** - `Enter` not `enter`, `Up` not `up`

#### Character Input
- **Single characters**: Any letter, number, or symbol (e.g., `a`, `1`, `:`, `+`)
- **Multi-character strings**: Not supported - use individual characters
- **Special characters**: Most work directly (`:`, `/`, `?`, etc.)

#### Command Sequences
- **Start app**: Begin with `s(2000)` to allow initialization
- **Enter command mode**: Use `:` to enter command input
- **Execute commands**: End command input with `Enter`
- **Navigate UI**: Use arrow keys (`Up`, `Down`, etc.)
- **Exit app**: Use `:,q,Enter` or just `q` if already in command mode

#### Example Patterns

**Basic table browsing:**
```
s(2000),:,t,a,b,l,e,Enter,s(1000),Down,Down,s(1000),:,q,Enter
```

**Quick demo with minimal delays:**
```
s(1000),:,t,a,b,l,e,Enter,s(500),Down,s(500),d,s(500),:,q,Enter
```

**Detailed demo with explanatory pauses:**
```
s(3000),:,s(200),t,s(200),a,s(200),b,s(200),l,s(200),e,s(200),Enter,s(2000),Down,s(1000),Down,s(1000),d,s(3000),:,q,s(1000),Enter,s(1000)
```

#### Common Patterns
- **App startup**: `s(2000)` or `s(3000)`
- **Command entry**: `:,s(100),c,o,m,m,a,n,d,Enter`
- **Navigation**: `Down,s(500),Down,s(500)` 
- **App exit**: `:,q,Enter,s(1000)`
- **Quick exit**: `q,Enter` (if already in command mode)

#### Troubleshooting
- **Commands not working**: Check for extra spaces or incorrect case
- **Too fast**: Add more `s()` delays between actions
- **App doesn't start**: Increase initial sleep time (`s(3000)`)
- **Commands ignored**: Ensure proper comma separation
- **Unexpected behavior**: Test individual command segments

#### With verbose logging
```shell
./rel8 -vv
```

## Database Connection

The application uses the `DB_DATABASE_CONNECTION_STRING` environment variable to connect to your database. Supported formats:

- MySQL: `mysql://user:password@host:port/database` or `user:password@tcp(host:port)/database`
- PostgreSQL: `postgres://user:password@host:port/database`
- SQLite: `file:path/to/database.db` or `:memory:`

## Building from Source

```shell
go build
```
