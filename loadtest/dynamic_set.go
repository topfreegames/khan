package loadtest

import "fmt"

// DynamicStringSet represents a dynamic (add/remove) set of strings where strings are mapped to integers in [0, len(maps))
type DynamicStringSet struct {
	stringToInt map[string]int
	intToString []string
}

// DynamicStringSetOutOfBoundsError represents the error when trying to read unexisting position in the set
type DynamicStringSetOutOfBoundsError struct {
	Size  int
	Index int
}

func (e *DynamicStringSetOutOfBoundsError) Error() string {
	return fmt.Sprintf("Trying to access invalid position '%v' in dynamic string set with size '%v'.", e.Index, e.Size)
}

// NewDynamicStringSet returns a new DynamicStringSet
func NewDynamicStringSet() *DynamicStringSet {
	return &DynamicStringSet{
		stringToInt: make(map[string]int),
	}
}

// AddString adds a string to the set
func (d *DynamicStringSet) AddString(str string) {
	if _, ok := d.stringToInt[str]; !ok {
		idx := d.Len()
		d.stringToInt[str] = idx
		d.intToString = append(d.intToString, str)
	}
}

// RemoveString removes a string from the set
func (d *DynamicStringSet) RemoveString(str string) {
	if idx, ok := d.stringToInt[str]; ok {
		sz := d.Len()
		delete(d.stringToInt, str)
		d.intToString[idx] = d.intToString[sz-1]
		d.intToString = d.intToString[:sz-1]
	}
}

// Len returns the number of elements in the set
func (d *DynamicStringSet) Len() int {
	return len(d.stringToInt)
}

// Get returns the string in the specified integer index
func (d *DynamicStringSet) Get(idx int) (string, error) {
	sz := d.Len()
	if 0 <= idx && idx < sz {
		return d.intToString[idx], nil
	}
	return "", &DynamicStringSetOutOfBoundsError{
		Size:  sz,
		Index: idx,
	}
}

// Has returns a boolean telling whether a string is in the set or not
func (d *DynamicStringSet) Has(str string) bool {
	_, ok := d.stringToInt[str]
	return ok
}
