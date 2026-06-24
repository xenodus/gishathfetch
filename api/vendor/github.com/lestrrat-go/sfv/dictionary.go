package sfv

import (
	"bytes"
	"fmt"

	"github.com/lestrrat-go/blackmagic"
)

// Dictionary represents an ordered map of string keys to values in the SFV format.
// Values can be Items, BareItems, or InnerLists. Dictionary maintains insertion
// order and serializes as semicolon-separated key=value pairs according to RFC 9651.
type Dictionary struct {
	keys   []string
	values map[string]any
}

// NewDictionary creates a new empty Dictionary. A Dictionary represents
// a Structured Field Dictionary, which is an ordered map of string keys
// to values (Items, BareItems, or InnerLists).
func NewDictionary() *Dictionary {
	return &Dictionary{
		keys:   make([]string, 0),
		values: make(map[string]any),
	}
}

// Set adds or updates a key-value pair in the dictionary.
// The value must be an Item, BareItem, or *InnerList.
// Returns an error if the value type is not supported.
func (d *Dictionary) Set(key string, value any) error {
	switch value.(type) {
	case Item, BareItem, *InnerList:
		// ok. no op
	default:
		return fmt.Errorf("value must be of type Item, BareItem, or *InnerList, got %T", value)
	}

	if _, exists := d.values[key]; !exists {
		d.keys = append(d.keys, key)
	}
	d.values[key] = value
	return nil
}

// GetValue retrieves the value associated with the given key and assigns
// it to dst. Returns an error if the key is not found or if assignment fails.
func (d *Dictionary) GetValue(key string, dst any) error {
	value, exists := d.values[key]
	if !exists {
		return fmt.Errorf("key %q not found in dictionary", key)
	}
	return blackmagic.AssignIfCompatible(dst, value)
}

// MarshalSFV implements the Marshaler interface for Dictionary
func (d *Dictionary) MarshalSFV() ([]byte, error) {
	if d == nil || len(d.keys) == 0 {
		return []byte{}, nil
	}

	var buf bytes.Buffer
	first := true

	for _, key := range d.keys {
		var value any
		if err := d.GetValue(key, &value); err != nil {
			continue
		}

		// Add separator between dictionary entries
		if !first {
			buf.WriteString(", ")
		}
		first = false

		// Write the key
		buf.WriteString(key)

		// Check if this is a Boolean true value (bare key)
		isBareKey := false
		switch v := value.(type) {
		case Item:
			if v.Type() == BooleanType {
				var b bool
				if err := v.GetValue(&b); err == nil && b {
					isBareKey = true
				}
			}
		case BareItem:
			if v.Type() == BooleanType {
				var b bool
				if err := v.GetValue(&b); err == nil && b {
					isBareKey = true
				}
			}
		}

		// For bare keys (Boolean true), we still need to marshal to get parameters
		if isBareKey {
			// For Boolean true, don't include the =?1 part, just parameters
			if item, ok := value.(Item); ok && item.Parameters() != nil && item.Parameters().Len() > 0 {
				paramBytes, err := item.Parameters().MarshalSFV()
				if err != nil {
					return nil, fmt.Errorf("error marshaling parameters for dictionary key %q: %w", key, err)
				}
				buf.Write(paramBytes)
			}
			// BareItems don't have parameters, so no need to handle that case
		} else {
			// Regular values - include equals and full marshaling
			buf.WriteByte('=')
			var valueBytes []byte
			var err error

			switch v := value.(type) {
			case Item:
				valueBytes, err = v.MarshalSFV()
			case BareItem:
				// Convert BareItem to Item for marshaling
				item := v.ToItem()
				valueBytes, err = item.MarshalSFV()
			case *InnerList:
				valueBytes, err = v.MarshalSFV()
			default:
				return nil, fmt.Errorf("unsupported dictionary value type: %T", v)
			}

			if err != nil {
				return nil, fmt.Errorf("error marshaling dictionary value for key %q: %w", key, err)
			}

			buf.Write(valueBytes)
		}
	}

	return buf.Bytes(), nil
}

// Keys returns the ordered list of keys in the dictionary
func (d *Dictionary) Keys() []string {
	if d == nil {
		return nil
	}
	return d.keys
}
