package sfv

import (
	"bytes"
	"strconv"
	"strings"
)

// DecimalItem represents a decimal value,
// with optional parameters.
//
// DecimalItem implements the Item interface.
type DecimalItem = FullItem[*DecimalBareItem, float64]

var _ Item = (*DecimalItem)(nil)

// DecimalBareItem is a bare item that represents a decimal value.
// Bare items cannot have parameters. Some constructs
// may require a bare item instead of a full decimal item
// (e.g. dictionary values).
type DecimalBareItem struct {
	uvalue[float64]
}

var _ BareItem = (*DecimalBareItem)(nil)

// Decimal creates a new Decimal (DecimalItem) with the
// given float64 value. This function does NOT validate the value
// to ensure it is a valid decimal (Validation only happens
// when the item is marshaled/parsed).
//
// If you need a bare decimal item, use BareDecimal() instead.
func Decimal(f float64) *DecimalItem {
	return BareDecimal(f).toItem()
}

func (d *DecimalBareItem) toItem() *DecimalItem {
	return &DecimalItem{
		bare:   d,
		params: NewParameters(),
	}
}

// BareDecimal creates a new DecimalBareItem with the given float64 value.
// This function does NOT validate the value to ensure it is a
// valid decimal (Validation only happens when the item is
// marshaled/parsed).
//
// If you need a full decimal item (with parameters), use Decimal() instead.
func BareDecimal(f float64) *DecimalBareItem {
	var v DecimalBareItem
	_ = v.SetValue(f)
	return &v
}

// ToItem converts the DecimalBareItem to a full Item.
func (d *DecimalBareItem) ToItem() Item {
	return d.toItem()
}

// MarshalSFV implements the Marshaler interface for DecimalBareItem.
func (d DecimalBareItem) MarshalSFV() ([]byte, error) {
	var buf bytes.Buffer

	// Format with up to 3 decimal places, removing trailing zeros
	str := strconv.FormatFloat(d.value, 'f', 3, 64)
	str = strings.TrimRight(str, "0")
	if str[len(str)-1] == '.' {
		// If the last character is a dot, we need to add a zero
		// to avoid an invalid format
		str += "0"
	}
	buf.WriteString(str)
	return buf.Bytes(), nil
}

// Type returns the type of the DecimalBareItem, useful when
// you have a list of BareItems and need to know the type
// of each item.
func (d DecimalBareItem) Type() int {
	return DecimalType
}

// IntegerItem represents an integer value,
// with optional parameters.
//
// IntegerItem implements the Item interface.
type IntegerItem = FullItem[*IntegerBareItem, int64]

var _ Item = (*IntegerItem)(nil)

// IntegerBareItem is a bare item that represents an integer value.
// Bare items cannot have parameters. Some constructs
// may require a bare item instead of a full integer item
// (e.g. dictionary values).
type IntegerBareItem struct {
	uvalue[int64]
}

var _ BareItem = (*IntegerBareItem)(nil)

// Integer creates a new Integer (IntegerItem) with the
// given int64 value. This function does NOT validate the value
// to ensure it is a valid integer (Validation only happens
// when the item is marshaled/parsed).
//
// If you need a bare integer item, use BareInteger() instead.
func Integer(i int64) *IntegerItem {
	return BareInteger(i).toItem()
}

func (i *IntegerBareItem) toItem() *IntegerItem {
	return &IntegerItem{
		bare:   i,
		params: NewParameters(),
	}
}

// BareInteger creates a new IntegerBareItem with the given int64 value.
// This function does NOT validate the value to ensure it is a
// valid integer (Validation only happens when the item is
// marshaled/parsed).
//
// If you need a full integer item (with parameters), use Integer() instead.
func BareInteger(i int64) *IntegerBareItem {
	var v IntegerBareItem
	_ = v.SetValue(i)
	return &v
}

// ToItem converts the IntegerBareItem to a full Item.
func (i *IntegerBareItem) ToItem() Item {
	return i.toItem()
}

// MarshalSFV implements the Marshaler interface for IntegerBareItem.
func (i IntegerBareItem) MarshalSFV() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(strconv.FormatInt(i.value, 10))
	return buf.Bytes(), nil
}

// Type returns the type of the IntegerBareItem, useful when
// you have a list of BareItems and need to know the type
// of each item.
func (i IntegerBareItem) Type() int {
	return IntegerType
}
