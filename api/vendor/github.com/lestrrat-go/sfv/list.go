package sfv

import (
	"bytes"
	"fmt"
)

// InnerList represents a grouped sequence of Items with optional parameters
// in the SFV format. InnerLists are used within Lists and Dictionaries to
// group related items together as a single value.
type InnerList struct {
	values []Item
	params *Parameters
}

// NewInnerList creates a new empty InnerList with properly initialized parameters.
// An InnerList represents a grouped sequence of Items with optional parameters
// in the SFV format.
func NewInnerList() *InnerList {
	return &InnerList{
		values: make([]Item, 0),
		params: NewParameters(),
	}
}

// Add adds an item to the inner list. The item must be an Item or BareItem.
// BareItems are automatically converted to Items. Returns an error if the
// item type is not supported.
func (il *InnerList) Add(in any) error {
	var item Item
	switch v := in.(type) {
	case Item:
		item = v
	case BareItem:
		item = v.ToItem()
	default:
		return fmt.Errorf("item must be of type Item or BareItem, got %T", item)
	}
	il.values = append(il.values, item)
	return nil
}

// Len returns the number of values in the inner list
func (il *InnerList) Len() int {
	if il == nil {
		return 0
	}
	return len(il.values)
}

// Get returns the value at the specified index
func (il *InnerList) Get(index int) (Item, bool) {
	if il == nil || index < 0 || index >= len(il.values) {
		return nil, false
	}
	return il.values[index], true
}

// MarshalSFV implements the Marshaler interface for InnerList
func (il *InnerList) MarshalSFV() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('(')

	for i := range il.Len() {
		if i > 0 {
			buf.WriteByte(' ')
		}

		item, ok := il.Get(i)
		if !ok {
			continue
		}

		itemBytes, err := item.MarshalSFV()
		if err != nil {
			return nil, err
		}

		buf.Write(itemBytes)
	}

	buf.WriteByte(')')

	// Add parameters if any
	if il.params != nil && il.params.Len() > 0 {
		paramBytes, err := il.params.MarshalSFV()
		if err != nil {
			return nil, err
		}
		buf.Write(paramBytes)
	}

	return buf.Bytes(), nil
}

// Parameters returns the parameters associated with this InnerList
func (il *InnerList) Parameters() *Parameters {
	if il == nil {
		return nil
	}
	return il.params
}

// List represents an ordered sequence of Items and InnerLists in the SFV format.
// Lists can contain Items (with optional parameters) and InnerLists as comma-separated
// values according to RFC 9651.
type List struct {
	values []any
}

// Add adds an item to the list. The item must be an Item, BareItem, or *InnerList.
// BareItems are automatically converted to Items. Returns an error if the
// item type is not supported.
func (l *List) Add(in any) error {
	// Process the input to ensure it's a proper SFV item
	switch v := in.(type) {
	case Item:
		l.values = append(l.values, v)
	case BareItem:
		l.values = append(l.values, v.ToItem())
	case *InnerList:
		l.values = append(l.values, v)
	default:
		return fmt.Errorf("list item must be of type Item, BareItem, or *InnerList, got %T", in)
	}
	return nil
}

// MarshalSFV implements the Marshaler interface for List
func (l List) MarshalSFV() ([]byte, error) {
	if l.Len() == 0 {
		return nil, nil
	}

	var buf bytes.Buffer
	for i := range l.Len() {
		value, ok := l.Get(i)
		if !ok {
			return nil, fmt.Errorf("index %d out of range for list of length %d", i, l.Len())
		}

		if i > 0 {
			buf.WriteString(", ")
		}

		vfsv, err := valueToSFV(value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to SFV: %w", err)
		}

		item, err := vfsv.MarshalSFV()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal value to SFV: %w", err)
		}

		buf.Write(item)
	}

	return buf.Bytes(), nil
}

// Len returns the number of values in the list
func (l *List) Len() int {
	if l == nil {
		return 0
	}
	return len(l.values)
}

// Get returns the value at the specified index
func (l *List) Get(index int) (any, bool) {
	if l == nil || index < 0 || index >= len(l.values) {
		return nil, false
	}
	return l.values[index], true
}
