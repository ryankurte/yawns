package plugins

import (
	"fmt"
)

// GetOption fetches an option by name from the provided map or returns the default instance
func GetOption(name string, defaultOption interface{}, options map[string]interface{}) interface{} {
	if a, ok := options[name]; ok {
		return a
	}
	return defaultOption
}

// GetOptionString fetches an option string from the provided map or returns the provided default string
func GetOptionString(name string, defaultOption string, options map[string]interface{}) (string, error) {
	obj := GetOption(name, defaultOption, options)
	if str, ok := obj.(string); ok {
		return str, nil
	}
	if str, ok := obj.(*string); ok {
		return *str, nil
	}
	return "", fmt.Errorf("Error: option %s must be of type 'string' not `%t`", name, obj)
}

// GetOptionUint fetches an option uint from the provided map or returns the provided default uint
func GetOptionUint(name string, defaultOption uint32, options map[string]interface{}) (uint32, error) {
	obj := GetOption(name, defaultOption, options)
	if u, ok := obj.(uint); ok {
		return uint32(u), nil
	}
	if u, ok := obj.(int); ok {
		return uint32(u), nil
	}
	if u, ok := obj.(uint32); ok {
		return uint32(u), nil
	}
	return 0, fmt.Errorf("Error: option %s must be of type 'int' not `%t`", name, obj)
}
