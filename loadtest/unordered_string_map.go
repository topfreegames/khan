package loadtest

import "fmt"

type unorderedStringMapValue struct {
	index   int
	content interface{}
}

// UnorderedStringMap represents a map from strings to interface{}s that can be looped using zero-based indexes with order based on previously executed set/remove operations
type UnorderedStringMap struct {
	stringToInt map[string]unorderedStringMapValue
	intToString []string
}

// UnorderedStringMapOutOfBoundsError represents the error when trying to read unexisting position in the key space
type UnorderedStringMapOutOfBoundsError struct {
	Size  int
	Index int
}

func (e *UnorderedStringMapOutOfBoundsError) Error() string {
	return fmt.Sprintf("Trying to access invalid position '%v' in UnorderedStringMap with size '%v'.", e.Index, e.Size)
}

// NewUnorderedStringMap returns a new UnorderedStringMap
func NewUnorderedStringMap() *UnorderedStringMap {
	return &UnorderedStringMap{
		stringToInt: make(map[string]unorderedStringMapValue),
	}
}

// Set maps a string to a value
func (d *UnorderedStringMap) Set(key string, value interface{}) {
	if _, ok := d.stringToInt[key]; !ok {
		idx := d.Len()
		d.stringToInt[key] = unorderedStringMapValue{idx, value}
		d.intToString = append(d.intToString, key)
	}
}

// Remove removes a string key
func (d *UnorderedStringMap) Remove(key string) {
	if value, ok := d.stringToInt[key]; ok {
		sz := d.Len()
		delete(d.stringToInt, key)
		d.intToString[value.index] = d.intToString[sz-1]
		d.intToString = d.intToString[:sz-1]
	}
}

// Len returns the number of elements
func (d *UnorderedStringMap) Len() int {
	return len(d.stringToInt)
}

// GetKey returns the string key at the specified integer index
func (d *UnorderedStringMap) GetKey(idx int) (string, error) {
	sz := d.Len()
	if 0 <= idx && idx < sz {
		return d.intToString[idx], nil
	}
	return "", &UnorderedStringMapOutOfBoundsError{
		Size:  sz,
		Index: idx,
	}
}

// Has returns a boolean telling whether a string is in the key space or not
func (d *UnorderedStringMap) Has(str string) bool {
	_, ok := d.stringToInt[str]
	return ok
}
