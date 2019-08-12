package loadtest

import "fmt"

type unorderedStringMapValue struct {
	index   int
	content interface{}
}

// UnorderedStringMap represents a map from strings to interface{}s that can be looped using zero-based indexes with order based on previously executed Set/Remove operations
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

// Get returns the interface{} content for a key
func (d *UnorderedStringMap) Get(key string) interface{} {
	value, ok := d.stringToInt[key]
	if ok {
		return value.content
	}
	return nil
}

// Set maps a string to a value
func (d *UnorderedStringMap) Set(key string, content interface{}) {
	if value, ok := d.stringToInt[key]; !ok {
		idx := d.Len()
		d.stringToInt[key] = unorderedStringMapValue{idx, content}
		d.intToString = append(d.intToString, key)
	} else {
		d.stringToInt[key] = unorderedStringMapValue{value.index, content}
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
func (d *UnorderedStringMap) Has(key string) bool {
	_, ok := d.stringToInt[key]
	return ok
}
