package command

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/mirkobrombin/go-cli-builder/v1/log"
)

// AddFlag adds a string flag to the command.
//
// Parameters:
//   - name: The name of the flag.
//   - usage: The usage description of the flag.
//   - allowArg: Whether the flag allows an argument.
//
// Returns:
//   - A pointer to the string value of the flag.
func (c *Command) AddFlag(name, shortName, usage, defaultValue string, allowArg, required bool) *string {
	if c.Flags == nil {
		c.Flags = flag.NewFlagSet(c.Name, flag.ExitOnError)
		c.ArgFlags = make(map[string]bool)
	}
	if c.ShortFlagMap == nil {
		c.ShortFlagMap = make(map[string]string)
	}
	c.ArgFlags[name] = allowArg
	if shortName != "" {
		c.ShortFlagMap[shortName] = name
	}
	if required {
		c.RequiredFlags = append(c.RequiredFlags, name)
	}
	return c.Flags.String(name, defaultValue, usage)
}

// AddIntFlag adds an integer flag to the command.
//
// Parameters:
//   - name: The name of the flag.
//   - usage: The usage description of the flag.
//   - allowArg: Whether the flag allows an argument.
//
// Returns:
//   - A pointer to the integer value of the flag.
func (c *Command) AddIntFlag(name, shortName, usage string, defaultValue int, allowArg, required bool) *int {
	if c.Flags == nil {
		c.Flags = flag.NewFlagSet(c.Name, flag.ExitOnError)
		c.ArgFlags = make(map[string]bool)
	}
	if c.ShortFlagMap == nil {
		c.ShortFlagMap = make(map[string]string)
	}
	c.ArgFlags[name] = allowArg
	if shortName != "" {
		c.ShortFlagMap[shortName] = name
	}
	if required {
		c.RequiredFlags = append(c.RequiredFlags, name)
	}
	return c.Flags.Int(name, defaultValue, usage)
}

// AddBoolFlag adds a boolean flag to the command.
//
// Parameters:
//   - name: The name of the flag.
//   - usage: The usage description of the flag.
//   - allowArg: Whether the flag allows an argument.
//
// Returns:
//   - A pointer to the boolean value of the flag.
func (c *Command) AddBoolFlag(name, shortName, usage string, defaultValue, allowArg, required bool) *bool {
	if c.Flags == nil {
		c.Flags = flag.NewFlagSet(c.Name, flag.ExitOnError)
		c.ArgFlags = make(map[string]bool)
	}
	if c.ShortFlagMap == nil {
		c.ShortFlagMap = make(map[string]string)
	}
	c.ArgFlags[name] = allowArg
	if shortName != "" {
		c.ShortFlagMap[shortName] = name
	}
	if required {
		c.RequiredFlags = append(c.RequiredFlags, name)
	}

	return c.Flags.Bool(name, defaultValue, usage)
}

// GetFlagString retrieves the string value of a flag.
//
// Parameters:
//   - name: The name of the flag.
//
// Returns:
//   - The string value of the flag, or an empty string if the flag does not exist.
func (c *Command) GetFlagString(name string) string {
	f := c.Flags.Lookup(name)
	if f == nil {
		return ""
	}
	return f.Value.String()
}

// GetFlagInt retrieves the integer value of a flag.
//
// Parameters:
//   - name: The name of the flag.
//
// Returns:
//   - The integer value of the flag, or -1 if the flag does not exist or is not an integer.
func (c *Command) GetFlagInt(name string) int {
	f := c.Flags.Lookup(name)
	if f == nil {
		return -1
	}
	val, err := strconv.Atoi(f.Value.String())
	if err != nil {
		return -1
	}
	return val
}

// GetFlagBool retrieves the boolean value of a flag.
//
// Parameters:
//   - name: The name of the flag.
//
// Returns:
//   - The boolean value of the flag, or false if the flag does not exist.
func (c *Command) GetFlagBool(name string) bool {
	f := c.Flags.Lookup(name)
	if f == nil {
		return false
	}
	val, err := strconv.ParseBool(f.Value.String())
	if err != nil {
		return false
	}
	return val
}

// GetEnv retrieves the value of an environment variable.
//
// Parameters:
//   - key: The name of the environment variable.
//   - defaultValue: The default value to return if the variable is not set.
//
// Returns:
//   - The value of the environment variable, or the default value if not set.
func (c *Command) GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetEnvInt retrieves the integer value of an environment variable.
//
// Parameters:
//   - key: The name of the environment variable.
//   - defaultValue: The default value to return if the variable is not set or not an integer.
//
// Returns:
//   - The integer value of the environment variable, or the default value if not set or not an integer.
func (c *Command) GetEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// GetEnvBool retrieves the boolean value of an environment variable.
//
// Parameters:
//   - key: The name of the environment variable.
//   - defaultValue: The default value to return if the variable is not set or not a boolean.
//
// Returns:
//   - The boolean value of the environment variable, or the default value if not set or not a boolean.
func (c *Command) GetEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

// PrintCommandHelp prints the help message for a single command.
func (c *Command) PrintCommandHelp() {
	fmt.Printf("%s: %s\n\nUsage: %s %s [flags]\n\n", c.Name, c.Description, os.Args[0], c.Name)
	if c.Flags != nil {
		fmt.Println("Flags:")
		c.Flags.PrintDefaults()
	}
}

// setupLogger sets up the logger functions for the command.
//
// Parameters:
//   - name: The name of the command.
func (c *Command) SetupLogger(name string) {
	c.Logger = &log.DefaultLogger{ComponentName: name} //assegnazione logger
}
