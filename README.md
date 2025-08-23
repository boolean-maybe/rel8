# rel8

A terminal-based database browser and management tool.

## Usage

to run from IDEA select `Emulate terminal in output console` in Run configuration
and add TERM=xterm-256color env variable

### Basic Usage
```shell
./rel8
```

### Command Line Options

- `-m, -mock`: Use mock data instead of real database connection (useful for testing and development)
- `-v`: Verbosity level (use `-v`, `-vv`, or `-vvv` for increasing levels of debug output)
- `-demo`: Run demo mode with specified script or file (see [demo.md](demo.md) for details)
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

#### Running demo mode
```shell
# Quick example
./rel8 -m -demo="s(1000),:,t,a,b,l,e,Enter,Down,d,s(500),:,q,Enter"

# From file
./rel8 -m -demo="demo.txt"
```

For detailed demo usage, commands, scripting, and examples, see [demo.md](demo.md).

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
