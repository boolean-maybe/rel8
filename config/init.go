package config

import (
	"flag"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"log"
	"log/slog"
	"net/url"
	"os"
)

func Configure() (string, bool, string) {
	var verbosity int
	var useMock bool
	var demoScript string

	// Count the number of -v flags from os.Args to support -v, -vv, -vvv syntax
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-v":
			verbosity = 1
		case "-vv":
			verbosity = 2
		case "-vvv":
			verbosity = 3
		}
	}

	// Define the flags but don't parse verbosity since we handle it manually above
	flag.IntVar(&verbosity, "v", verbosity, "verbosity level: 0=Error, 1=Warn, 2=Info, 3=Debug (use -v, -vv, -vvv)")
	flag.BoolVar(&useMock, "m", false, "use mock data instead of real database connection")
	flag.BoolVar(&useMock, "mock", false, "use mock data instead of real database connection")
	flag.StringVar(&demoScript, "demo", "", "run demo mode with specified script or file (e.g., 's(1000),a,b,Enter' or 'demo.txt')")

	// Filter out the verbosity flags before parsing
	var filteredArgs []string
	for _, arg := range os.Args[1:] {
		if arg != "-v" && arg != "-vv" && arg != "-vvv" {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	// Temporarily replace os.Args for flag parsing
	originalArgs := os.Args
	os.Args = append([]string{os.Args[0]}, filteredArgs...)
	flag.Parse()
	os.Args = originalArgs

	// Set logging level based on verbosity
	var logLevel slog.Level
	switch verbosity {
	case 0:
		logLevel = slog.LevelError
	case 1:
		logLevel = slog.LevelWarn
	case 2:
		logLevel = slog.LevelInfo
	case 3:
		logLevel = slog.LevelDebug
	default:
		logLevel = slog.LevelError
	}

	// Setup structured logging to rel8.log file
	logFile, err := os.OpenFile("rel8.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewTextHandler(logFile, opts))
	slog.SetDefault(logger)

	slog.Info("Application starting", "version", "1.0")

	viper.SetDefault("database.connection_string", "postgres://user:password@localhost:5432/mydb?sslmode=disable")
	viper.SetEnvPrefix("DB")
	viper.AutomaticEnv()
	viper.BindEnv("database.connection_string", "DB_DATABASE_CONNECTION_STRING")

	connStr := viper.GetString("database.connection_string")
	slog.Info("Database connection string", "connection_string", connStr)
	if useMock {
		slog.Info("Mock mode enabled - will use mock data instead of real database")
	}
	if demoScript != "" {
		slog.Info("Demo mode enabled", "script", demoScript)
	}
	debugConnectionString(connStr)
	return connStr, useMock, demoScript
}

func debugConnectionString(connStr string) {

	slog.Info("=== Connection String Debug ===")
	slog.Info("Raw connection string", "connection_string", connStr)

	// Try to parse as URL
	u, err := url.Parse(connStr)
	if err != nil {
		slog.Error("Failed to parse as URL", "error", err)
		return
	}

	slog.Info("URL components", "scheme", u.Scheme, "host", u.Host, "path", u.Path)

	if u.User != nil {
		slog.Info("User credentials", "username", u.User.Username())
		if password, set := u.User.Password(); set {
			slog.Info("Password", "password", password)
		} else {
			slog.Info("Password not set")
		}
	} else {
		slog.Info("User info not set")
	}

	if u.RawQuery != "" {
		slog.Info("Query parameters", "raw_query", u.RawQuery)
		params := u.Query()
		for key, values := range params {
			slog.Info("Query parameter", "key", key, "values", values)
		}
	}
	slog.Info("===============================")
}
