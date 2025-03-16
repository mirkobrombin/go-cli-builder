package command

import "strconv"

// GetString retrieves the string value of a root flag.
func (rf *RootFlags) GetString(name string) string {
	if rf == nil {
		return ""
	}
	f := rf.Lookup(name)
	if f == nil {
		return ""
	}
	return f.Value.String()
}

// GetInt retrieves the integer value of a root flag.
func (rf *RootFlags) GetInt(name string) int {
	if rf == nil {
		return -1
	}
	f := rf.Lookup(name)
	if f == nil {
		return -1
	}
	val, err := strconv.Atoi(f.Value.String())
	if err != nil {
		return -1
	}
	return val
}

// GetBool retrieves the boolean value of a root flag.
func (rf *RootFlags) GetBool(name string) bool {
	if rf == nil {
		return false
	}
	f := rf.Lookup(name)
	if f == nil {
		return false
	}
	val, err := strconv.ParseBool(f.Value.String())
	if err != nil {
		return false
	}
	return val
}
