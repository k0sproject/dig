// Package dig provides a map[string]any Mapping type that has ruby-like "dig" functionality.
//
// It can be used for example to access and manipulate arbitrary nested YAML/JSON structures.
package dig

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Mapping is a nested key-value map where the keys are strings and values are any. In Ruby it is called a Hash (with string keys), in YAML it's called a "mapping".
type Mapping map[string]any

// UnmarshalText for supporting json.Unmarshal
func (m *Mapping) UnmarshalJSON(text []byte) error {
	var result map[string]any
	if err := json.Unmarshal(text, &result); err != nil {
		return err
	}
	*m = cleanUpInterfaceMap(result)
	return nil
}

// UnmarshalYAML for supporting yaml.Unmarshal
func (m *Mapping) UnmarshalYAML(unmarshal func(any) error) error {
	var result map[string]any
	if err := unmarshal(&result); err != nil {
		return err
	}
	*m = cleanUpInterfaceMap(result)
	return nil
}

// Dig is a simplistic implementation of a Ruby-like Hash.dig functionality.
//
// It returns a value from a (deeply) nested tree structure.
func (m *Mapping) Dig(keys ...string) any {
	if len(keys) == 0 {
		return nil
	}
	v, ok := (*m)[keys[0]]
	if !ok {
		return nil
	}
	switch v := v.(type) {
	case Mapping:
		if len(keys) == 1 {
			return v
		}
		return v.Dig(keys[1:]...)
	default:
		if len(keys) > 1 {
			return nil
		}
		return v
	}
}

// DigString is like Dig but returns the value as string
func (m *Mapping) DigString(keys ...string) string {
	v := m.Dig(keys...)
	val, ok := v.(string)
	if !ok {
		return ""
	}
	return val
}

// DigMapping always returns a mapping, creating missing or overwriting non-mapping branches in between
func (m *Mapping) DigMapping(keys ...string) Mapping {
	k := keys[0]
	cur := (*m)[k]
	switch v := cur.(type) {
	case Mapping:
		if len(keys) > 1 {
			return v.DigMapping(keys[1:]...)
		}
		return v
	default:
		n := Mapping{}
		(*m)[k] = n
		if len(keys) > 1 {
			return n.DigMapping(keys[1:]...)
		}
		return n
	}
}

// Dup creates a dereferenced copy of the Mapping
func (m *Mapping) Dup() Mapping {
	newMap := make(Mapping, len(*m))
	for k, v := range *m {
		newMap[k] = deepCopy(v)
	}
	return newMap
}

// HasKey checks if the key exists in the Mapping.
func (m *Mapping) HasKey(key string) bool {
	_, ok := (*m)[key]
	return ok
}

// HasMapping checks if the key exists in the Mapping and is a Mapping.
func (m *Mapping) HasMapping(key string) bool {
	v, ok := (*m)[key]
	if !ok {
		return false
	}
	_, ok = v.(Mapping)
	return ok
}

// MergeOptions are used to configure the Merge function.
type MergeOptions struct {
	// Overwrite existing values in the target map
	Overwrite bool
	// Nillify removes keys from the target map if the value is nil in the source map
	Nillify bool
}

type MergeOption func(*MergeOptions)

// WithOverwrite sets the Overwrite option to true.
func WithOverwrite() MergeOption {
	return func(o *MergeOptions) {
		o.Overwrite = true
	}
}

// WithNillify sets the Nillify option to true.
func WithNillify() MergeOption {
	return func(o *MergeOptions) {
		o.Nillify = true
	}
}

// Merge deep merges the source map into the target map. Regardless of options, Mappings will be merged recursively. Slices are treated as any single value and are not combined.
func (m Mapping) Merge(source Mapping, opts ...MergeOption) {
	options := MergeOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	for k, v := range source {
		switch v := v.(type) {
		case Mapping:
			if !m.HasKey(k) {
				m[k] = v.Dup()
			} else if m.HasMapping(k) {
				m.DigMapping(k).Merge(v, opts...)
			} else if options.Overwrite {
				m[k] = v.Dup()
			}
		case nil:
			if options.Nillify {
				m[k] = nil
			}
		default:
			if !m.HasKey(k) || options.Overwrite {
				m[k] = deepCopy(v)
			}
		}
	}
}

// deepCopy performs a deep copy of the value using reflection
func deepCopy(value any) any {
	if value == nil {
		return nil
	}

	val := reflect.ValueOf(value)

	switch val.Kind() {
	case reflect.Map:
		newMap := reflect.MakeMap(val.Type())
		for _, key := range val.MapKeys() {
			newMap.SetMapIndex(key, reflect.ValueOf(deepCopy(val.MapIndex(key).Interface())))
		}
		if plainmap, ok := newMap.Interface().(map[string]any); ok {
			return cleanUpInterfaceMap(plainmap)
		}
		return newMap.Interface()
	case reflect.Slice:
		if val.IsNil() {
			return nil
		}
		newSlice := reflect.MakeSlice(val.Type(), val.Len(), val.Cap())
		for i := 0; i < val.Len(); i++ {
			newSlice.Index(i).Set(reflect.ValueOf(deepCopy(val.Index(i).Interface())))
		}
		if plainslice, ok := newSlice.Interface().([]any); ok {
			return cleanUpInterfaceArray(plainslice)
		}
		return newSlice.Interface()

	default:
		return value
	}
}

// Cleans up a slice of interfaces into slice of actual values
func cleanUpInterfaceArray(in []any) []any {
	result := make([]any, len(in))
	for i, v := range in {
		result[i] = cleanUpValue(v)
	}
	return result
}

// Cleans up the map keys to be strings
func cleanUpInterfaceMap(in map[string]any) Mapping {
	result := make(Mapping, len(in))
	for k, v := range in {
		result[k] = cleanUpValue(v)
	}
	return result
}

func stringifyKeys(in map[any]any) map[string]any {
	result := make(map[string]any)
	for k, v := range in {
		result[fmt.Sprintf("%v", k)] = v
	}
	return result
}

// Cleans up the value in the map, recurses in case of arrays and maps
func cleanUpValue(v any) any {
	switch v := v.(type) {
	case []any:
		return cleanUpInterfaceArray(v)
	case map[string]any:
		return cleanUpInterfaceMap(v)
	case map[any]any:
		return cleanUpInterfaceMap(stringifyKeys(v))
	default:
		return v
	}
}
