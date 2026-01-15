package binder

import (
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/mirkobrombin/go-foundation/pkg/hooks"
	freflect "github.com/mirkobrombin/go-foundation/pkg/reflect"
)

// HandlerFunc handles binding a set of string arguments to a field.
type HandlerFunc func(args []string) error

// Binder handles mapping and execution of field bindings.
type Binder struct {
	dst      any
	handlers map[string]HandlerFunc
	runner   *hooks.Runner
}

// NewBinder creates a new binder for the destination object.
func NewBinder(dst any) (*Binder, error) {
	b := &Binder{
		dst:      dst,
		handlers: make(map[string]HandlerFunc),
		runner:   hooks.NewRunner(),
	}

	if err := b.autoDiscover(); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Binder) autoDiscover() error {
	v := reflect.ValueOf(b.dst)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("binder: destination must be a pointer to a struct")
	}

	elem := v.Elem()
	typ := elem.Type()

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		fieldType := typ.Field(i)

		tag := fieldType.Tag.Get("flag")
		if tag == "" {
			tag = fieldType.Tag.Get("cli")
		}
		if tag == "" {
			continue
		}

		name := fieldType.Name
		if _, ok := b.handlers[name]; !ok {
			b.registerDefaultHandler(field, name)
		}
	}

	return nil
}

func (b *Binder) registerDefaultHandler(field reflect.Value, name string) {
	b.handlers[name] = func(args []string) error {
		var val string
		if len(args) > 0 {
			val = args[0]
		}

		// Handle boolean flag without value (implicitly true)
		if field.Kind() == reflect.Bool && len(args) == 0 {
			val = "true"
		}

		// Skip binding if empty value for non-string types (unless bool handled above)
		if val == "" && field.Kind() != reflect.String && field.Kind() != reflect.Bool {
			return nil
		}

		return freflect.Bind(field, val)
	}
}

// Handlers returns the registered handlers.
func (b *Binder) Handlers() map[string]HandlerFunc {
	return b.handlers
}

// AddBool registers a custom handler for a boolean field.
func (b *Binder) AddBool(key string, fn func(bool) error) {
	b.handlers[key] = func(args []string) error {
		var val bool
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%t", &val)
		} else {
			val = true
		}
		return fn(val)
	}
}

// AddInt registers a custom handler for an integer field.
func (b *Binder) AddInt(key string, fn func(int64) error) {
	b.handlers[key] = func(args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing value for %s", key)
		}
		var val int64
		if _, err := fmt.Sscanf(args[0], "%d", &val); err != nil {
			return err
		}
		return fn(val)
	}
}

// AddStrings registers a custom handler for a string slice field.
func (b *Binder) AddStrings(key string, fn func([]string) error) {
	b.handlers[key] = fn
}

// AddDuration registers a custom handler for a duration field.
func (b *Binder) AddDuration(key string, fn func(time.Duration) error) {
	b.handlers[key] = func(args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing value for %s", key)
		}
		val, err := time.ParseDuration(args[0])
		if err != nil {
			return err
		}
		return fn(val)
	}
}

// AddEnum registers a handler that validates against a set of allowed values.
func (b *Binder) AddEnum(key string, choices []string, fn func(string) error) {
	b.handlers[key] = func(args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing value for %s", key)
		}
		val := args[0]
		if !slices.Contains(choices, val) {
			return fmt.Errorf("invalid value %s for %s, allowed: %v", val, key, choices)
		}
		return fn(val)
	}
}

// Run executes the handler for the given key with the provided arguments.
func (b *Binder) Run(key string, args []string) error {
	h, ok := b.handlers[key]
	if !ok {
		return fmt.Errorf("no handler for key: %s", key)
	}
	return h(args)
}
