package cli

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/eiannone/keyboard" // Import the keyboard library
)

// clearScreen clears the terminal screen.
// This uses ANSI escape codes, which work on most modern terminals.
func clearScreen() {
	fmt.Print("\033[H\033[J")
}

func SelectFromSlice(prompt string, options []string, defaultValue string) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no options provided to select from")
	}

	// Open keyboard for raw input
	if err := keyboard.Open(); err != nil {
		return "", fmt.Errorf("could not open keyboard: %w", err)
	}
	// Ensure keyboard is closed when the function returns
	defer func() {
		_ = keyboard.Close()
	}()

	currentIndex := 0
	maxIndex := len(options) - 1
	if defaultValue != "" {
		currentIndex = slices.Index(options, defaultValue)
		if currentIndex == -1 {
			return "", fmt.Errorf("default value %s is not in options", defaultValue)
		}
	}

	for {
		clearScreen()
		fmt.Println(prompt)
		fmt.Println(strings.Repeat("-", len(prompt))) // Optional separator

		for i, option := range options {
			if i == currentIndex {
				fmt.Printf("> %s\n", option) // Indicate current selection
			} else {
				fmt.Printf("  %s\n", option)
			}
		}
		fmt.Println("\n(Use ↑/↓ arrows to navigate, Enter to select, Esc or Ctrl+C to quit)")

		// GetKey blocks until a key is pressed
		_, key, err := keyboard.GetKey()
		if err != nil {
			return "", fmt.Errorf("error reading key: %w", err)
		}

		switch key {
		case keyboard.KeyArrowUp:
			currentIndex--
			if currentIndex < 0 {
				currentIndex = maxIndex // Wrap around
			}
		case keyboard.KeyArrowDown:
			currentIndex++
			if currentIndex > maxIndex {
				currentIndex = 0 // Wrap around
			}
		case keyboard.KeyEnter:
			clearScreen() // Clear before printing the final result
			return options[currentIndex], nil
		case keyboard.KeyEsc, keyboard.KeyCtrlC: // Allow Esc or Ctrl+C to abort
			clearScreen()
			return "", fmt.Errorf("selection aborted by user")
		default:
		}
	}
}

func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		t := scanner.Text()
		tr := strings.TrimSpace(t)
		if tr != "" {
			lines = append(lines, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
