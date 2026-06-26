package sfv

import "github.com/lestrrat-go/blackmagic"

// BooleanItem represents a boolean value,
// with optional parameters.
//
// BooleanItem implements the Item interface.
type BooleanItem = FullItem[BooleanBareItem, bool]

var _ Item = (*BooleanItem)(nil)

// BooleanBareItem is a bare item that represents a boolean value.
// Bare items cannot have parameters. Some constructs
// may require a bare item instead of a full boolean item
// (e.g. dictionary values).
//
// BooleanBareItem uses immutable static objects - use True() and False()
// to get the singleton instances.
type BooleanBareItem bool

var _ BareItem = True()

// Boolean creates a new Boolean (BooleanItem) with the
// given bool value. This function uses the static True()/False()
// singleton objects internally.
//
// If you need a bare boolean item, use BareBoolean() instead.
func Boolean(b bool) *BooleanItem {
	return BareBoolean(b).toItem()
}

func (b BooleanBareItem) toItem() *BooleanItem {
	return &BooleanItem{
		bare:   b,
		params: NewParameters(),
	}
}

// BareBoolean creates a BooleanBareItem with the given bool value.
// This function returns the appropriate static singleton object
// (True() or False()).
//
// If you need a full boolean item (with parameters), use Boolean() instead.
func BareBoolean(b bool) BooleanBareItem {
	if b {
		return True()
	}
	return False()
}

// True returns the static singleton BooleanBareItem representing true.
func True() BooleanBareItem {
	return BooleanBareItem(true)
}

// False returns the static singleton BooleanBareItem representing false.
func False() BooleanBareItem {
	return BooleanBareItem(false)
}

// ToItem converts the BooleanBareItem to a full Item.
func (b BooleanBareItem) ToItem() Item {
	return b.toItem()
}

// SetValue returns the appropriate static singleton object for the given bool value.
func (b BooleanBareItem) SetValue(value bool) BooleanBareItem {
	if value {
		return True()
	}
	return False()
}

// MarshalSFV implements the Marshaler interface for BooleanBareItem.
var trueBareItemBytes = []byte("?1")
var falseBareItemBytes = []byte("?0")

func (b BooleanBareItem) MarshalSFV() ([]byte, error) {
	if bool(b) {
		return trueBareItemBytes, nil
	}
	return falseBareItemBytes, nil
}

// Type returns the type of the BooleanBareItem, useful when
// you have a list of BareItems and need to know the type
// of each item.
func (b BooleanBareItem) Type() int {
	return BooleanType
}

// GetValue retrieves the bool value from the BooleanBareItem.
func (b BooleanBareItem) GetValue(dst any) error {
	return blackmagic.AssignIfCompatible(dst, bool(b))
}
