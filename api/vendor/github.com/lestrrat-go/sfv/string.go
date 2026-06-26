package sfv

import (
	"strconv"
)

// StringItem represents a quoted string value,
// with optional parameters.
//
// StringItem implements the Item interface.
type StringItem = FullItem[*StringBareItem, string]

var _ Item = (*StringItem)(nil)

// StringBareItem is a bare item that represents a quoted string.
// Bare items cannot have parameters. Some constructs
// may require a bare item instead of a full string item
// (e.g. dictionary values).
type StringBareItem struct {
	uvalue[string]
}

var _ BareItem = (*StringBareItem)(nil)

// String creates a new String (StringItem) with the
// given string. This function does NOT validate the string
// to ensure it is a valid quoted string (Validation only happens
// when the item is marshaled/parsed).
//
// If you need a bare string item, use BareString() instead.
func String(s string) *StringItem {
	return BareString(s).toItem()
}

func (s *StringBareItem) toItem() *StringItem {
	return &StringItem{
		bare:   s,
		params: NewParameters(),
	}
}

// BareString creates a new StringBareItem with the given string.
// This function does NOT validate the string to ensure it is a
// valid quoted string (Validation only happens when the item is
// marshaled/parsed).
//
// If you need a full string item (with parameters), use String() instead.
func BareString(s string) *StringBareItem {
	var v StringBareItem
	_ = v.SetValue(s)
	return &v
}

// ToItem converts the StringBareItem to a full Item.
func (s *StringBareItem) ToItem() Item {
	return s.toItem()
}

// MarshalSFV implements the Marshaler interface for StringBareItem.
func (s StringBareItem) MarshalSFV() ([]byte, error) {
	quoted := strconv.Quote(s.value)
	return []byte(quoted), nil
}

// Type returns the type of the StringBareItem, useful when
// you have a list of BareItems and need to know the type
// of each item.
func (s StringBareItem) Type() int {
	return StringType
}
