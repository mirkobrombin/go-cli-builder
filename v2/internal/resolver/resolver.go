package resolver

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
)

// BindValue binds a string value to a reflect.Value.
//
// Example:
//
//	var x int
//	err := resolver.BindValue(reflect.ValueOf(&x).Elem(), "42")
func BindValue(val reflect.Value, value string) error {
	switch val.Kind() {
	case reflect.String:
		val.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration: %v", value)
			}
			val.SetInt(int64(d))
		} else {
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer: %v", value)
			}
			val.SetInt(i)
		}
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float: %v", value)
		}
		val.SetFloat(f)
	case reflect.Bool:
		b, err := ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean: %v", value)
		}
		val.SetBool(b)
	case reflect.Slice:
		if val.Type().Elem().Kind() == reflect.String {
			val.Set(reflect.Append(val, reflect.ValueOf(value)))
		} else {
			return fmt.Errorf("unsupported slice type: %v", val.Type().Elem().Kind())
		}
	default:
		return fmt.Errorf("unsupported type: %v", val.Kind())
	}
	return nil
}

// ParseBool parses a boolean string with support for more formats.
//
// Example:
//
//	b, err := resolver.ParseBool("yes")
func ParseBool(str string) (bool, error) {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True", "yes", "Yes", "YES", "on", "ON":
		return true, nil
	case "0", "f", "F", "false", "FALSE", "False", "no", "No", "NO", "off", "OFF":
		return false, nil
	}
	return false, fmt.Errorf("invalid boolean value: %s", str)
}

// GetValue returns the value to bind based on priority: CLI > Env > Default.
//
// Example:
//
//	val, err := resolver.GetValue(nil, "ENV_VAR", "default", false)
func GetValue(cliVal *string, envName, defVal string, isFlag bool) (string, error) {
	if cliVal != nil {
		return *cliVal, nil
	}

	if envName != "" {
		if val, ok := os.LookupEnv(envName); ok {
			return val, nil
		}
	}

	if defVal != "" {
		return defVal, nil
	}

	return "", nil
}
