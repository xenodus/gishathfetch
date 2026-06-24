package sfv

import (
	"bytes"
	"fmt"
)

// DisplayStringItem represents a percent-encoded display string value,
// with optional parameters.
//
// DisplayStringItem implements the Item interface.
type DisplayStringItem = FullItem[*DisplayStringBareItem, string]

var _ Item = (*DisplayStringItem)(nil)

// DisplayStringBareItem is a bare item that represents a percent-encoded display string.
// Bare items cannot have parameters. Some constructs
// may require a bare item instead of a full display string item
// (e.g. dictionary values).
type DisplayStringBareItem struct {
	uvalue[string]
}

var _ BareItem = (*DisplayStringBareItem)(nil)

// DisplayString creates a new DisplayString (DisplayStringItem) with the
// given string. This function does NOT validate the string
// to ensure it is a valid display string (Validation only happens
// when the item is marshaled/parsed).
//
// If you need a bare display string item, use BareDisplayString() instead.
func DisplayString(s string) *DisplayStringItem {
	return BareDisplayString(s).toItem()
}

func (d *DisplayStringBareItem) toItem() *DisplayStringItem {
	return &DisplayStringItem{
		bare:   d,
		params: NewParameters(),
	}
}

// BareDisplayString creates a new DisplayStringBareItem with the given string.
// This function does NOT validate the string to ensure it is a
// valid display string (Validation only happens when the item is
// marshaled/parsed).
//
// If you need a full display string item (with parameters), use DisplayString() instead.
func BareDisplayString(s string) *DisplayStringBareItem {
	var v DisplayStringBareItem
	_ = v.SetValue(s)
	return &v
}

// ToItem converts the DisplayStringBareItem to a full Item.
func (d *DisplayStringBareItem) ToItem() Item {
	return d.toItem()
}

// MarshalSFV implements the Marshaler interface for DisplayStringBareItem.
func (d DisplayStringBareItem) MarshalSFV() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('%')
	buf.WriteByte('"')
	// Percent-encode non-ASCII characters
	for _, r := range d.value {
		if r <= 127 && r >= 32 && r != '%' {
			// ASCII printable characters except %
			buf.WriteRune(r)
		} else {
			// Percent-encode everything else
			utf8Bytes := []byte(string(r))
			for _, b := range utf8Bytes {
				buf.WriteString(fmt.Sprintf("%%%.2x", b))
			}
		}
	}
	buf.WriteByte('"')
	return buf.Bytes(), nil
}

// Type returns the type of the DisplayStringBareItem, useful when
// you have a list of BareItems and need to know the type
// of each item.
func (d DisplayStringBareItem) Type() int {
	return DisplayStringType
}
