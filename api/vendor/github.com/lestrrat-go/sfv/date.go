package sfv

import (
	"bytes"
	"strconv"
)

// DateItem represents a Unix timestamp date value,
// with optional parameters.
//
// DateItem implements the Item interface.
type DateItem = FullItem[*DateBareItem, int64]

var _ Item = (*DateItem)(nil)

// DateBareItem is a bare item that represents a Unix timestamp date.
// Bare items cannot have parameters. Some constructs
// may require a bare item instead of a full date item
// (e.g. dictionary values).
type DateBareItem struct {
	uvalue[int64]
}

var _ BareItem = (*DateBareItem)(nil)

// Date creates a new Date (DateItem) with the
// given Unix timestamp. This function does NOT validate the timestamp
// to ensure it is a valid date (Validation only happens
// when the item is marshaled/parsed).
//
// If you need a bare date item, use BareDate() instead.
func Date(timestamp int64) *DateItem {
	return BareDate(timestamp).toItem()
}

func (d *DateBareItem) toItem() *DateItem {
	return &DateItem{
		bare:   d,
		params: NewParameters(),
	}
}

// BareDate creates a new DateBareItem with the given Unix timestamp.
// This function does NOT validate the timestamp to ensure it is a
// valid date (Validation only happens when the item is
// marshaled/parsed).
//
// If you need a full date item (with parameters), use Date() instead.
func BareDate(timestamp int64) *DateBareItem {
	var v DateBareItem
	_ = v.SetValue(timestamp)
	return &v
}

// ToItem converts the DateBareItem to a full Item.
func (d *DateBareItem) ToItem() Item {
	return d.toItem()
}

// MarshalSFV implements the Marshaler interface for DateBareItem.
func (d DateBareItem) MarshalSFV() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('@')
	buf.WriteString(strconv.FormatInt(d.value, 10))
	return buf.Bytes(), nil
}

// Type returns the type of the DateBareItem, useful when
// you have a list of BareItems and need to know the type
// of each item.
func (d DateBareItem) Type() int {
	return DateType
}
