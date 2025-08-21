package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"rel8/view"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

func loadDemoScript(scriptPath string) (string, error) {
	// Check if the script path looks like a file path
	if _, err := os.Stat(scriptPath); err == nil {
		// File exists, read it
		content, err := ioutil.ReadFile(scriptPath)
		if err != nil {
			return "", fmt.Errorf("failed to read demo script file '%s': %v", scriptPath, err)
		}

		// Process the content to remove comments and clean up
		processedContent := processScriptContent(string(content))
		return strings.TrimSpace(processedContent), nil
	}
	// Not a file, return as-is (assume it's an inline script)
	return scriptPath, nil
}

func processScriptContent(content string) string {
	// Split content into lines
	lines := strings.Split(content, "\n")
	var processedLines []string

	for _, line := range lines {
		// Remove comments (everything after #)
		if commentIndex := strings.Index(line, "#"); commentIndex != -1 {
			line = line[:commentIndex]
		}

		// Trim whitespace
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line != "" {
			processedLines = append(processedLines, line)
		}
	}

	// Join all non-empty lines with commas
	return strings.Join(processedLines, ",")
}

func clearTerminal() {
	// Try multiple methods to clear the terminal screen

	// Method 1: Use clear command
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err == nil {
		return
	}

	// Method 2: Use ANSI escape sequences
	fmt.Print("\033[2J\033[H") // Clear screen and move cursor to home

	// Method 3: Reset terminal state
	fmt.Print("\033c") // Reset terminal
}

func runDemo(v *view.ViewManager, commandString string) {
	// Start the view in a goroutine
	done := make(chan bool, 1)

	go func() {
		v.Run()
		done <- true
	}()

	// Parse and execute commands
	commands := strings.Split(commandString, ",")

	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)

		if cmd == "" {
			continue
		}

		// Handle sleep command (both sleep(100) and s(100) formats)
		if (strings.HasPrefix(cmd, "sleep(") && strings.HasSuffix(cmd, ")")) ||
			(strings.HasPrefix(cmd, "s(") && strings.HasSuffix(cmd, ")")) {

			var sleepStr string
			if strings.HasPrefix(cmd, "sleep(") {
				sleepStr = cmd[6 : len(cmd)-1] // Extract number between sleep( and )
			} else {
				sleepStr = cmd[2 : len(cmd)-1] // Extract number between s( and )
			}

			if sleepMs, err := strconv.Atoi(sleepStr); err == nil {
				time.Sleep(time.Duration(sleepMs) * time.Millisecond)
			} else {
				fmt.Printf("Invalid sleep duration: %s\n", sleepStr)
			}
			continue
		}

		// Handle special keys
		switch cmd {
		case "Enter":
			v.App.QueueEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
		case "Up":
			v.App.QueueEvent(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))
		case "Down":
			v.App.QueueEvent(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
		case "Left":
			v.App.QueueEvent(tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone))
		case "Right":
			v.App.QueueEvent(tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone))
		case "Tab":
			v.App.QueueEvent(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
		case "Escape":
			v.App.QueueEvent(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
		case "Backspace":
			v.App.QueueEvent(tcell.NewEventKey(tcell.KeyBackspace, 0, tcell.ModNone))
		case "Delete":
			v.App.QueueEvent(tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone))
		default:
			// Handle single character commands
			if len(cmd) == 1 {
				char := rune(cmd[0])
				v.App.QueueEvent(tcell.NewEventKey(tcell.KeyRune, char, tcell.ModNone))
			} else {
				fmt.Printf("Unknown command: %s\n", cmd)
			}
		}
	}

	// Wait for the app to finish and then clear the terminal
	<-done
	clearTerminal()
}

func demoMode(view *view.ViewManager, demoScript string) {
	if demoScript != "" {
		script, err := loadDemoScript(demoScript)
		if err != nil {
			fmt.Printf("Error loading demo script: %v\n", err)
			os.Exit(1)
		}
		runDemo(view, script)
		return
	}

	// Check for demo=script format (without dash)
	var demoCmdFromArgs string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "demo=") {
			demoCmdFromArgs = strings.TrimPrefix(arg, "demo=")
			break
		}
	}

	if demoCmdFromArgs != "" {
		script, err := loadDemoScript(demoCmdFromArgs)
		if err != nil {
			fmt.Printf("Error loading demo script: %v\n", err)
			os.Exit(1)
		}
		runDemo(view, script)
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "demo" {
		// Support legacy "demo" positional argument for backward compatibility
		defaultCommand := "s(2000),:,s(100),t,s(100),a,s(100),b,s(100),l,s(100),e,s(100),Enter,s(1000),Down,s(1000),d,s(1000),:,q,s(1000),Enter,s(1000)"
		demoCommand := defaultCommand

		// Check if custom demo command is provided as second argument
		if len(os.Args) > 2 {
			demoCommand = os.Args[2]
		}

		script, err := loadDemoScript(demoCommand)
		if err != nil {
			fmt.Printf("Error loading demo script: %v\n", err)
			os.Exit(1)
		}
		runDemo(view, script)
		return
	}
}
