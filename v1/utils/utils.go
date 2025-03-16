package utils

import (
	"os"
	"strconv"
	"strings"
	"syscall"
)

// Colorize adds colors to the output text.
//
// Parameters:
//   - text: The text to colorize.
//   - color: The color to apply (red, green, yellow, blue, reset).
//
// Returns:
//   - The colorized text.
func Colorize(text string, color string) string {
	colors := map[string]string{
		"red":    "\033[31m",
		"green":  "\033[32m",
		"yellow": "\033[33m",
		"blue":   "\033[34m",
		"reset":  "\033[0m",
	}
	return colors[color] + text + colors["reset"]
}

// SplitIgnore splits the ignore string into a slice of strings.
//
// Parameters:
//   - ignore: The comma-separated ignore string.
//
// Returns:
//   - A slice of strings representing the split ignore list.
func SplitIgnore(ignore string) []string {
	return strings.Split(ignore, ",")
}

// StringToInt converts a string to an integer.
//
// Parameters:
//   - s: The string to convert.
//
// Returns:
//   - The integer value and an error, if any.
func StringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// reexec restarts the process with the new arguments.
//
// Parameters:
//   - args: The new arguments for the process.
//
// Returns:
//   - An error, if any.
func Reexec(args []string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return syscall.Exec(exe, append([]string{exe}, args...), os.Environ())
}
